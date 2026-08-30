package internal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"strconv"

	"crypto/rand"
	"fmt"
	"js-bet/internal/assets"
	"js-bet/internal/components"
	"js-bet/internal/eventlog"
	"js-bet/internal/game"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// "github.com/a-h/templ"
	"github.com/andybalholm/brotli"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

const (
	PORT = 8080
)

var homepage []byte
var siteAssets assets.Assets
var db DBClient
var staticPath string
var sseHub *Hub

var jwt_secret []byte = make([]byte, 8)

func StartServer() {
	// Get access to the filesystem
	projectRoot, err := os.Getwd()
	log.Print(projectRoot)
	if err != nil {
		log.Panic(err)
	}
	staticPath = filepath.Join(projectRoot, "static")
	fileServer := http.FileServer(http.Dir(staticPath))

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Error loading .env file: %s", err)
	}
	secret_key, secret_found := os.LookupEnv("SECRET_KEY")
	if secret_found {
		jwt_secret = []byte(secret_key)
	} else {
		log.Print(">SECRET_KEY env var not found, falling back to creating random authentication secret key")
		log.Print(">Randomness can be manipulated, stay safe")
		_, err = rand.Read([]byte(jwt_secret))
		if err != nil {
			log.Panicf("Random authentication key failed to initialize, aborting: %s", err)
			return
		}
		log.Printf("%s", jwt_secret)
	}

	// Create new server Handler
	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
	mux.Handle("/game/", authMiddlewarePermissive(http.HandlerFunc(handleGame)))
	mux.HandleFunc("/user/promptLogin", handlePromptLoginRequest)
	mux.HandleFunc("/user/login", handleLoginRequest)
	mux.Handle("/user/placeBet", authMiddlewareStrict(http.HandlerFunc(handlePlaceBet)))

	// Setup event log for server
	eventlog.EventLog = eventlog.New()

	siteAssets = assets.New()
	siteAssets.ReadIcons(filepath.Join(staticPath, "icons"))

	port := fmt.Sprintf(":%d", PORT)
	s := &http.Server{
		Addr:           port,
		Handler:        mux,
		WriteTimeout:   time.Second * 5,
		ReadTimeout:    time.Second * 5,
		IdleTimeout:    time.Second * 5,
		MaxHeaderBytes: 1 << 20,
	}

	db = CreateClient()
	if err = db.InitDB(); err != nil {
		log.Panicf("Error initializing database: %v", err)
	}

	currentGame := game.New()

	sseHub = NewHub()
	go sseHub.Run()

	// Start first game and run until server closes
	go runGame(currentGame, sseHub)

	// go func() {
	log.Printf("Starting server on https://localhost:%d\n", PORT)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
	// }()
}

func runGame(gs game.GameState, hub *Hub) {
	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()

	var buffer bytes.Buffer
	buffer.Grow(300) // Preallocate buffer for average message length
	w := bufio.NewWriter(&buffer)

	for range ticker.C {
		// If health of either combatant reaches 0, start a new game
		buffer.Reset()

		gs.StepGame()
		if gs.Winner != game.NEITHER {
			AwardBets(gs.Winner)
		}

		if len(sseHub.clients) > 0 {
			// Render new gamestate into html for all clients
			sides := components.FighterSides(gs, siteAssets)
			err := sides.Render(context.TODO(), w)
			if err != nil {
				log.Panic(err)
			}

			events := components.EventLog(eventlog.EventLog)
			err = events.Render(context.TODO(), w)
			if err != nil {
				log.Panic(err)
			}
			w.Flush()
			hub.broadcast <- buffer.Bytes()
			// log.Printf("RENDERED")
		}
	}
}

