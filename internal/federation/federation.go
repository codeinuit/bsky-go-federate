package federation

import "context"

// Federation interface calls to Federated servers
type Federation interface {
	Post(context.Context, string) error
}
