package main

import (
	"context"
	"fmt"
	"os"

	"github.com/codeinuit/bsky-go-federate/internal/federation"
	"github.com/codeinuit/bsky-go-federate/internal/federation/mastodon"
	"github.com/joho/godotenv"
)

func main() {
	var tooter federation.Federation

	if err := godotenv.Load(); err != nil {
		fmt.Printf("error occurred :%w", err)
	}

	host := os.Getenv("MASTODON_SERVER_URL")
	cid := os.Getenv("MASTODON_APP_CLIENT_ID")
	csecret := os.Getenv("MASTODON_APP_CLIENT_SECRET")
	at := os.Getenv("MASTODON_APP_ACCESS_TOKEN")

	tooter = mastodon.NewClient(host, cid, csecret, at)
	err := tooter.Post(context.Background(), "ouifi")

	fmt.Printf("%w\n", err)
}
