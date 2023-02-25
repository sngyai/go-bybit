package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/sngyai/go-bybit"
	"github.com/sngyai/go-bybit/rest"
)

// HttpProxy  = "http://127.0.0.1:6152"
// SocksProxy = "socks5://127.0.0.1:6153"
func proxyClient(proxyURL string) *http.Client {
	if proxyURL == "" {
		return nil
	}
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyURL)
	}

	httpTransport := &http.Transport{
		Proxy: proxy,
	}

	httpClient := &http.Client{
		Transport: httpTransport,
	}
	return httpClient
}

func main() {
	client := proxyClient("socks5://127.0.0.1:1086")
	b := rest.NewClient().WithAuth("XAz3CVkZoz0jInEYo2", "2LZ5CTPq9UZgzZXs4eM4V3ZyHfCIP3LTM0Wt").
		WithTestnet().WithHTTPClient(client)

	symbol := bybit.SymbolV5BTCUSDT
	// 获取持仓
	res, err := b.V5().Position().GetPositionInfo(rest.V5GetPositionInfoParam{
		Category: bybit.CategoryV5Linear,
		Symbol:   &symbol,
	})

	if err != nil {
		log.Println(err)
	}
	log.Printf("positions: %#v\n", res.Result)

	// 获取K线
	klines, err := b.V5().Market().GetKline(rest.V5GetKlineParam{
		Category: "spot",
		Symbol:   "AAVEUSDT",
		Interval: bybit.IntervalD,
	})
	if err != nil {
		log.Printf("%v", err)
		return
	}

	log.Printf("klines: %#v\n", klines)

	//// 创建委托
	//symbol := "BTCUSD"
	//side := "Buy"
	//orderType := "Limit"
	//qty := 10
	//price := 23000.0
	//timeInForce := "GoodTillCancel"
	//_, _, order, err := b.CreateOrder(side, orderType, price, qty, timeInForce, 0, 0, false, false, "", symbol)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//log.Printf("Create order: %#v", order)

	// 获取委托单
}
