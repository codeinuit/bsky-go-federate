package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/codeinuit/bsky-go-federate/internal/federation/mastodon"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Message struct {
	DID    string  `json:"did"`
	Commit *Commit `json:"commit,omitempty"`
}

type Commit struct {
	Record json.RawMessage `json:"record,omitempty`
}

func bskyFiresky(ctx context.Context) error {
	d := websocket.DefaultDialer
	datachansempai := make(chan Message)

	userdid := os.Getenv("BSKY_USER_DID")

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT)
	defer stop()
	con, _, err := d.DialContext(ctx, fmt.Sprintf("wss://jetstream2.us-east.bsky.network/subscribe?wantedCollections=app.bsky.feed.post&wantedDids=%s", userdid), nil)
	if err != nil {
		return fmt.Errorf("dial failure: %w", err)
	}

	// running reader loop
	go func() {
		for {
			msg := Message{}
			err := con.ReadJSON(&msg)
			if err != nil {
				slog.Error("connexion closed")
				close(datachansempai)
				return
			}

			datachansempai <- msg
		}
	}()

	for {
		message := Message{}

		select {
		case <-ctx.Done():
			slog.Info("context canceled")
			con.Close()
			return nil
		case message = <-datachansempai:

		}

		if err != nil {
			slog.Error(err.Error())
			continue
		}

		var post bsky.FeedPost
		if err = json.Unmarshal(message.Commit.Record, &post); err != nil {
			slog.Error(err.Error())
			continue
		}

		slog.Info(post.Text)
	}
}

func main() {
	//var tooter federation.Federation
	log := slog.Default()

	if err := godotenv.Load(); err != nil {
		log.Error("error occurred: ", err.Error())
	}

	host := os.Getenv("MASTODON_SERVER_URL")
	cid := os.Getenv("MASTODON_APP_CLIENT_ID")
	csecret := os.Getenv("MASTODON_APP_CLIENT_SECRET")
	at := os.Getenv("MASTODON_APP_ACCESS_TOKEN")

	_ = mastodon.NewClient(host, cid, csecret, at)

	bskyFiresky(context.Background())
}
