package relay

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type hub struct {
	mu       sync.RWMutex
	channels map[string]map[*websocket.Conn]struct{}
}

func newHub() *hub {
	return &hub{channels: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *hub) join(channel string, conn *websocket.Conn) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*websocket.Conn]struct{})
	}
	h.channels[channel][conn] = struct{}{}
	return len(h.channels[channel])
}

func (h *hub) leave(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for channel, clients := range h.channels {
		if _, ok := clients[conn]; ok {
			delete(clients, conn)
			for peer := range clients {
				_ = peer.WriteJSON(map[string]any{
					"type":    "system",
					"message": "A user has left the channel",
					"channel": channel,
				})
			}
		}
		if len(clients) == 0 {
			delete(h.channels, channel)
		}
	}
}

func (h *hub) contains(channel string, conn *websocket.Conn) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.channels[channel]
	if clients == nil {
		return false
	}
	_, ok := clients[conn]
	return ok
}

func (h *hub) broadcast(channel string, sender *websocket.Conn, payload any) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.channels[channel]
	count := 0
	for peer := range clients {
		if peer == sender {
			continue
		}
		if err := peer.WriteJSON(payload); err == nil {
			count++
		}
	}
	return count
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ListenAndServe starts the channel-based WebSocket relay used by the MCP
// server and the Figma plugin. It mirrors src/socket.ts in the original repo.
func ListenAndServe(addr string) error {
	h := newHub()
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORS(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Header.Get("Upgrade") == "" {
			writeCORS(w)
			_, _ = w.Write([]byte("WebSocket server running"))
			return
		}

		conn, err := upgrader.Upgrade(w, r, http.Header{
			"Access-Control-Allow-Origin": []string{"*"},
		})
		if err != nil {
			log.Printf("upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		defer h.leave(conn)

		log.Printf("client connected")
		_ = conn.WriteJSON(map[string]any{
			"type":    "system",
			"message": "Please join a channel to start chatting",
		})

		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				log.Printf("client disconnected: %v", err)
				return
			}

			var msg map[string]any
			if err := json.Unmarshal(raw, &msg); err != nil {
				_ = conn.WriteJSON(map[string]any{"type": "error", "message": "Invalid JSON"})
				continue
			}

			msgType, _ := msg["type"].(string)
			channel, _ := msg["channel"].(string)
			id, _ := msg["id"].(string)
			log.Printf("message type=%s channel=%s id=%s", msgType, channel, id)

			switch msgType {
			case "join":
				if channel == "" {
					_ = conn.WriteJSON(map[string]any{"type": "error", "message": "Channel name is required"})
					continue
				}
				size := h.join(channel, conn)
				log.Printf("client joined channel %q (%d clients)", channel, size)

				_ = conn.WriteJSON(map[string]any{
					"type":    "system",
					"message": "Joined channel: " + channel,
					"channel": channel,
				})
				_ = conn.WriteJSON(map[string]any{
					"type": "system",
					"message": map[string]any{
						"id":     id,
						"result": "Connected to channel: " + channel,
					},
					"channel": channel,
				})

				h.broadcast(channel, conn, map[string]any{
					"type":    "system",
					"message": "A new user has joined the channel",
					"channel": channel,
				})

			case "message":
				if channel == "" {
					_ = conn.WriteJSON(map[string]any{"type": "error", "message": "Channel name is required"})
					continue
				}
				if !h.contains(channel, conn) {
					_ = conn.WriteJSON(map[string]any{"type": "error", "message": "You must join the channel first"})
					continue
				}
				count := h.broadcast(channel, conn, map[string]any{
					"type":    "broadcast",
					"message": msg["message"],
					"sender":  "peer",
					"channel": channel,
				})
				log.Printf("broadcasted to %d peers in channel %q", count, channel)

			case "progress_update":
				if channel == "" || !h.contains(channel, conn) {
					continue
				}
				h.broadcast(channel, conn, msg)
			}
		}
	})

	log.Printf("WebSocket relay listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func writeCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
