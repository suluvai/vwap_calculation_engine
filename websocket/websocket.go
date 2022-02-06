package websocket

import "context"

type Client interface {
	Connect(ctx context.Context) error
	Subscribe(ctx context.Context, tradingPairs []string) error
	Disconnect()
	Listen(ctx context.Context, receiver chan interface{})
}
