package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

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
	client := proxyClient("socks5://127.0.0.1:1086")
	wsClient := ws.NewWebsocketClient().WithBaseURL(ws.TestWebsocketBaseURL).WithAuth("APIKey", "APISecret").WithDialer(&client)

	//err := v1Multiple(wsClient)
	//if err != nil {
	//	return
	//}

	err := v5Single(wsClient)
	if err != nil {
		return
	}
	time.Sleep(3600 * time.Second)
}

func v1Single(wsClient *ws.WebSocketClient) error {
	svc, err := wsv1.NewWSClient(wsClient).PublicV1()
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = svc.SubscribeTrade(bybit.SymbolSpotBTCUSDT, func(response wsv1.PublicV1TradeResponse) error {
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

func v1Multiple(wsClient *ws.WebSocketClient) error {
	var executors []ws.WebsocketExecutor
	svcRoot := wsv1.NewWSClient(wsClient)
	{
		svc, err := svcRoot.PublicV1()
		if err != nil {
			return err
		}
		_, err = svc.SubscribeTrade(bybit.SymbolSpotBTCUSDT, func(response wsv1.PublicV1TradeResponse) error {
			fmt.Printf("v1 recv bybit.SymbolSpotBTCUSDT1: %v\n", response)
			return nil
		})
		if err != nil {
			return err
		}
		executors = append(executors, svc)
	}
	{
		svc, err := svcRoot.PublicV2()
		if err != nil {
			return err
		}
		_, err = svc.SubscribeTrade(bybit.SymbolSpotBTCUSDT, func(response wsv1.PublicV2TradeResponse) error {
			fmt.Printf("v1 recv bybit.SymbolSpotBTCUSDT2: %v\n", response)
			return nil
		})
		if err != nil {
			return err
		}
		executors = append(executors, svc)
	}

	wsClient.Start(context.Background(), executors)

	return nil
}

func v5Single(wsClient *ws.WebSocketClient) error {
	svc, err := wsv5.NewWSClient(wsClient).Public(bybit.CategoryV5Spot)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = svc.SubscribeTickers(wsv5.PublicTickersParamKey{
		Symbol: bybit.SymbolV5BTCUSDT,
	},
		func(response wsv5.PublicTickersResponse) error {
			fmt.Printf("v5 recv tickers: %v\n", response)
			return nil
		})
	if err != nil {
		fmt.Println(err)
		return err
	}

	errHandler := func(isWebsocketClosed bool, err error) {
		fmt.Printf("handle ws failed, isWebsocketClosed: %b, err: %v", isWebsocketClosed, err)
	}
	go func() {
		svc.Start(context.Background(), errHandler)
		v5Single(wsClient)
	}()
	return nil
}

func v5Multiple(wsClient *ws.WebSocketClient) error {
	var executors []ws.WebsocketExecutor
	var unsubscribe1, unsubscribe2 func() error
	svcRoot := wsv5.NewWSClient(wsClient)
	{
		svc, err := svcRoot.Public(bybit.CategoryV5Spot)
		if err != nil {
			return err
		}
		unsubscribe1, err = svc.SubscribeTickers(
			wsv5.PublicTickersParamKey{
				Symbol: bybit.SymbolV5BTCUSDT,
			},
			func(response wsv5.PublicTickersResponse) error {
				fmt.Printf("v5 recv ticker: %v\n", response)
				return nil
			})
		if err != nil {
			return err
		}
		executors = append(executors, svc)
	}
	{
		svc, err := svcRoot.Public(bybit.CategoryV5Spot)
		if err != nil {
			return err
		}
		unsubscribe2, err = svc.SubscribeOrderBook(wsv5.PublicOrderBookParamKey{
			Depth:  50,
			Symbol: bybit.SymbolV5BTCUSDT,
		},
			func(response wsv5.PublicOrderBookResponse) error {
				fmt.Printf("v5 recv orderbook: %v\n", response)
				return nil
			})
		if err != nil {
			return err
		}
		executors = append(executors, svc)
	}

	go func() {
		time.Sleep(3 * time.Second)
		unsubscribe1()
		fmt.Println("UnsubscribeTickers")
		time.Sleep(3 * time.Second)
		unsubscribe2()
		fmt.Println("UnsubscribeOrderBook")
	}()

	wsClient.Start(context.Background(), executors)

	return nil
}
