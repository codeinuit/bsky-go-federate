package mastodon

import (
	"context"

	"github.com/mattn/go-mastodon"
)

// Mastodon represents an instance for a client
type Mastodon struct {
	cli *mastodon.Client
}

func NewClient(s, cid, csec, at string) Mastodon {
	c := mastodon.Config{
		Server:       s,
		ClientID:     cid,
		ClientSecret: csec,
		AccessToken:  at,
	}

	return Mastodon{
		cli: mastodon.NewClient(&c),
	}
}

func (m Mastodon) Post(ctx context.Context, msg string) error {
	_, err := m.cli.PostStatus(ctx, &mastodon.Toot{
		Status:     msg,
		Visibility: "public",
	})

	return err
}
