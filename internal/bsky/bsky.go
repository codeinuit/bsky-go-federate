package bsky

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/gorilla/websocket"
)

type Message struct {
	DID    string  `json:"did"`
	Commit *Commit `json:"commit,omitempty"`
}

type Commit struct {
	Record json.RawMessage `json:"record,omitempty`
}

type CallbackOnMessageFunc func(context.Context, string) error

// BSkyService is the BSky listener instance
type BSkyService struct {
	Account  string
	Callback CallbackOnMessageFunc
}

func NewBSkyService() *BSkyService {
	return &BSkyService{}
}

func (bs *BSkyService) WithListenedAccountDID(userdid string) *BSkyService {
	bs.Account = userdid

	return bs
}

func (bs *BSkyService) WithCallbackOnMessage(callback CallbackOnMessageFunc) *BSkyService {
	bs.Callback = callback

	return bs
}

func (bs *BSkyService) Listen(ctx context.Context) error {
	d := websocket.DefaultDialer
	msgchan := make(chan Message)

	con, _, err := d.DialContext(ctx, fmt.Sprintf("wss://jetstream2.us-east.bsky.network/subscribe?wantedCollections=app.bsky.feed.post&wantedDids=%s", bs.Account), nil)
	if err != nil {
		return fmt.Errorf("dial failure: %w", err)
	}

	go func() {
		for {
			msg := Message{}
			err := con.ReadJSON(&msg)
			if err != nil {
				slog.Error("connexion closed")
				close(msgchan)
				return
			}

			msgchan <- msg
		}
	}()

	for {
		message := Message{}

		select {
		case <-ctx.Done():
			slog.Info("context canceled")
			con.Close()
			return nil
		case message = <-msgchan:
		}

		var post bsky.FeedPost
		if err = json.Unmarshal(message.Commit.Record, &post); err != nil {
			slog.Error(err.Error())
			continue
		}

		if post.Reply != nil {
			slog.Debug("reply received; skipping...")
			continue
		}

		bs.Callback(ctx, post.Text)
	}
}
