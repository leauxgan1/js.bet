package main

import (
	"code-root/src/components"
	"code-root/src/game"
	"code-root/src/assets"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	datastar "github.com/starfederation/datastar/sdk/go"
)

const (
	PORT = 8080
)


var frameCount int
var homepage []byte 
var currentGame game.GameState
var siteAssets assets.Assets

func main() {
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("GET /", homepageHandler)
	mux.HandleFunc("GET /game", getGame)

	dir, err := os.ReadFile("../static/homepage.html") 
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
	sse.MergeFragmentTempl(components.Game(currentGame,siteAssets))
}