/*
Connects user to SSE connection to get game updates
Attempts to serve the html with different forms of compression depending on the accepted content encodings of the client
*/
func handleGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Incorrect method for endpoint '/game/', expected POST", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)
	rc.SetWriteDeadline(time.Time{})

	encodings := r.Header.Get("Accept-Encoding")
	var brotliWriter *brotli.Writer = nil
	var gzipWriter *gzip.Writer = nil
	switch {
	case strings.Contains(encodings, "br"):
		w.Header().Set("Content-Encoding", "br")
		brotliWriter = brotli.NewWriterOptions(w, brotli.WriterOptions{Quality: 5, LGWin: 24})
	case strings.Contains(encodings, "gzip"):
		w.Header().Set("Content-Encoding", "gzip")
		var err error
		gzipWriter, err = gzip.NewWriterLevel(w, 5)
		if err != nil {
			gzipWriter = nil
			break
		}
	}

	// Create client and register them in the SSE fan-out system
	client := make(chan []byte, 8)
	sseHub.register <- client
	defer func() { sseHub.unregister <- client }()

	// Create flusher to flush the stream and send out bytes
	flusher, ok := w.(http.Flusher)
	if !ok {
		// SHOULD USE A NON-STREAMING METHOD HERE (POLLING)
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	// Attain the authenticated userId if it is present and show the user's view, otherwise show an unauthenticated view
	userId, foundUser := r.Context().Value("userId").(int64)
	log.Printf("Userid found: %d", userId)
	if foundUser {
		// Populate userID in user->client map
		userClientMap[userId] = client
		defer func() { delete(userClientMap, userId) }()

		// Send swap of user gold which is updated when the player bets or wins a bet
		userState := game.UserState{
			UserName: "Default Username",
			Gold:     DefaultGold,
		}
		userUI := renderUserProfile(userState)
		// Send directly via SSE
		if brotliWriter != nil {
			WriteSSE(brotliWriter, userUI)
			brotliWriter.Flush()
		} else if gzipWriter != nil {
			WriteSSE(gzipWriter, userUI)
			gzipWriter.Flush()
		} else {
			WriteSSE(w, userUI)
		}
		flusher.Flush()
	}

	for {
		select {
		case html, ok := <-client:
			if !ok {
				return
			}
			var writeErr error
			if brotliWriter != nil {
				// log.Printf("Compressing with brotli\n")
				writeErr = WriteSSE(brotliWriter, html)
				err := brotliWriter.Flush()
				if err != nil {
					fmt.Printf("error flushing writer %v", err)
				}
			} else if gzipWriter != nil {
				// log.Printf("Compressing with gzip\n")
				writeErr = WriteSSE(gzipWriter, html)
				err := gzipWriter.Flush()
				if err != nil {
					fmt.Printf("error flushing writer %v", err)
				}
			} else {
				// log.Printf("Sending uncompressed html \n")
				writeErr = WriteSSE(w, html)
			}
			if writeErr != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}

	}

}

/*
Attempt to create a JWT to authorize a new or existing user so they can use the authenticated features and endpoints
*/
func handleLoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Panic("Incorrect method for endpoint 'user/login', expected POST")
		return
	}
	r.ParseForm()
	userName := r.FormValue("name")
	passWord := r.FormValue("pass")
	w.Header().Set("Content-Type", "text/html")

	userId, err := db.CheckAddUser(userName, passWord)
	if err != nil {
		_, err := fmt.Fprint(w, "<div> Unable to add or find user in the database</div>")
		if err != nil {
			log.Panic(err)
		}
		return
	}

	claims := UserClaims{
		UserID:   userId,
		Password: passWord,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "js.bet",
			Subject:   fmt.Sprintf("%d", userId),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwt_secret)
	if err != nil {
		log.Printf("Error signing string: %s", err)
	}
	log.Printf("Adding cookie with token %v", signed)
	cookie := http.Cookie{
		Name:     "jwt_token",
		Value:    signed,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   86400,
	}
	http.SetCookie(w, &cookie)
	log.Printf("Added cookie %v", cookie)

	http.Redirect(w, r, "/game/", http.StatusFound)

	_, err = fmt.Fprintf(w, "<div>Received user with name: %s, added to db!</div>", userName)
	if err != nil {
		log.Panic(err)
		return
	}
}

/*
On a GET request, send a signup popup gui to the user
*/
func handlePromptLoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Panic("Incorrect method for endpoint 'user/promptLogin', expected GET")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	signup := components.PopupLogin()
	err := signup.Render(context.Background(), w)
	if err != nil {
		log.Panic(err)
		return
	}
}

/*
Get authentication token from user and place a bet with their ID
*/
func handlePlaceBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	userId, foundUser := r.Context().Value("userId").(int64)
	log.Printf("Userid found: %d", userId)
	if !foundUser {
		http.Error(w, "Unable to find user to place bet for", http.StatusForbidden)
		return
	}

	r.ParseForm()
	if r.Form == nil {
		log.Panic("Unable to parse form")
		http.Error(w, "Unable to find user to place bet for", http.StatusBadRequest)
		return
	}
	betSide := r.FormValue("betside")

	var isLeft bool
	switch betSide {
	case "left":
		isLeft = true
	case "right":
		isLeft = false
	default:
		log.Print("Bet side not found")
		return
	}

	betAmount, err := strconv.Atoi(r.FormValue("betamount"))
	if err != nil {
		log.Print("Unable to determine bet amount from form values")
		http.Error(w, "Unable to determine bet amount from form values", http.StatusBadRequest)
		return
	}
	SetBet(userId, betAmount, isLeft)
	fmt.Fprintf(w, "Placed bet amount for $%d", betAmount)
}
