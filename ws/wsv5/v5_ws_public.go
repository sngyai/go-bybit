package wsv5

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sngyai/go-bybit"
	"github.com/sngyai/go-bybit/ws"
)

// PublicServiceI :
type PublicServiceI interface {
	Start(context.Context, ErrHandler) error
	Run() error
	Ping() error
	Close() error

	SubscribeOrderBook(
		PublicOrderBookParamKey,
		func(PublicOrderBookResponse) error,
	) (func() error, error)
}

// PublicService :
type PublicService struct {
	client     *ws.WebSocketClient
	connection *websocket.Conn

	paramOrderBookMap map[PublicOrderBookParamKey]func(PublicOrderBookResponse) error
}

const (
	// PublicPath :
	PublicPath = "/v5/public"
)

// PublicPathFor :
func PublicPathFor(category bybit.CategoryV5) string {
	return PublicPath + "/" + string(category)
}

// PublicTopic :
type PublicTopic string

const (
	// PublicTopicOrderBook :
	PublicTopicOrderBook = "orderbook"
)

// judgeTopic :
func (s *PublicService) judgeTopic(respBody []byte) (PublicTopic, error) {
	parsedData := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &parsedData); err != nil {
		return "", err
	}
	if topic, ok := parsedData["topic"].(string); ok {
		switch {
		case strings.Contains(topic, "orderbook"):
			return PublicTopicOrderBook, nil
		}
	}
	return "", nil
}

// parseResponse :
func (s *PublicService) parseResponse(respBody []byte, response interface{}) error {
	if err := json.Unmarshal(respBody, &response); err != nil {
		return err
	}
	return nil
}

// Start :
func (s *PublicService) Start(ctx context.Context, errHandler ErrHandler) error {
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
func (s *PublicService) Run() error {
	_, message, err := s.connection.ReadMessage()
	if err != nil {
		return err
	}

	topic, err := s.judgeTopic(message)
	if err != nil {
		return err
	}
	switch topic {
	case PublicTopicOrderBook:
		var resp PublicOrderBookResponse
		if err := s.parseResponse(message, &resp); err != nil {
			return err
		}
		f, err := s.retrieveOrderBookFunc(resp.Key())
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
func (s *PublicService) Ping() error {
	if err := s.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}
	return nil
}

// Close :
func (s *PublicService) Close() error {
	if err := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return err
	}
	return nil
}
