package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/suluvai/vwap_calculation_engine/conf"
	"github.com/suluvai/vwap_calculation_engine/vwap"
	"github.com/suluvai/vwap_calculation_engine/websocket/coinbase_client"
)

func Run(ctx context.Context, wsclient *coinbase_client.WebSocketClient, list *vwap.DataSet) error {

	messageReceiver := make(chan interface{})
	go wsclient.Listen(ctx, messageReceiver)

	for data := range messageReceiver {
		switch matchResponse := data.(type) {
		case *coinbase_client.MatchResponse:
			if matchResponse != nil && matchResponse.Price != "" {
				price, err := decimal.NewFromString(matchResponse.Price)
				if err != nil {
					return errors.New("Error in converting price to decimal. " + matchResponse.Price + ", Error: " + err.Error())
				}

				volume, err := decimal.NewFromString(matchResponse.Size)
				if err != nil {
					return errors.New("Error in converting size to decimal. " + matchResponse.Size + ", Error: " + err.Error())
				}

				list.Update(vwap.DataPoint{
					Price:     price,
					Volume:    volume,
					ProductId: matchResponse.ProductId,
				})

				results, _ := json.Marshal(list.VWAP)
				log.Printf(string(results))
			}
		}
	}

	return nil
}

// Application's main program
func main() {
	ctx := context.Background()

	log.Info("====================== Volume Weighted Average Price Calculation Engine ======================")

	// Set log level.
	log.Info("Log level is " + conf.Configuration.LogLevel)
	switch conf.Configuration.LogLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Warn("LogLevel not found, defaulting to INFO")
		log.SetLevel(log.InfoLevel)
	}

	// Intercepting shutdown signals.
	go func() {
		// `signal.Notify` registers the given channel to
		// receive notifications of the specified signals.
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

		s := <-interrupt

		log.Info("Received signal: ", s)
		os.Exit(0)
	}()

	list, err := vwap.NewDataSet([]vwap.DataPoint{}, conf.Configuration.VwapWindowSize)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to websocket
	wsclient := coinbase_client.NewWebSocketClient(ctx, conf.Configuration.WebServerURL)
	wsclient.Connect(wsclient.Ctx)
	err = wsclient.Subscribe(wsclient.Ctx, conf.Configuration.TradingPairs)
	if err != nil {
		log.Fatal("service subscription err: ", err)
	}

	err = Run(wsclient.Ctx, wsclient, &list)
	if err != nil {
		log.Fatal(err)
	}

	wsclient.Disconnect()
	log.Info("Exiting Application")
}
