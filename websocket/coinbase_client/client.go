package coinbase_client

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Send pings to peer with this period
const pingPeriod = 30 * time.Second

// WebSocketClient returns websocket connection
type WebSocketClient struct {
	wsconn    *websocket.Conn
	Url       string
	Ctx       context.Context
	CtxCancel context.CancelFunc
	mu        sync.RWMutex
}

// NewWebSocketClient creates new websocket connection
func NewWebSocketClient(ctx context.Context, url string) *WebSocketClient {

	wsclient := WebSocketClient{}
	wsclient.Ctx, wsclient.CtxCancel = context.WithCancel(ctx)
	wsclient.Url = url

	return &wsclient
}

// Establishes new web socket connection.
func (wsclient *WebSocketClient) Connect(ctx context.Context) error {

	if wsclient.wsconn != nil {
		return nil
	}

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return errors.New("timed out")
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			ws, _, err := websocket.DefaultDialer.Dial(wsclient.Url, nil)
			if err != nil {
				log.Error("Cannot connect to websocket: ", wsclient.Url)
				continue
			}
			log.Info("connected to websocket: ", wsclient.Url)
			wsclient.wsconn = ws
			return nil
		}
	}
}

// Disconnect will send close message and shutdown websocket connection
func (wsclient *WebSocketClient) Disconnect() {
	wsclient.CtxCancel()
	wsclient.mu.Lock()
	log.Info("Closing websocket connection")
	if wsclient.wsconn != nil {
		wsclient.wsconn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		wsclient.wsconn.Close()
		wsclient.wsconn = nil
	}
	wsclient.mu.Unlock()
}

// Subscribe subscribes to the websocket.
func (wsclient *WebSocketClient) Subscribe(ctx context.Context, tradingPairs []string) error {
	if len(tradingPairs) == 0 {
		return errors.New("tradingPairs must be provided")
	}

	log.Infof("Subscribing for Trading Pairs: %v", tradingPairs)

	subscription := SubscribeReq{
		MessageType: MessageType{
			Type: RequestTypeSubscribe,
		},
		Channels: []Channel{
			{
				Name:       "matches",
				ProductIds: tradingPairs,
			},
		},
	}

	wsclient.mu.Lock()
	defer wsclient.mu.Unlock()

	err := wsclient.wsconn.WriteJSON(subscription)
	if err != nil {
		return errors.New("failed to send subscription: " + err.Error())
	}

	var response SubscribeResponse
	err = wsclient.wsconn.ReadJSON(&response)
	if err != nil {
		return errors.New("failed to receive subscription response: " + err.Error())
	}

	if response.Type == "error" {
		return errors.New("failed to subscribe: " + response.Message)
	}

	subscriptionResponse, err := json.Marshal(response)
	if err != nil {
		return errors.New("failed to marshal subscription: " + err.Error())
	}
	log.Info("subscriptionResponse: ", string(subscriptionResponse))
	return nil
}

// Parses the message received.
func (wsclient *WebSocketClient) parsePayload(msg []byte) (*MatchResponse, error) {
	messageType := MessageType{}
	err := json.Unmarshal(msg, &messageType)
	if err != nil {
		return nil, err
	}

	if messageType.Type == "match" || messageType.Type == "last_match" {
		matchResponse := MatchResponse{}
		err = json.Unmarshal(msg, &matchResponse)
		if err != nil {
			return nil, err
		}
		return &matchResponse, nil
	}
	return nil, errors.New("parsePayload, received unsupported message type: " + messageType.Type)
}

// Listens for messages, typically the messages this client is subscribed for.
func (wsclient *WebSocketClient) Listen(ctx context.Context, receiver chan interface{}) {
	ticker := time.NewTicker(time.Nanosecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wsclient.Disconnect()
		case <-ticker.C:

			if wsclient.wsconn == nil {
				log.Fatalf("Websocket connection is nil")
				return
			}

			wsclient.mu.Lock()
			messageType, message, err := wsclient.wsconn.ReadMessage()
			wsclient.mu.Unlock()
			if err != nil {
				log.Errorf("Listen, error: %v\n", err)
				wsclient.Disconnect()
			}

			switch messageType {
			case websocket.PingMessage:
				log.Debugf("MessageType: websocket.PingMessage, Message: %s", string(message))
				wsclient.mu.Lock()
				err := wsclient.wsconn.WriteMessage(websocket.PongMessage, []byte(time.Now().String()))
				wsclient.mu.Unlock()
				if err != nil {
					wsclient.Disconnect()
				}
			case websocket.TextMessage:
				log.Debugf("MessageType: websocket.TextMessage, Message: %s", string(message))
				matchResponse, err := wsclient.parsePayload(message)
				if err != nil {
					wsclient.Disconnect()
				}
				receiver <- matchResponse
			case websocket.BinaryMessage:
				log.Debugf("MessageType: websocket.TextMessage, Ignoring Binary Messages")
			case websocket.PongMessage:
				wsclient.mu.Lock()
				err := wsclient.wsconn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pingPeriod/2))
				wsclient.mu.Unlock()
				if err != nil {
					wsclient.Disconnect()
				}
			case websocket.CloseMessage:
				log.Debugf("MessageType: websocket.CloseMessage, Message: %s", string(message))
				log.Info("Received CloseMessage from Server.Proceeding to close connection...")
				wsclient.Disconnect()
			}
		}
	}
}
