package wsv5

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

// PrivateServiceI :
type PrivateServiceI interface {
	Start(context.Context, ErrHandler) error
	Subscribe() error
	Run() error
	Ping() error
	Close() error

	SubscribeOrder(
		func(PrivateOrderResponse) error,
	) (func() error, error)

	SubscribePosition(
		func(PrivatePositionResponse) error,
	) (func() error, error)
}

// PrivateService :
type PrivateService struct {
	client     *ws.WebSocketClient
	connection *websocket.Conn

	paramOrderMap    map[PrivateParamKey]func(PrivateOrderResponse) error
	paramPositionMap map[PrivateParamKey]func(PrivatePositionResponse) error
	paramWalletMap   map[PrivateParamKey]func(PrivateWalletResponse) error
}

const (
	// PrivatePath :
	PrivatePath = "/v5/private"
)

// PrivateTopic :
type PrivateTopic string

const (
	// PrivateTopicOrder :
	PrivateTopicOrder = "order"

	// PrivateTopicPosition :
	PrivateTopicPosition = "position"

	// PrivateTopicWallet :
	PrivateTopicWallet = "wallet"
)

// PrivateParamKey :
type PrivateParamKey struct {
	Topic PrivateTopic
}

// judgeTopic :
func (s *PrivateService) judgeTopic(respBody []byte) (PrivateTopic, error) {
	parsedData := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &parsedData); err != nil {
		return "", err
	}
	if topic, ok := parsedData["topic"].(string); ok {
		return PrivateTopic(topic), nil
	}
	if authStatus, ok := parsedData["success"].(bool); ok {
		if !authStatus {
			return "", errors.New("auth failed: " + parsedData["ret_msg"].(string))
		}
	}
	return "", nil
}

// parseResponse :
func (s *PrivateService) parseResponse(respBody []byte, response interface{}) error {
	if err := json.Unmarshal(respBody, &response); err != nil {
		return err
	}
	return nil
}

// Subscribe : Apply for authentication when establishing a connection.
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

// ErrHandler :
type ErrHandler func(isWebsocketClosed bool, err error)

// Start :
func (s *PrivateService) Start(ctx context.Context, errHandler ErrHandler) error {
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer s.connection.Close()
		_ = s.connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		s.connection.SetPongHandler(func(string) error {
			_ = s.connection.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			if err := s.Run(); err != nil {
				errHandler(ws.IsErrWebsocketClosed(err), err)
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
			return nil
		case <-ticker.C:
			if err := s.Ping(); err != nil {
				return err
			}
		case <-ctx.Done():
			log.Println("interrupt")

			if err := s.Close(); err != nil {
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

// Run :
func (s *PrivateService) Run() error {
	_, message, err := s.connection.ReadMessage()
	if err != nil {
		return err
	}

	topic, err := s.judgeTopic(message)
	if err != nil {
		return err
	}
	switch topic {
	case PrivateTopicOrder:
		var resp PrivateOrderResponse
		if err := s.parseResponse(message, &resp); err != nil {
			return err
		}
		f, err := s.retrieveOrderFunc(resp.Key())
		if err != nil {
			return err
		}
		if err := f(resp); err != nil {
			return err
		}
	case PrivateTopicPosition:
		var resp PrivatePositionResponse
		if err := s.parseResponse(message, &resp); err != nil {
			return err
		}
		f, err := s.retrievePositionFunc(resp.Key())
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
