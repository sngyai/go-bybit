package wsv5

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
)

// SubscribeWallet :
func (s *PrivateService) SubscribeWallet(
	f func(PrivateWalletResponse) error,
) (func() error, error) {
	key := PrivateParamKey{
		Topic: PrivateTopicWallet,
	}
	if err := s.addParamWalletFunc(key, f); err != nil {
		return nil, err
	}
	param := struct {
		Op   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		Op:   "subscribe",
		Args: []interface{}{PrivateTopicWallet},
	}
	buf, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	if err := s.connection.WriteMessage(websocket.TextMessage, buf); err != nil {
		return nil, err
	}
	return func() error {
		param := struct {
			Op   string        `json:"op"`
			Args []interface{} `json:"args"`
		}{
			Op:   "unsubscribe",
			Args: []interface{}{PrivateTopicWallet},
		}
		buf, err := json.Marshal(param)
		if err != nil {
			return err
		}
		if err := s.connection.WriteMessage(websocket.TextMessage, []byte(buf)); err != nil {
			return err
		}
		s.removeParamWalletFunc(key)
		return nil
	}, nil
}

// PrivateWalletResponse :
type PrivateWalletResponse struct {
	ID           string                         `json:"id"`
	Topic        PrivateTopic                   `json:"topic"`
	CreationTime int64                          `json:"creationTime"`
	Data         []V5WebsocketPrivateWalletData `json:"data"`
}

// V5WebsocketPrivateWalletData :
type V5WebsocketPrivateWalletData struct {
	AccountIMRate          string              `json:"accountIMRate"`
	AccountMMRate          string              `json:"accountMMRate"`
	TotalEquity            string              `json:"totalEquity"`
	TotalWalletBalance     string              `json:"totalWalletBalance"`
	TotalMarginBalance     string              `json:"totalMarginBalance"`
	TotalAvailableBalance  string              `json:"totalAvailableBalance"`
	TotalPerpUPL           string              `json:"totalPerpUPL"`
	TotalInitialMargin     string              `json:"totalInitialMargin"`
	TotalMaintenanceMargin string              `json:"totalMaintenanceMargin"`
	AccountType            bybit.AccountType   `json:"accountType"`
	Coins                  []PrivateWalletCoin `json:"coin"`
}

// PrivateWalletCoin :
type PrivateWalletCoin struct {
	Coin                bybit.Coin `json:"coin"`
	Equity              string     `json:"equity"`
	UsdValue            string     `json:"usdValue"`
	WalletBalance       string     `json:"walletBalance"`
	AvailableToWithdraw string     `json:"availableToWithdraw"`
	AvailableToBorrow   string     `json:"availableToBorrow"`
	BorrowAmount        string     `json:"borrowAmount"`
	AccruedInterest     string     `json:"accruedInterest"`
	TotalOrderIM        string     `json:"totalOrderIM"`
	TotalPositionIM     string     `json:"totalPositionIM"`
	TotalPositionMM     string     `json:"totalPositionMM"`
	UnrealisedPnl       string     `json:"unrealisedPnl"`
	CumRealisedPnl      string     `json:"cumRealisedPnl"`
}

// Key :
func (r *PrivateWalletResponse) Key() PrivateParamKey {
	return PrivateParamKey{
		Topic: r.Topic,
	}
}

// addParamWalletFunc :
func (s *PrivateService) addParamWalletFunc(param PrivateParamKey, f func(PrivateWalletResponse) error) error {
	if _, exist := s.paramWalletMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramWalletMap[param] = f
	return nil
}

// removeParamWalletFunc :
func (s *PrivateService) removeParamWalletFunc(key PrivateParamKey) {
	delete(s.paramWalletMap, key)
}

// retrieveWalletFunc :
func (s *PrivateService) retrieveWalletFunc(key PrivateParamKey) (func(PrivateWalletResponse) error, error) {
	f, exist := s.paramWalletMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}
