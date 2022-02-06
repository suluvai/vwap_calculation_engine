package coinbase_client

const (
	RequestTypeSubscribe string = "subscribe"
)

type MessageType struct {
	Type string `json:"type"`
}

type Channel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

// SubscribeReq is a request to be sent to the Coinbase websocket.
type SubscribeReq struct {
	MessageType
	ProductIds []string  `json:"product_ids"`
	Channels   []Channel `json:"channels"`
}

type ChannelResponse struct {
	Name       string   `json:"name"`
	ProductIDs []string `json:"product_ids"`
}

type SubscribeResponse struct {
	MessageType
	Channels  []ChannelResponse `json:"channels"`
	Message   string            `json:"message,omitempty"`
	Size      string            `json:"size"`
	Price     string            `json:"price"`
	ProductId string            `json:"product_id"`
}

type MatchResponse struct {
	MessageType
	Time         string `json:"time"`
	ProductId    string `json:"product_id"`
	Sequence     int64  `json:"sequence"`
	TradeId      int64  `json:"trade_id"`
	MakerOrderId string `json:"maker_order_id"`
	TakerOrderId string `json:"taker_order_id"`
	Side         string `json:"side"`
	Size         string `json:"size"`
	Price        string `json:"price"`
}
