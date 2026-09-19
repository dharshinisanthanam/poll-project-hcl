package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"

	"live-polling/backend/repository"

	"github.com/redis/go-redis/v9"
)

type Hub struct {
	redisRepo     *repository.RedisRepo
	clients       map[string]map[*Client]bool // pollID -> set of Clients
	subscriptions map[string]*redis.PubSub    // pollID -> Redis subscription
	subCancels    map[string]context.CancelFunc
	Register      chan *Client
	Unregister    chan *Client
	Broadcast     chan *BroadcastMessage
	mutex         sync.RWMutex
}

type BroadcastMessage struct {
	PollID  string
	Message []byte
}

func NewHub(redisRepo *repository.RedisRepo) *Hub {
	return &Hub{
		redisRepo:     redisRepo,
		clients:       make(map[string]map[*Client]bool),
		subscriptions: make(map[string]*redis.PubSub),
		subCancels:    make(map[string]context.CancelFunc),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Broadcast:     make(chan *BroadcastMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			if _, ok := h.clients[client.PollID]; !ok {
				h.clients[client.PollID] = make(map[*Client]bool)
				// Start Redis subscriber for this pollID
				h.startRedisSubscription(client.PollID)
			}
			h.clients[client.PollID][client] = true
			h.mutex.Unlock()

		case client := <-h.Unregister:
			h.mutex.Lock()
			if pollClients, ok := h.clients[client.PollID]; ok {
				if _, exists := pollClients[client]; exists {
					delete(pollClients, client)
					close(client.Send)

					// If no clients left watching this poll, cleanup Redis subscription
					if len(pollClients) == 0 {
						delete(h.clients, client.PollID)
						h.stopRedisSubscription(client.PollID)
					}
				}
			}
			h.mutex.Unlock()

		case bm := <-h.Broadcast:
			h.mutex.RLock()
			if pollClients, ok := h.clients[bm.PollID]; ok {
				for client := range pollClients {
					select {
					case client.Send <- bm.Message:
					default:
						close(client.Send)
						delete(pollClients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) startRedisSubscription(pollID string) {
	ctx, cancel := context.WithCancel(context.Background())
	pubsub := h.redisRepo.SubscribePollUpdates(ctx, pollID)
	h.subscriptions[pollID] = pubsub
	h.subCancels[pollID] = cancel

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			h.Broadcast <- &BroadcastMessage{
				PollID:  pollID,
				Message: []byte(msg.Payload),
			}
		}
	}()
}

func (h *Hub) stopRedisSubscription(pollID string) {
	if pubsub, ok := h.subscriptions[pollID]; ok {
		_ = pubsub.Close()
		delete(h.subscriptions, pollID)
	}
	if cancel, ok := h.subCancels[pollID]; ok {
		cancel()
		delete(h.subCancels, pollID)
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request, pollID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		Hub:    h,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		PollID: pollID,
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
