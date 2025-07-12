package main

import (
	"code-root/src/assets"
	"code-root/src/components"
	"code-root/src/eventlog"
	"code-root/src/game"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	
	datastar "github.com/starfederation/datastar/sdk/go/datastar"
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
var fileServer http.Handler

func main() {
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	projectRoot = filepath.Dir(projectRoot)
	staticPath = filepath.Join(projectRoot,"static")

	fileServer = http.FileServer(http.Dir(staticPath))

	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("/", homepageHandler)
	mux.HandleFunc("/game/", getGame)
	// mux.HandleFunc("/user/signup", handleSignupRequest)
	// mux.HandleFunc("/user/new", handleNewUserRequest)
	// mux.HandleFunc("/user/login", handleLoginRequest)
	// mux.HandleFunc("/user/gold", handleGetUserInfo)
	// mux.HandleFunc("/placeBet", handlePlaceBet)

	dir, err := os.ReadFile(filepath.Join(staticPath,"homepage.html")) 
	if err != nil {
		log.Panic(err)
	}
	homepage = dir

	// Setup event log for server
	eventlog.EventLog = eventlog.New()

	siteAssets = assets.New()
	siteAssets.ReadIcons(filepath.Join(staticPath,"icons"))

	port := fmt.Sprintf(":%d",PORT)
	s := &http.Server {
		Addr:           port,
		Handler:       	mux,
		WriteTimeout: time.Second * 5,
		ReadTimeout: time.Second * 5,
		MaxHeaderBytes: 1 << 20,
	}
	db = CreateClient()
	if err = db.InitDB(); err != nil {
		log.Panicf("Error initializing database: %v",err)
	}

	log.Printf("Starting server on https://localhost:%d\n",PORT)

	// Run server in new goroutine
	go func() {
		if err :=	s.ListenAndServe(); err != nil {
			log.Panic(err)
		}
	}()

	// Start first game and run until server closes
	currentGame = game.New()
	for {
		// If health of either combatant reaches 0, start a new game
		currentGame.StepGame()

		time.Sleep(time.Second * 1)
		log.Printf("GameState is %v\n",currentGame)
	}

}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	fileServer.ServeHTTP(w,r)
}

func getGame(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	sse := datastar.NewSSE(w,r)
	for {
		log.Print("/game/ reached!")
		sides := components.FighterSides(currentGame,siteAssets)
		err := sse.PatchElementTempl(sides)
		if err != nil {
			log.Panic(err)
		}
		events := components.EventLog(eventlog.EventLog)
		err = sse.PatchElementTempl(events)
		if err != nil {
			log.Panic(err)
		}
		commands := currentGame.AudioPlayers.FormatAudioPlayer() // Get current audio command based on game state
		err = sse.ExecuteScript(commands)
		if err != nil {
			http.Error(w,"Unable to play audio via execute script", 500)
			continue
		}
		err = rc.SetWriteDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			log.Panic(err)
		}
		time.Sleep(time.Second)
	}
}

func handleLoginRequest(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	userName := params.Get("name")
	w.Header().Set("Content-Type","text/html")
	_, err := db.CheckAddUser(userName)
	if err != nil {
		log.Panic(err)
		return 
	}
	w.Write([]byte("<div>Received user with name: "));
	w.Write([]byte(userName));
	w.Write([]byte(", added to db!</div>"));
}

func handleNewUserRequest(w http.ResponseWriter, r *http.Request) {
	// On a POST request, accept a username an password as params, santitize them, and if unique, add them as a new user to the database
	if r.Method != http.MethodPost {
		log.Panic("Incorrect method for endpoint 'user/signup', expected GET");
		return;
	}
	params := r.URL.Query()
	userName := params.Get("name")
	w.Header().Set("Content-Type","text/html")
	_, err := db.CheckAddUser(userName)
	if err != nil {
		log.Panic(err)
		return 
	}
	w.Write([]byte("<div>Received user with name: "));
	w.Write([]byte(userName));
	w.Write([]byte(", added to db!</div>"));
}

func handleSignupRequest(w http.ResponseWriter, r *http.Request) {
	// On a GET request, send a signup popup gui to the user
	if r.Method != http.MethodGet {
		log.Panic("Incorrect method for endpoint 'user/signup', expected GET")
		return
	}
	sse := datastar.NewSSE(w,r)
	err := sse.PatchElementTempl(components.PopupSignup())
	if err != nil {
		log.Panic(err)
		return 
	}
}

func handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get("name")
	w.Header().Set("Content-Type","text/html")
	gold, err := db.GetUserGold(name)
	if err != nil {
		w.Write(fmt.Appendf([]byte{},"<div>Got name: %s</div> <div> Gold data unavailable... </div>",name))
		return
	}
	w.Write(fmt.Appendf([]byte{},"<div>Got name: %s</div> <div> Has %d gold... </div>",name,gold))
}

type BetShape struct {
	BetAmount int `json:"betamount"`
	BetSide string `json:"betside"`
}

func handlePlaceBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	log.Printf("Body received: %s", r.Body)
	decoder := json.NewDecoder(r.Body)
	var betInfo BetShape
	err := decoder.Decode(&betInfo)
	if err != nil {
		log.Panicf("Unable to decode json request: %v",err)
	}
	log.Printf("Received amount: %d and side %s",betInfo.BetAmount,betInfo.BetSide)
	if betInfo.BetSide == "Left" {
		log.Printf("Bet on left!")
	} else if betInfo.BetSide == "Right" {
		log.Printf("Bet on right!")
	}
}
