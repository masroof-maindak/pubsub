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

		err = wsSubscriberTransmission(ws, r, ct)

		if err != nil {
			logger.Log.Err(err).Msg("websocket died")
		}
	}
}

func wsSubscriberTransmission(ws *websocket.Conn, r *http.Request, ct utils.ClientType) error {
	stopReading := make(chan error)
	pingTicker := time.NewTicker(pingPeriod)

	if ct == utils.Subscriber {
		resolver.OnSubscriberConnect(ws, r, stopReading)
	}

	defer func() {
		if ct == utils.Subscriber {
			resolver.OnSubscriberDisconnect(ws)
		}

		pingTicker.Stop()

		close(stopReading)
	}()

	go readFromWs(ws, stopReading, ct)

	for {
		select {
		case reason := <-stopReading: // Hit when this is chan is closed from the other side
			logger.Log.Info().Err(reason).Msg("WS reading stopped")
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

func readFromWs(ws *websocket.Conn, readerFinished chan error, ct utils.ClientType) {
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(appdata string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msgJson map[string]any

		// Blocks on read call. When the channel is closed, an error is returned here.
		err := ws.ReadJSON(&msgJson)
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				readerFinished <- nil
			} else {
				readerFinished <- fmt.Errorf("reading from connection failed: %w", err)
			}
			break
		}

		if ct == utils.Subscriber {
			// Parse recieved message as 'SUBSCRIBE' or 'UNSUBSCRIBE'
			// Ensure topic is provided
			action, ok1 := msgJson["action"].(string)
			topic, ok2 := msgJson["topic"].(string)
			if !ok1 && !ok2 {
				closeErr := fmt.Errorf("json object didn't contain key 'text'")

				logger.Log.Error().Err(closeErr)

				err := utils.WriteCloseMsg(ws, websocket.CloseUnsupportedData, closeErr)

				if err != nil {
					readerFinished <- fmt.Errorf("failed to send close message: %w", err)
				} else {
					readerFinished <- closeErr
				}

				break
			}

			// Subscribe or unsubscribe
			var actionErr error = nil

			switch action {
			case "SUBSCRIBE":
				actionErr = resolver.OnSubscriberSubscribe(ws, readerFinished, topic)
				if actionErr != nil {
					actionErr = fmt.Errorf("failed to subscribe to topic [%s] w/ err: %w", topic, actionErr)
				}
			case "UNSUBSCRIBE":
				actionErr = resolver.OnSubscriberUnsubscribe(ws, topic)
				if actionErr != nil {
					actionErr = fmt.Errorf("failed to unsubscribe from topic [%s] w/ err: %w", topic, actionErr)
				}
			default:
				actionErr = fmt.Errorf("Invalid action provided. Valid options are SUBSCRIBE or UNSUBSCRIBE")
			}

			if actionErr != nil {
				utils.WriteCloseMsg(ws, websocket.CloseUnsupportedData, actionErr)
				readerFinished <- actionErr
				break
			}

		} else { // Publisher

			// Parse recieved message as "topic", "text"
			text, ok1 := msgJson["text"].(string)
			topic, ok2 := msgJson["topic"].(string)
			if !ok1 && !ok2 {
				closeErr := fmt.Errorf("json object didn't contain key 'text' or 'topic'")

				err := utils.WriteCloseMsg(ws, websocket.CloseUnsupportedData, closeErr)

				if err != nil {
					readerFinished <- fmt.Errorf("failed to send close message: %w", err)
				} else {
					readerFinished <- closeErr
				}

				break
			}

			// Send to subscribers
			resolver.OnPublisherPublish(topic, text)
		}
	}
}
