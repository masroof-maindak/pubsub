package resolver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/masroof-maindak/pubsub/internal/db"
	"github.com/masroof-maindak/pubsub/internal/logger"

	"github.com/gorilla/websocket"
)

type TopicMap struct {
	mu    sync.Mutex
	conns map[string][]*websocket.Conn
}

var (
	m = TopicMap{
		conns: make(map[string][]*websocket.Conn),
	}

	unsubscribed []*websocket.Conn
)

func OnSubscriberConnect(
	ws *websocket.Conn,
	r *http.Request,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	unsubscribed = append(unsubscribed, ws)
}

func OnSubscriberDisconnect(
	ws *websocket.Conn,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ok := false
	unsubscribed, ok = removeWsFromSliceIfExists(ws, unsubscribed)
	if ok {
		return
	}

	// If it wasn't in the 'unsubscribed' list, then iterate through all topics, find the correct ws, and remove it
	for t, topicSubcribers := range m.conns {
		updatedSubs, _ := removeWsFromSliceIfExists(ws, topicSubcribers)
		m.conns[t] = updatedSubs
	}
}

func OnSubscriberSubscribe(
	ws *websocket.Conn,
	topic string,
) error {
	m.mu.Lock()

	topicSubcribers, ok := m.conns[topic]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("Topic doesn't exist.")
	}
	topicSubcribers = append(topicSubcribers, ws)
	m.conns[topic] = topicSubcribers

	m.mu.Unlock()

	lastMsg, err := db.GetLatestMessage(topic)
	if err != nil {
		return err
	}

	if lastMsg != "" {
		historyPayload := fmt.Sprintf("[HISTORY] %s", lastMsg)
		err = ws.WriteMessage(websocket.TextMessage, []byte(historyPayload))
		if err != nil {
			return fmt.Errorf("failed to send history: %w", err)
		}
	}

	// TEMPORARY
	// TODO: print current count of all lists

	return nil
}

func OnSubscriberUnsubscribe(
	ws *websocket.Conn,
	topic string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Topic existence check
	topicSubcribers, ok := m.conns[topic]
	if !ok {
		return fmt.Errorf("Topic doesn't exist.")
	}

	count := 0
	wsIdxInSearchTopic := 0

	// Count number of topics with this ws in it
	for t, topicSubcribers := range m.conns {
		wsIdx := 0
		wsIdx = getWsIndexInSlice(ws, topicSubcribers)

		if wsIdx == -1 { // not found
			if t == topic {
				return fmt.Errorf("Topic was never subscribed to in the first place!")
			}
		} else { // found
			count += 1
			if t == topic {
				wsIdxInSearchTopic = wsIdx
			}
		}
	}

	// Remove ws from that topic
	topicSubcribers, removedSc := removeWsFromSlice(wsIdxInSearchTopic, topicSubcribers)
	m.conns[topic] = topicSubcribers
	count -= 1

	// If ws is in no more topics, add it to unsubscribed
	if count == 0 {
		unsubscribed = append(unsubscribed, removedSc)
	}

	return nil
}

func OnPublisherPublish(topic string, message string) {
	// Save message to DB
	err := db.SaveLatestMessage(topic, message)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Database error")
	}

	m.mu.Lock()

	// Create topic if not exists
	topicSubcribers, ok := m.conns[topic]
	if !ok {
		m.conns[topic] = make([]*websocket.Conn, 0)
		topicSubcribers = m.conns[topic]
	}
	m.mu.Unlock()

	// Ping all topic subscribers
	for _, ws := range topicSubcribers {
		err := ws.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			OnSubscriberDisconnect(ws)
		}
	}
}

func removeWsFromSliceIfExists(ws *websocket.Conn, s []*websocket.Conn) ([]*websocket.Conn, bool) {
	i := getWsIndexInSlice(ws, s)

	if i != -1 {
		s, _ = removeWsFromSlice(i, s)
		return s, true
	}

	return s, false
}

func getWsIndexInSlice(ws *websocket.Conn, sockets []*websocket.Conn) int {
	i := -1

	for idx, websocket := range sockets {
		if websocket == ws {
			i = idx
		}
	}

	return i
}

func removeWsFromSlice(i int, s []*websocket.Conn) ([]*websocket.Conn, *websocket.Conn) {
	val := s[i]
	s[i] = s[len(s)-1]
	s = s[:len(s)-1]
	return s, val
}

func LoadTopicsFromDB() error {
	topics, err := db.GetAllTopics()
	if err != nil {
		return fmt.Errorf("Failed to get topics from DB: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range topics {
		m.conns[t] = make([]*websocket.Conn, 0)
	}

	return nil
}
