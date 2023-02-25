package v5

import (
	"github.com/sngyai/go-bybit/ws"
)

// V5WebsocketLinearServiceI :
type V5WebsocketLinearServiceI interface {
}

// V5WebsocketLinearService :
type V5WebsocketLinearService struct {
	Client *ws.WebSocketClient
}
