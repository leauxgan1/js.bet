package internal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"strconv"

	// "encoding/json"
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

	"github.com/andybalholm/brotli"
)

const (
	PORT = 8080
)

var frameCount int
var homepage []byte
var currentGame game.GameState
var siteAssets assets.Assets
var db DBClient
var basePath string
var staticPath string

var sseHub *Hub

func StartServer() {
	// Get access to the filesystem
	projectRoot, err := os.Getwd()
	log.Print(projectRoot)
	if err != nil {
		log.Panic(err)
	}
	staticPath = filepath.Join(projectRoot, "static")
	fileServer := http.FileServer(http.Dir(staticPath))

	// Create new server Handler
	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
	mux.HandleFunc("/game/", handleGame)
	mux.HandleFunc("/user/signup", handleSignupRequest)
	mux.HandleFunc("/user/new", handleNewUserRequest)
	mux.HandleFunc("/user/login", handleLoginRequest)
	mux.HandleFunc("/user/gold", handleGetUserInfo)
	mux.HandleFunc("/placeBet", handlePlaceBet)

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
		MaxHeaderBytes: 1 << 20,
	}

	db = CreateClient()
	if err = db.InitDB(); err != nil {
		log.Panicf("Error initializing database: %v", err)
	}

	log.Printf("Starting server on https://localhost:%d\n", PORT)

	currentGame = game.New()

	sseHub = NewHub()
	go sseHub.Run()
	// Start first game and run until server closes
	go runGame(currentGame, sseHub)

	if err := s.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func runGame(gs game.GameState, hub *Hub) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var buffer bytes.Buffer
	buffer.Grow(300)
	w := bufio.NewWriter(&buffer)

	for range ticker.C {
		// If health of either combatant reaches 0, start a new game
		buffer.Reset()

		gs.StepGame()

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
		log.Panic("Incorrect method for endpoint '/game/', expected POST")
		return
	}
	rc := http.NewResponseController(w)
	rc.SetWriteDeadline(time.Time{})

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Play-Audio", "attack,block")

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

	client := make(chan []byte, 8)

	sseHub.register <- client
	defer func() { sseHub.unregister <- client }()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case html, ok := <-client:
			if !ok {
				return
			}
			var writeErr error
			if brotliWriter != nil {
				log.Printf("Compressing with brotli\n")
				writeErr = WriteSSE(brotliWriter, html)
				err := brotliWriter.Flush()
				if err != nil {
					fmt.Printf("error flushing writer %v", err)
				}
			} else if gzipWriter != nil {
				log.Printf("Compressing with gzip\n")
				writeErr = WriteSSE(gzipWriter, html)
				err := gzipWriter.Flush()
				if err != nil {
					fmt.Printf("error flushing writer %v", err)
				}
			} else {
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

func handleLoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Panic("Incorrect method for endpoint 'user/signup', expected POST")
		return
	}
	params := r.URL.Query()
	userName := params.Get("name")
	w.Header().Set("Content-Type", "text/html")
	_, err := db.CheckAddUser(userName)
	if err != nil {
		log.Panic(err)
		return
	}
	_, err = fmt.Fprintf(w, "<div>Received user with name: %s, added to db!</div>", userName)
	if err != nil {
		log.Panic(err)
		return
	}
}

func handleNewUserRequest(w http.ResponseWriter, r *http.Request) {
	// On a POST request, accept a username an password as params, santitize them, and if unique, add them as a new user to the database
	if r.Method != http.MethodPost {
		log.Panic("Incorrect method for endpoint 'user/signup', expected POST")
		return
	}
	params := r.URL.Query()
	userName := params.Get("name")
	w.Header().Set("Content-Type", "text/html")
	_, err := db.CheckAddUser(userName)
	if err != nil {
		log.Panic(err)
		return
	}
	_, err = fmt.Fprintf(w, "<div>Received user with name: %s, added to db!</div>", userName)
	if err != nil {
		log.Panic(err)
		return
	}
}

func handleSignupRequest(w http.ResponseWriter, r *http.Request) {
	// On a GET request, send a signup popup gui to the user
	if r.Method != http.MethodGet {
		log.Panic("Incorrect method for endpoint 'user/signup', expected GET")
		return
	}

	signup := components.PopupSignup()
	err := signup.Render(context.Background(), w)

	if err != nil {
		log.Panic(err)
		return
	}
}

func handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get("name")
	w.Header().Set("Content-Type", "text/html")
	gold, err := db.GetUserGold(name)

	if err != nil {
		_, err = fmt.Fprintf(w, "<div>Got name: %s</div> <div> Gold data unavailable... </div>", name)
		if err != nil {
			log.Panic(err)
			return
		}
	} else {
		_, err = fmt.Fprintf(w, "<div>Got name: %s</div> <div> Has %d gold... </div>", name, gold)
		if err != nil {
			log.Panic(err)
			return
		}
	}
}

type betShape struct {
	BetSide   string `json:"betside"`
	BetAmount int    `json:"betamount"`
}

func handlePlaceBet(w http.ResponseWriter, r *http.Request) {
	// Show a popup temporarily to confirm the user has bet some amount
	if r.Method != http.MethodPost {
		return
	}
	r.ParseForm()
	if r.Form == nil {
		log.Panic("Unable to parse form")
	}
	betAmount, err := strconv.Atoi(r.FormValue("betamount"))
	if err != nil {
		log.Panic("Unable to determine bet amount from form values")
	}
	game.SetBet("TempUsername", betAmount)

	fmt.Fprintf(w, "Place bet amount for $%d", betAmount)

}
