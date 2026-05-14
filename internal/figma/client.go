package figma

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	url string

	mu             sync.Mutex
	conn           *websocket.Conn
	currentChannel string
	pending        map[string]*pendingRequest
	closed         chan struct{}
}

type pendingRequest struct {
	result   chan commandResult
	timer    *time.Timer
	deadline time.Duration
}

type commandResult struct {
	value any
	err   error
}

type relayEnvelope struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type,omitempty"`
	Channel string          `json:"channel,omitempty"`
	Message json.RawMessage `json:"message,omitempty"`
}

type figmaResponse struct {
	ID     string `json:"id"`
	Result any   `json:"result,omitempty"`
	Error  any   `json:"error,omitempty"`
}

func NewClient(url string) *Client {
	return &Client{
		url:     url,
		pending: make(map[string]*pendingRequest),
		closed:  make(chan struct{}),
	}
}

func (c *Client) ConnectAsync() {
	go c.reconnectLoop()
}

func (c *Client) reconnectLoop() {
	for {
		select {
		case <-c.closed:
			return
		default:
		}

		c.mu.Lock()
		alreadyConnected := c.conn != nil
		c.mu.Unlock()
		if alreadyConnected {
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("connecting to Figma relay at %s", c.url)
		conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
		if err != nil {
			log.Printf("relay connection failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.currentChannel = ""
		c.mu.Unlock()

		log.Printf("connected to Figma relay")
		c.readLoop(conn)
		time.Sleep(2 * time.Second)
	}
}

func (c *Client) Close() {
	close(c.closed)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) JoinChannel(channel string) (any, error) {
	if channel == "" {
		return nil, errors.New("channel is required")
	}
	result, err := c.SendCommand("join", map[string]any{"channel": channel}, 30*time.Second)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.currentChannel = channel
	c.mu.Unlock()
	return result, nil
}

func (c *Client) SendCommand(command string, params map[string]any, timeout time.Duration) (any, error) {
	c.mu.Lock()
	conn := c.conn
	channel := c.currentChannel
	c.mu.Unlock()

	if conn == nil {
		return nil, errors.New("not connected to Figma relay. Start the relay and wait for the server to reconnect")
	}
	if command != "join" && channel == "" {
		return nil, errors.New("must join a channel before sending commands")
	}
	if params == nil {
		params = map[string]any{}
	}

	id := uuid.NewString()
	params["commandId"] = id

	envelope := map[string]any{
		"id": id,
		"message": map[string]any{
			"id":      id,
			"command": command,
			"params":  params,
		},
	}
	if command == "join" {
		envelope["type"] = "join"
		envelope["channel"] = params["channel"]
	} else {
		envelope["type"] = "message"
		envelope["channel"] = channel
	}

	req := &pendingRequest{
		result:   make(chan commandResult, 1),
		deadline: timeout,
	}
	req.timer = time.AfterFunc(timeout, func() {
		c.finish(id, commandResult{err: errors.New("request to Figma timed out")})
	})

	c.mu.Lock()
	c.pending[id] = req
	c.mu.Unlock()

	log.Printf("sending command to Figma: %s", command)
	if err := conn.WriteJSON(envelope); err != nil {
		c.finish(id, commandResult{err: err})
		return nil, err
	}

	res := <-req.result
	return res.value, res.err
}

func (c *Client) readLoop(conn *websocket.Conn) {
	defer func() {
		c.mu.Lock()
		if c.conn == conn {
			c.conn = nil
			c.currentChannel = ""
		}
		for id := range c.pending {
			c.finishLocked(id, commandResult{err: errors.New("connection closed")})
		}
		c.mu.Unlock()
		_ = conn.Close()
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Printf("relay read failed: %v", err)
			return
		}

		var env relayEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			log.Printf("invalid relay message: %v", err)
			continue
		}

		if env.Type == "progress_update" {
			c.handleProgress(env)
			continue
		}

		if len(env.Message) == 0 {
			continue
		}

		var resp figmaResponse
		if err := json.Unmarshal(env.Message, &resp); err != nil || resp.ID == "" {
			log.Printf("relay event: %s", string(env.Message))
			continue
		}

		if resp.Error != nil {
			c.finish(resp.ID, commandResult{err: fmt.Errorf("%v", resp.Error)})
			continue
		}
		c.finish(resp.ID, commandResult{value: resp.Result})
	}
}

func (c *Client) handleProgress(env relayEnvelope) {
	var payload struct {
		Data struct {
			CommandType string `json:"commandType"`
			Progress    int    `json:"progress"`
			Message     string `json:"message"`
		} `json:"data"`
	}
	_ = json.Unmarshal(env.Message, &payload)
	log.Printf("progress %s %d%%: %s", payload.Data.CommandType, payload.Data.Progress, payload.Data.Message)

	id := env.ID
	if id == "" {
		return
	}

	c.mu.Lock()
	req := c.pending[id]
	if req != nil {
		req.timer.Reset(60 * time.Second)
	}
	c.mu.Unlock()
}

func (c *Client) finish(id string, result commandResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.finishLocked(id, result)
}

func (c *Client) finishLocked(id string, result commandResult) {
	req := c.pending[id]
	if req == nil {
		return
	}
	delete(c.pending, id)
	req.timer.Stop()
	req.result <- result
	close(req.result)
}
