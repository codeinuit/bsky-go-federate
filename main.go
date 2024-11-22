package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/codeinuit/bsky-go-federate/internal/federation"
	"github.com/codeinuit/bsky-go-federate/internal/federation/mastodon"
	"github.com/joho/godotenv"
)

func main() {
	var tooter federation.Federation
	log := slog.Default()

	if err := godotenv.Load(); err != nil {
		log.Error("error occurred: ", err.Error())
	}

	host := os.Getenv("MASTODON_SERVER_URL")
	cid := os.Getenv("MASTODON_APP_CLIENT_ID")
	csecret := os.Getenv("MASTODON_APP_CLIENT_SECRET")
	at := os.Getenv("MASTODON_APP_ACCESS_TOKEN")

	tooter = mastodon.NewClient(host, cid, csecret, at)
	if err := tooter.Post(context.Background(), "ouifi"); err != nil {
		log.Error("could not post to mastodon: ", err.Error())
	}
}
