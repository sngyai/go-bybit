package main

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
	"github.com/sngyai/go-bybit/ws"
	"github.com/sngyai/go-bybit/ws/wsv1"
	"github.com/sngyai/go-bybit/ws/wsv5"
	"golang.org/x/net/proxy"
)

func proxyClient(proxyURL string) websocket.Dialer {
	prx, _ := url.Parse(proxyURL) // some not-exist-proxy

	netDialer, _ := proxy.SOCKS5("tcp", prx.Host, nil, &net.Dialer{})
	return websocket.Dialer{NetDial: netDialer.Dial}
}

func main() {
	//client := proxyClient("socks5://127.0.0.1:1086")
	wsClient := ws.NewWebsocketClient()
	//.WithDialer(&client)
	//err := exampleV1(wsClient)
	//if err != nil {
	//	return
	//}

	err := exampleV5(wsClient)
	if err != nil {
		return
	}
}

func exampleV1(wsClient *ws.WebSocketClient) error {
	svc, err := wsv1.NewWSClient(wsClient).PublicV1()
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = svc.SubscribeTrade(bybit.SymbolSpotBTCUSDT, func(response wsv1.SpotWebsocketV1PublicV1TradeResponse) error {
		fmt.Printf("v1 recv: %v\n", response)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	svc.Start(context.Background())

	return nil
}

func exampleV5(wsClient *ws.WebSocketClient) error {
	svc, err := wsv5.NewWSClient(wsClient).Public(bybit.CategoryV5Spot)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = svc.SubscribeOrderBook(wsv5.PublicOrderBookParamKey{
		Depth:  5,
		Symbol: bybit.SymbolV5BTCUSDT,
	},
		func(response wsv5.PublicOrderBookResponse) error {
			fmt.Printf("v5 recv: %v\n", response)
			return nil
		})
	if err != nil {
		fmt.Println(err)
		return err
	}
	svc.Start(context.Background(), nil)
	return nil
}
