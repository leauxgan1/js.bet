package internal

import (
	"bufio"
	"bytes"
	"context"
	"js-bet/internal/components"
	"js-bet/internal/game"
	"log"
)

func renderUserProfile(userState game.UserState) []byte {
	var buffer bytes.Buffer
	buffer.Grow(300)
	bufWriter := bufio.NewWriter(&buffer)
	fragment := components.Profile(userState)
	err := fragment.Render(context.Background(), bufWriter)
	if err != nil {
		log.Printf("Error rendering user UI: %s", err)
	}
	bufWriter.Flush()
	return buffer.Bytes()
}
