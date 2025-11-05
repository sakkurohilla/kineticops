package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
)

// No need to import github.com/gofiber/contrib/websocket here!

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
	maxClients int // Enterprise connection limit
	// lastMessages keeps the most recent message per host_id so new clients can
	// receive a warm-up snapshot when they connect. We store the message along
	// with the monotonic sequence id so older messages don't overwrite newer
	// ones.
	lastMessages map[int64]struct {
		Seq uint64
		Msg []byte
	}
	// broadcastQueue decouples producers from the actual per-client sends so a
	// single slow client won't block the publisher. Workers read from this
	// queue and fan-out messages to active clients.
	broadcastQueue chan []byte
	workerCount    int
	// maximum number of host warm-up messages to retain to avoid unbounded
	// memory growth in high-scale scenarios.
	maxLastMessages int
}

func NewHub() *Hub {
	workerCount := runtime.NumCPU()
	queueSize := 10000
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		maxClients: 10000, // Enterprise limit
		lastMessages: make(map[int64]struct {
			Seq uint64
			Msg []byte
		}),
		broadcastQueue:  make(chan []byte, queueSize),
		workerCount:     workerCount,
		maxLastMessages: 2000,
	}
}

func (h *Hub) Run() {
	// start broadcast worker pool
	for i := 0; i < h.workerCount; i++ {
		go func(id int) {
			for msg := range h.broadcastQueue {
				h.mu.Lock()
				var msgData map[string]interface{}
				targetUserID := int64(-1)
				if json.Unmarshal(msg, &msgData) == nil {
					if userID, ok := msgData["target_user_id"]; ok {
						if uid, ok := userID.(float64); ok {
							targetUserID = int64(uid)
						}
					}
				}

				for client := range h.clients {
					if targetUserID != -1 && client.userID != targetUserID {
						continue
					}
					select {
					case client.send <- msg:
					default:
						// If client's buffer is full, drop the client to reclaim resources
						close(client.send)
						delete(h.clients, client)
					}
				}
				h.mu.Unlock()
			}
		}(i)
	}

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Check connection limit
			if len(h.clients) >= h.maxClients {
				fmt.Printf("[WS HUB] Connection limit reached, rejecting client user=%d\n", client.userID)
				close(client.send)
				h.mu.Unlock()
				continue
			}
			h.clients[client] = true
			fmt.Printf("[WS HUB] registered client user=%d, total_clients=%d\n", client.userID, len(h.clients))
			// send warm-up messages (last known metrics) to the newly registered client
			// best-effort: skip warming when lastMessages map is too large
			if len(h.lastMessages) <= h.maxLastMessages {
				for _, entry := range h.lastMessages {
					select {
					case client.send <- entry.Msg:
					default:
						// skip if buffer full
					}
				}
			}
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				fmt.Printf("[WS HUB] unregistered client user=%d, total_clients=%d\n", client.userID, len(h.clients))
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			// enqueue into broadcastQueue; drop if queue is full to avoid blocking
			select {
			case h.broadcastQueue <- message:
			default:
				// queue full — drop message and log
				log.Printf("[WS HUB] broadcast queue full, dropping message\n")
			}
		}
	}
}

func (h *Hub) Broadcast(msg []byte) {
	// Lock while iterating/modifying the clients map to avoid concurrent map access
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			// if send would block, remove the client to avoid blocking the hub
			fmt.Printf("[WS HUB] removing client user=%d due to blocked send\n", c.userID)
			close(c.send)
			delete(h.clients, c)
		}
	}
}

// RememberMessage stores a message (if it contains host_id) for warm-up on new
// client registration.
func (h *Hub) RememberMessage(msg []byte) {
	// best-effort parse host_id from JSON
	var tmp map[string]interface{}
	if err := json.Unmarshal(msg, &tmp); err != nil {
		return
	}
	if hidRaw, ok := tmp["host_id"]; ok {
		// JSON numbers decode as float64
		if fid, ok2 := hidRaw.(float64); ok2 {
			hid := int64(fid)
			// extract seq (if present) and only replace stored message when seq is newer
			var seq uint64 = 0
			if sRaw, ok := tmp["seq"]; ok {
				switch s := sRaw.(type) {
				case float64:
					seq = uint64(s)
				case int64:
					seq = uint64(s)
				}
			}
			h.mu.Lock()
			existing, ok := h.lastMessages[hid]
			if !ok || seq > existing.Seq {
				h.lastMessages[hid] = struct {
					Seq uint64
					Msg []byte
				}{Seq: seq, Msg: msg}
			}
			h.mu.Unlock()
		}
	}
}

// Global hub reference for simple cross-package broadcasts (set from main)
var globalHub *Hub

// SetGlobalHub registers the hub instance so other packages can broadcast directly.
func SetGlobalHub(h *Hub) {
	globalHub = h
}

// GetGlobalHub returns the global hub instance (may be nil)
func GetGlobalHub() *Hub {
	return globalHub
}

// BroadcastToClients sends msg to all connected websocket clients if hub is available.
// This is a best-effort non-blocking call.
func BroadcastToClients(msg []byte) {
	if globalHub == nil {
		return
	}
	globalHub.Broadcast(msg)
}

// ClientCount returns the current number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// GetGlobalClientCount returns the number of connected clients via the global hub
func GetGlobalClientCount() int {
	if globalHub == nil {
		return 0
	}
	return globalHub.ClientCount()
}

//func (h *Hub) Broadcast(data []byte) {
//	h.broadcast <- data
//}
