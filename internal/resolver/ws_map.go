package resolver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type SubscriberConn struct {
	socket  *websocket.Conn
	errChan *chan error
}

type TopicMap struct {
	mu    sync.Mutex
	conns map[string][]*SubscriberConn
}

var (
	m = TopicMap{
		conns: make(map[string][]*SubscriberConn),
	}

	unsubscribed []*SubscriberConn
)

func OnSubscriberConnect(
	ws *websocket.Conn,
	r *http.Request,
	readerFinished chan error,
) {

	m.mu.Lock()
	defer m.mu.Unlock()

	sc := SubscriberConn{ws, &readerFinished}
	unsubscribed = append(unsubscribed, &sc)
}

func OnSubscriberDisconnect(
	ws *websocket.Conn,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if removeWsFromSliceIfExists(ws, unsubscribed) {
		return
	}

	// If it wasn't in the 'unsubscribed' list, then iterate through all topics, find the correct ws, and remove it
	for _, topicSubcribers := range m.conns {
		removeWsFromSliceIfExists(ws, topicSubcribers)
	}
}

func OnSubscriberSubscribe(
	ws *websocket.Conn,
	errChan *chan error,
	topic string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for t, topicSubcribers := range m.conns {
		if t == topic {
			topicSubcribers = append(topicSubcribers, &SubscriberConn{ws, errChan})
			return nil
		}
	}

	return fmt.Errorf("Topic doesn't exist.")
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

		if wsIdx != -1 {
			count += 1
			if t == topic {
				return fmt.Errorf("Topic was never subscribed to in the first place!")
			}
		} else {
			if t == topic {
				wsIdxInSearchTopic = wsIdx
			}
		}

	}

	// Remove ws from that topic
	ret := removeWsFromSlice(wsIdxInSearchTopic, topicSubcribers)
	count -= 1

	// If ws is in no more topics, add it to unsubscribed
	if count == 0 {
		unsubscribed = append(unsubscribed, ret)
	}

	return nil
}

func removeWsFromSliceIfExists(ws *websocket.Conn, s []*SubscriberConn) bool {
	i := getWsIndexInSlice(ws, s)

	if i != -1 {
		removeWsFromSlice(i, s)
		return true
	}

	return false
}

func getWsIndexInSlice(ws *websocket.Conn, s []*SubscriberConn) int {
	i := -1

	for idx, sc := range s {
		if sc.socket == ws {
			i = idx
		}
	}

	return i
}

func removeWsFromSlice(i int, s []*SubscriberConn) *SubscriberConn {
	ret := s[i]
	s[i] = s[len(s)-1]
	s = s[:len(s)-1]
	return ret
}
