package command

import (
	"context"
	"time"
)

type Command interface {
	Run(context.Context) error
}

// rpcTimeout allows requests to complete independently from the server's ctx,
// supporting graceful shutdown rather than immediate cancellation.
var rpcTimeout = 3 * time.Second
