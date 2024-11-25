package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	intrnbsky "github.com/codeinuit/bsky-go-federate/internal/bsky"
	"github.com/codeinuit/bsky-go-federate/internal/federation"
	"github.com/codeinuit/bsky-go-federate/internal/federation/mastodon"
	"github.com/joho/godotenv"
)

func main() {
	var tooter federation.Federation

	if err := godotenv.Load(); err != nil {
		slog.Error("could not loa: " + err.Error())

		os.Exit(1)
	}

	host := os.Getenv("MASTODON_SERVER_URL")
	cid := os.Getenv("MASTODON_APP_CLIENT_ID")
	csecret := os.Getenv("MASTODON_APP_CLIENT_SECRET")
	at := os.Getenv("MASTODON_APP_ACCESS_TOKEN")
	userdid := os.Getenv("BSKY_USER_DID")

	tooter = mastodon.NewClient(host, cid, csecret, at)

	lstnr := intrnbsky.NewBSkyService().
		WithListenedAccountDID(userdid).
		WithCallbackOnMessage(tooter.Post)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	lstnr.Listen(ctx)
}
