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

func main() {
	exePath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	basePath = filepath.Dir(exePath)
	staticPath := filepath.Join(basePath,"static")
	fmt.Println(basePath)
	fmt.Println(staticPath)

	// Setup event log for server
	eventlog.EventLog = eventlog.New()

	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("GET /", homepageHandler)
	mux.HandleFunc("GET /game", getGame)
	mux.HandleFunc("GET /user/{name}", handleGetUserInfo)

	dir, err := os.ReadFile(filepath.Join(staticPath,"homepage.html")) 
	if err != nil {
		log.Panic(err)
	}
	homepage = dir

	siteAssets = assets.New()
	siteAssets.ReadIcons("../static/icons")

	port := fmt.Sprintf(":%d",PORT)
	s := &http.Server {
		Addr:           port,
		Handler:       	mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	db = CreateClient()

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
	w.Write(homepage)
}

func getGame(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w,r)
	// for {
	sse.MergeFragmentTempl(components.Game(currentGame,siteAssets,eventlog.EventLog))
	// 	time.Sleep(time.Millisecond * 500)
	// }
	
}

func handleLoginRequest(w http.ResponseWriter, r *http.Request) {
	
}

func handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get("name")
	w.Header().Set("Content-Type","text/html")
	// gold := db.GetUserGold(name)
	w.Write([]byte(fmt.Sprintf("<div>Got name: %s</div>",name)))
}

func handlePlaceBet(w http.ResponseWriter, r *http.Request) {
	
}



