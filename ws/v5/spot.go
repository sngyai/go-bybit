package v5

import (
	"github.com/sngyai/go-bybit/ws"
)

// V5WebsocketSpotServiceI :
type V5WebsocketSpotServiceI interface {
}

// V5WebsocketSpotService :
type V5WebsocketSpotService struct {
	Client *ws.WebSocketClient
}
