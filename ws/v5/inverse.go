package v5

import (
	"github.com/sngyai/go-bybit/ws"
)

// V5WebsocketInverseServiceI :
type V5WebsocketInverseServiceI interface {
}

// V5WebsocketInverseService :
type V5WebsocketInverseService struct {
	client *ws.WebSocketClient
}
