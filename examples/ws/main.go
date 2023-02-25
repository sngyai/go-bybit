package main

import (
	"context"
	"fmt"

	"github.com/sngyai/go-bybit"
	"github.com/sngyai/go-bybit/ws"
	"github.com/sngyai/go-bybit/ws/v1"
)

func main() {
	wsClient := ws.NewWebsocketClient()
	svc, err := v1.NewWSClientV1(wsClient).PublicV1()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = svc.SubscribeTrade(bybit.SymbolSpotBTCUSDT, func(response v1.SpotWebsocketV1PublicV1TradeResponse) error {
		fmt.Printf("recv: %v\n", response)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	svc.Start(context.Background())
}
