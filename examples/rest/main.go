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
	proxy := proxyClient("socks5://127.0.0.1:1086")
	client := rest.NewClient().WithAuth("XAz3CVkZoz0jInEYo2", "2LZ5CTPq9UZgzZXs4eM4V3ZyHfCIP3LTM0Wt").WithTestnet().WithHTTPClient(proxy)

	symbol := bybit.SymbolV5BTCUSDT
	//// 获取InstrumentsInfo
	//res, err := client.V5().Market().GetInstrumentsInfo(rest.V5GetInstrumentsInfoParam{
	//	Category: "spot",
	//})
	//
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//log.Printf("InstrumentsInfo: %#v\n", res.Result.Spot.List)
	//
	//// 获取K线
	//klines, err := client.V5().Market().GetKline(rest.V5GetKlineParam{
	//	Category: "spot",
	//	Symbol:   symbol,
	//	Interval: bybit.IntervalD,
	//})
	//if err != nil {
	//	log.Printf("%v", err)
	//	return
	//}
	//
	//log.Printf("klines: %#v\n", klines)

	//res, err := client.V5().Position().GetPositionInfo(rest.V5GetPositionInfoParam{
	//	Category: bybit.CategoryV5Linear,
	//	Symbol:   &symbol,
	//})
	//
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Printf("positions: %#v\n", res.Result)

	res1, err := client.V5().Account().GetWalletBalance(bybit.AccountTypeUnified, []bybit.Coin{})
	if err != nil {
		log.Println(err)
	}
	log.Printf("positions: %#v\n", res1.Result)

	// 创建委托
	price := "23000.0"
	//timeInForce := bybit.TimeInForce("GoodTillCancel")
	order, err := client.V5().Order().CreateOrder(rest.V5CreateOrderParam{
		Category:  bybit.CategoryV5Spot,
		Symbol:    symbol,
		Side:      bybit.SideBuy,
		OrderType: bybit.OrderTypeLimit,
		Qty:       "0.01",
		Price:     &price,
		//TimeInForce: &timeInForce,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Create order: %#v", order)

	orderList, err := client.V5().Order().GetOpenOrders(rest.V5GetOpenOrdersParam{
		Category: bybit.CategoryV5Spot,
		Symbol:   &symbol,
	})
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("GetOpenOrders: %v", orderList)

	for _, o := range orderList.Result.List {
		cancel, err := client.V5().Order().CancelOrder(rest.V5CancelOrderParam{
			Category: bybit.CategoryV5Spot,
			Symbol:   symbol,
			OrderID:  &o.OrderID,
		})
		if err != nil {
			log.Printf("cancel order failed, error: %v\n", err)
			return
		}
		log.Printf("CancelOrder: %#v\n", cancel)
	}

	// 获取委托单
}
