package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/masroof-maindak/pubsub/internal/logger"
	"github.com/masroof-maindak/pubsub/internal/resolver"
	"github.com/masroof-maindak/pubsub/internal/utils"

	"github.com/gorilla/websocket"
)

const (
	pongWait   = 20 * time.Second
	pingPeriod = pongWait * 9 / 10
)

var (
	upgrader = websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func CreateWsConn(ct utils.ClientType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			_, ok := err.(websocket.HandshakeError)
			if !ok {
				logger.Log.Error().Err(err).Msg("error establishing websocket connection")
			}
			return
		}
		defer ws.Close()

		if ct == utils.Subscriber {
			err = wsSubscriberTransmission(ws, r)
		} else {
			// read from publisher
		}

		if err != nil {
			logger.Log.Err(err).Msg("websocket died")
		}
	}
}

func wsSubscriberTransmission(ws *websocket.Conn, r *http.Request) error {
	stopReading := make(chan error)
	pingTicker := time.NewTicker(pingPeriod)

	resolver.OnSubscriberConnect(ws, r, stopReading)

	defer func() {
		resolver.OnSubscriberDisconnect(ws)
		pingTicker.Stop()
		ws.Close()
	}()

	go readFromSubscriberWs(ws, stopReading)

	for {
		select {
		case reason := <-stopReading: // if someone signalled the end of reading or wants us to be closed
			return reason
		case <-pingTicker.C: // Send sporadic pings
			err := ws.WriteControl(
				websocket.PingMessage,
				[]byte{},
				time.Now().Add(utils.CloseDeadline),
			)

			if err != nil {
				return fmt.Errorf("failed to send ping: %w", err)
			}
		}
	}
}

// FIXME
func readFromSubscriberWs(ws *websocket.Conn, readerFinished chan error) {
	defer close(readerFinished)

	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(appdata string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msgJson map[string]any

		// Blocks on read call. Closes return an error here.
		err := ws.ReadJSON(&msgJson)
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				readerFinished <- nil
			} else {
				readerFinished <- fmt.Errorf("reading from connection failed: %w", err)
			}
			break
		}

		// TODO: Parse recieved message as 'SUBSCRIBE' or 'UNSUBSCRIBE'
		// Ensure topic is provided
		textStr, ok := msgJson["text"].(string)
		if !ok {
			closeErr := fmt.Errorf("json object didn't contain key 'text'")

			err := utils.WriteCloseMsg(ws, websocket.CloseUnsupportedData, closeErr)

			if err != nil {
				readerFinished <- fmt.Errorf("failed to send close message: %w", err)
			} else {
				readerFinished <- closeErr
			}

			break
		}

		// TODO: Propagate message to OnSubscriber{Un/Subscribe}
		err = resolver.OnSubscriberSubscribe(ws, &readerFinished, textStr)
		if err != nil {
			readerFinished <- fmt.Errorf("resolver failed to write w/ err: %w", err)
			break
		}
	}
}
