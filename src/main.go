package main

import (
	"code-root/src/assets"
	"code-root/src/components"
	"code-root/src/game"
	"code-root/src/eventlog"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"


	datastar "github.com/starfederation/datastar/sdk/go"
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
	mux.HandleFunc("/", homepageHandler)

	// Handlers
	// mux.HandleFunc("/", homepageHandler)
	mux.HandleFunc("/game/", getGame)
	mux.HandleFunc("/user/new/", handleLoginRequest)
	mux.HandleFunc("/user/gold/", handleGetUserInfo)
	mux.HandleFunc("/placeBet/", handlePlaceBet)

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
	runServer := func() {
		if err :=	s.ListenAndServe(); err != nil {
			log.Panic(err)
		}
	}
	go runServer()

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
		gameRender := components.Game(currentGame,siteAssets,eventlog.EventLog)
		err := sse.MergeFragmentTempl(gameRender)
		if err != nil {
			panic(err)
		}

		err = rc.SetWriteDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Millisecond * 500)
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

func handlePlaceBet(w http.ResponseWriter, r *http.Request) {
	
}



