package wsv1

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit/ws"
)

// PrivateService :
type PrivateService struct {
	client     *ws.WebSocketClient
	connection *websocket.Conn

	paramOutboundAccountInfoMap map[PrivateParamKey]func(PrivateOutboundAccountInfoResponse) error
}

const (
	// PrivatePath :
	PrivatePath = "/spot/ws"
)

// PrivateEventType :
type PrivateEventType string

const (
	// OutboundAccountInfo :
	OutboundAccountInfo = "outboundAccountInfo"
)

// PrivateParamKey :
type PrivateParamKey struct {
	EventType PrivateEventType
}

// PrivateOutboundAccountInfoResponse :
type PrivateOutboundAccountInfoResponse struct {
	Content PrivateOutboundAccountInfoResponseContent
}

// PrivateOutboundAccountInfoResponseContent :
type PrivateOutboundAccountInfoResponseContent struct {
	EventType            PrivateEventType      `json:"e"`
	Timestamp            string                `json:"E"`
	AllowTrade           bool                  `json:"T"`
	AllowWithdraw        bool                  `json:"W"`
	AllowWDeposit        bool                  `json:"D"`
	WalletBalanceChanges []WalletBalanceChange `json:"B"`
}

// WalletBalanceChange :
type WalletBalanceChange struct {
	SymbolName       string `json:"a"`
	AvailableBalance string `json:"f"`
	ReservedBalance  string `json:"l"`
}

// UnmarshalJSON :
func (r *PrivateOutboundAccountInfoResponse) UnmarshalJSON(data []byte) error {
	parsedArrayData := []map[string]interface{}{}
	if err := json.Unmarshal(data, &parsedArrayData); err != nil {
		return err
	}
	if len(parsedArrayData) != 1 {
		return errors.New("unexpected response")
	}
	buf, err := json.Marshal(parsedArrayData[0])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, &r.Content); err != nil {
		return err
	}
	return nil
}

// MarshalJSON :
func (r *PrivateOutboundAccountInfoResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Content)
}

// Key :
func (r *PrivateOutboundAccountInfoResponse) Key() PrivateParamKey {
	return PrivateParamKey{
		EventType: r.Content.EventType,
	}
}

// addParamOutboundAccountInfoFunc :
func (s *PrivateService) addParamOutboundAccountInfoFunc(param PrivateParamKey, f func(PrivateOutboundAccountInfoResponse) error) error {
	if _, exist := s.paramOutboundAccountInfoMap[param]; exist {
		return errors.New("already registered for this param")
	}
	s.paramOutboundAccountInfoMap[param] = f
	return nil
}

// retrieveOutboundAccountInfoFunc :
func (s *PrivateService) retrieveOutboundAccountInfoFunc(key PrivateParamKey) (func(PrivateOutboundAccountInfoResponse) error, error) {
	f, exist := s.paramOutboundAccountInfoMap[key]
	if !exist {
		return nil, errors.New("func not found")
	}
	return f, nil
}

type spotWebsocketV1PrivateEventJudge struct {
	EventType PrivateEventType
}

func (r *spotWebsocketV1PrivateEventJudge) UnmarshalJSON(data []byte) error {
	parsedData := map[string]interface{}{}
	if err := json.Unmarshal(data, &parsedData); err == nil {
		if event, ok := parsedData["e"].(string); ok {
			r.EventType = PrivateEventType(event)
		}
		if authStatus, ok := parsedData["auth"].(string); ok {
			if authStatus != "success" {
				return errors.New("auth failed")
			}
		}
		return nil
	}

	parsedArrayData := []map[string]interface{}{}
	if err := json.Unmarshal(data, &parsedArrayData); err != nil {
		return err
	}
	if len(parsedArrayData) != 1 {
		return errors.New("unexpected response")
	}
	r.EventType = PrivateEventType(parsedArrayData[0]["e"].(string))
	return nil
}

// judgeEventType :
func (s *PrivateService) judgeEventType(respBody []byte) (PrivateEventType, error) {
	var result spotWebsocketV1PrivateEventJudge
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.EventType, nil
}

// parseResponse :
func (s *PrivateService) parseResponse(respBody []byte, response interface{}) error {
	if err := json.Unmarshal(respBody, &response); err != nil {
		return err
	}
	return nil
}

// Subscribe :
func (s *PrivateService) Subscribe() error {
	param, err := s.client.BuildAuthParam()
	if err != nil {
		return err
	}
	if err := s.connection.WriteMessage(websocket.TextMessage, param); err != nil {
		return err
	}
	return nil
}

// RegisterFuncOutboundAccountInfo :
func (s *PrivateService) RegisterFuncOutboundAccountInfo(f func(PrivateOutboundAccountInfoResponse) error) error {
	key := PrivateParamKey{
		EventType: OutboundAccountInfo,
	}
	if err := s.addParamOutboundAccountInfoFunc(key, f); err != nil {
		return err
	}
	return nil
}

// Start :
func (s *PrivateService) Start(ctx context.Context) {
	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			if err := s.Run(); err != nil {
				if ws.IsErrWebsocketClosed(err) {
					return
				}
				log.Println(err)
				return
			}
		}
	}()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := s.Ping(); err != nil {
				return
			}
		case <-ctx.Done():
			log.Println("interrupt")

			if err := s.Close(); err != nil {
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// Run :
func (s *PrivateService) Run() error {
	_, message, err := s.connection.ReadMessage()
	if err != nil {
		return err
	}

	topic, err := s.judgeEventType(message)
	if err != nil {
		return err
	}
	switch topic {
	case OutboundAccountInfo:
		var resp PrivateOutboundAccountInfoResponse
		if err := s.parseResponse(message, &resp); err != nil {
			return err
		}
		f, err := s.retrieveOutboundAccountInfoFunc(resp.Key())
		if err != nil {
			return err
		}
		if err := f(resp); err != nil {
			return err
		}
	}
	return nil
}

// Ping :
func (s *PrivateService) Ping() error {
	if err := s.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}
	return nil
}

// Close :
func (s *PrivateService) Close() error {
	if err := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return err
	}
	return nil
}
