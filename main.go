package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/codeinuit/bsky-go-federate/internal/federation"
	"github.com/codeinuit/bsky-go-federate/internal/federation/mastodon"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go/log"
)

func newBSkyFirehoseListener(ctx context.Context, callback func(context.Context, string) error) error {
	d := websocket.DefaultDialer

	con, _, err := d.Dial("wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos", nil)
	//con.SetReadDeadline(time.Time{})
	if err != nil {
		return fmt.Errorf("dial failure: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT)
	defer stop()

	rscb := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if evt.Repo != "did:plc:3yv2u4562lwfy4xibtqruj6i" {
				return nil
			}

			rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
			if err != nil {
				fmt.Println(err)
			}
			slog.Info("message received")

			for _, op := range evt.Ops {
				ek := repomgr.EventKind(op.Action)
				switch ek {
				case repomgr.EvtKindCreateRecord:
					rc, rec, err := rr.GetRecord(ctx, op.Path)
					if err != nil {
						e := fmt.Errorf("getting record %s (%s) within seq %d for %s: %w", op.Path, *op.Cid, evt.Seq, evt.Repo, err)
						log.Error(e)
						continue
					}

					if lexutil.LexLink(rc) != *op.Cid {
						//return fmt.Errorf("mismatch in record and op cid: %s != %s", rc, *op.Cid)
						continue
					}

					banana := lexutil.LexiconTypeDecoder{
						Val: rec,
					}

					var pst = bsky.FeedPost{}
					b, err := banana.MarshalJSON()
					if err != nil {
						fmt.Println(err)
					}

					err = json.Unmarshal(b, &pst)
					if err != nil {
						fmt.Println(err)
					}

					if pst.LexiconTypeID == "app.bsky.feed.post" && pst.Reply == nil {
						slog.Info("posting message")
						err = callback(ctx, pst.Text)
						if err != nil {
							panic(err)

						}
					}
				}

			}

			return nil
		},
		Error: func(errf *events.ErrorFrame) error {
			err = fmt.Errorf("error frame: %s: %s", errf.Error, errf.Message)
			slog.Error(err.Error())
			return err
		},
	}

	seqScheduler := sequential.NewScheduler(con.RemoteAddr().String(), rscb.EventHandler)
	err = events.HandleRepoStream(ctx, con, seqScheduler)

	if err != nil {
		slog.Error("error occurred while exiting: ", err.Error())
	}

	return err
}

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

	newBSkyFirehoseListener(context.Background(), tooter.Post)
}
