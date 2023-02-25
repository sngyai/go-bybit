package v5

import (
	"github.com/sngyai/go-bybit/ws"
)

// V5WebsocketOptionServiceI :
type V5WebsocketOptionServiceI interface {
}

// V5WebsocketOptionService :
type V5WebsocketOptionService struct {
	client *ws.WebSocketClient
}
