package coinbase_client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suluvai/vwap_calculation_engine/websocket/coinbase_client"
)

const DefaultURL = "wss://ws-feed.exchange.coinbase.com"

func TestNewWebSocketClient(t *testing.T) {
	ctx := context.Background()

	wsclient := coinbase_client.NewWebSocketClient(ctx, DefaultURL)
	require.NotNil(t, wsclient)
}

func TestWebsocketSubscribe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	listOfTestScenarios := []struct {
		Name            string
		TradingPairs    []string
		ExpectedToError bool
	}{
		{
			Name:            "TradingPairs",
			TradingPairs:    []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
			ExpectedToError: false,
		},
		{
			Name:            "InvalidPairs",
			TradingPairs:    []string{"xxx-USD"},
			ExpectedToError: true,
		},
	}

	receiver := make(chan interface{})

	for _, eachTest := range listOfTestScenarios {

		t.Run(eachTest.Name, func(t *testing.T) {
			t.Parallel()

			wsclient := coinbase_client.NewWebSocketClient(ctx, DefaultURL)
			require.NotNil(t, wsclient)

			err := wsclient.Connect(wsclient.Ctx)
			require.NoError(t, err)

			err = wsclient.Subscribe(wsclient.Ctx, eachTest.TradingPairs)
			if eachTest.ExpectedToError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				go wsclient.Listen(wsclient.Ctx, receiver)

				var counter int
				minNumberOfMsgsExpected := 3

				// Check the first two messages.
				for msg := range receiver {
					switch msg := msg.(type) {
					case *coinbase_client.MatchResponse:
						if counter >= minNumberOfMsgsExpected {
							break
						}

						if msg.Type == "last_match" {
							require.Contains(t, eachTest.TradingPairs, msg.ProductId)
						}
						counter++
					}
				}
			}
		})
	}
}
