package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"figma-mcp-design/internal/figma"
)

type Server struct {
	figma *figma.Client
	tools []tool
}

func NewServer(figmaClient *figma.Client) *Server {
	return &Server{
		figma: figmaClient,
		tools: tools(),
	}
}

func (s *Server) Serve(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	encoder := json.NewEncoder(w)

	log.Printf("Figma MCP Go server running on stdio")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg jsonrpcMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("invalid JSON-RPC message: %v", err)
			continue
		}

		if msg.ID == nil {
			s.handleNotification(msg)
			continue
		}

		result, rpcErr := s.handleRequest(msg)
		resp := jsonrpcMessage{JSONRPC: "2.0", ID: msg.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) handleNotification(msg jsonrpcMessage) {
	switch msg.Method {
	case "notifications/initialized", "initialized", "$/cancelRequest":
		return
	default:
		log.Printf("ignored notification: %s", msg.Method)
	}
}

func (s *Server) handleRequest(msg jsonrpcMessage) (any, *rpcError) {
	switch msg.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities": map[string]any{
				"tools":   map[string]any{},
				"prompts": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":        "FigmaMCPDesign",
				"title":       "Figma MCP Design",
				"version":     "1.0.0",
				"releaseDate": "2026-05-13",
			},
		}, nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		return map[string]any{"tools": s.tools}, nil

	case "tools/call":
		result, err := s.callTool(msg.Params)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: err.Error()}
		}
		return result, nil

	case "prompts/list":
		ps := prompts()
		items := make([]map[string]string, 0, len(ps))
		for _, p := range ps {
			items = append(items, map[string]string{"name": p.Name, "description": p.Description})
		}
		return map[string]any{"prompts": items}, nil

	case "prompts/get":
		result, err := getPrompt(msg.Params)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: err.Error()}
		}
		return result, nil
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found: " + msg.Method}
	}
}

func (s *Server) callTool(raw json.RawMessage) (toolCallResult, error) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return errorResult(err), err
	}
	if params.Arguments == nil {
		params.Arguments = map[string]any{}
	}
	if !s.hasTool(params.Name) {
		err := fmt.Errorf("unknown tool: %s", params.Name)
		return errorResult(err), err
	}

	if err := validateRequired(params.Name, params.Arguments, s.tools); err != nil {
		return errorResult(err), err
	}

	result, err := s.dispatch(params.Name, params.Arguments)
	if err != nil {
		return toolCallResult{
			Content: []contentItem{{Type: "text", Text: fmt.Sprintf("Error calling %s: %v", params.Name, err)}},
			IsError: true,
		}, nil
	}

	return formatToolResult(params.Name, params.Arguments, result)
}

// dispatch issues the plugin command with default values that mirror the
func (s *Server) dispatch(name string, args map[string]any) (any, error) {
	switch name {
	case "join_channel":
		channel, _ := args["channel"].(string)
		return s.figma.JoinChannel(channel)

	case "get_nodes_info":
		return s.getNodesInfo(args)

	case "create_rectangle":
		payload := copyArgs(args)
		payload["name"] = nameOr(args["name"], "Rectangle")
		return s.figma.SendCommand(name, payload, 30*time.Second)

	case "create_frame":
		payload := copyArgs(args)
		payload["name"] = nameOr(args["name"], "Frame")
		payload["fillColor"] = defaultColor(args["fillColor"], map[string]any{"r": 1, "g": 1, "b": 1, "a": 1})
		return s.figma.SendCommand(name, payload, 30*time.Second)

	case "create_text":
		payload := copyArgs(args)
		payload["fontSize"] = numberOrDefault(args["fontSize"], 14)
		payload["fontWeight"] = numberOrDefault(args["fontWeight"], 400)
		payload["fontColor"] = defaultColor(args["fontColor"], map[string]any{"r": 0, "g": 0, "b": 0, "a": 1})
		payload["name"] = nameOr(args["name"], "Text")
		return s.figma.SendCommand(name, payload, 30*time.Second)

	case "set_fill_color":
		return s.figma.SendCommand(name, map[string]any{
			"nodeId": args["nodeId"],
			"color": map[string]any{
				"r": args["r"], "g": args["g"], "b": args["b"], "a": numberOrDefault(args["a"], 1),
			},
		}, 30*time.Second)

	case "set_stroke_color":
		return s.figma.SendCommand(name, map[string]any{
			"nodeId": args["nodeId"],
			"color": map[string]any{
				"r": args["r"], "g": args["g"], "b": args["b"], "a": numberOrDefault(args["a"], 1),
			},
			"weight": numberOrDefault(args["weight"], 1),
		}, 30*time.Second)

	case "set_corner_radius":
		payload := copyArgs(args)
		if _, ok := payload["corners"]; !ok {
			payload["corners"] = []any{true, true, true, true}
		}
		return s.figma.SendCommand(name, payload, 30*time.Second)

	case "set_layout_mode":
		payload := copyArgs(args)
		if _, ok := payload["layoutWrap"]; !ok {
			payload["layoutWrap"] = "NO_WRAP"
		}
		return s.figma.SendCommand(name, payload, 30*time.Second)

	case "export_node_as_image":
		payload := copyArgs(args)
		if _, ok := payload["format"]; !ok {
			payload["format"] = "PNG"
		}
		payload["scale"] = numberOrDefault(args["scale"], 1)
		return s.figma.SendCommand(name, payload, 60*time.Second)

	case "create_image":
		return s.figma.SendCommand(name, args, 120*time.Second)

	case "get_variables":
		return s.figma.SendCommand(name, args, 60*time.Second)

	case "get_instance_overrides":
		return s.figma.SendCommand(name, map[string]any{
			"instanceNodeId": args["nodeId"],
		}, 30*time.Second)

	case "set_instance_overrides":
		payload := map[string]any{
			"sourceInstanceId": args["sourceInstanceId"],
			"targetNodeIds":    args["targetNodeIds"],
		}
		if payload["targetNodeIds"] == nil {
			payload["targetNodeIds"] = []any{}
		}
		return s.figma.SendCommand(name, payload, 60*time.Second)

	case "scan_text_nodes":
		payload := copyArgs(args)
		payload["useChunking"] = true
		payload["chunkSize"] = 10
		return s.figma.SendCommand(name, payload, 5*time.Minute)

	case "set_multiple_text_contents", "set_multiple_annotations", "scan_nodes_by_types":
		return s.figma.SendCommand(name, args, 5*time.Minute)

	case "get_annotations":
		payload := copyArgs(args)
		if _, ok := payload["includeCategories"]; !ok {
			payload["includeCategories"] = true
		}
		return s.figma.SendCommand(name, payload, 30*time.Second)

	default:
		return s.figma.SendCommand(name, args, 30*time.Second)
	}
}

// getNodesInfo fans the tool out to N parallel get_node_info calls so the
// plugin can stream node details one at a time, matching the TS server which
// uses Promise.all over get_node_info.
func (s *Server) getNodesInfo(args map[string]any) (any, error) {
	rawIDs, _ := args["nodeIds"].([]any)
	if len(rawIDs) == 0 {
		return []any{}, nil
	}

	type indexed struct {
		idx    int
		result any
		err    error
	}
	out := make(chan indexed, len(rawIDs))
	for i, id := range rawIDs {
		go func(i int, id any) {
			res, err := s.figma.SendCommand("get_node_info", map[string]any{"nodeId": id}, 30*time.Second)
			out <- indexed{idx: i, result: res, err: err}
		}(i, id)
	}

	results := make([]any, len(rawIDs))
	for i := 0; i < len(rawIDs); i++ {
		item := <-out
		if item.err != nil {
			return nil, item.err
		}
		results[item.idx] = item.result
	}
	return results, nil
}

func (s *Server) hasTool(name string) bool {
	for _, t := range s.tools {
		if t.Name == name {
			return true
		}
	}
	return false
}

func validateRequired(name string, args map[string]any, all []tool) error {
	for _, t := range all {
		if t.Name != name {
			continue
		}
		var required []string
		if values, ok := t.InputSchema["required"].([]string); ok {
			required = values
		} else if values, ok := t.InputSchema["required"].([]any); ok {
			for _, value := range values {
				required = append(required, fmt.Sprint(value))
			}
		}
		for _, key := range required {
			if _, ok := args[key]; !ok {
				return fmt.Errorf("missing required argument %q for %s", key, name)
			}
		}
		return nil
	}
	return nil
}

func getPrompt(raw json.RawMessage) (any, error) {
	var params struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, err
	}
	for _, p := range prompts() {
		if p.Name == params.Name {
			return map[string]any{
				"description": p.Description,
				"messages": []map[string]any{
					{
						"role": "assistant",
						"content": map[string]any{
							"type": "text",
							"text": p.Text,
						},
					},
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("unknown prompt: %s", params.Name)
}

func copyArgs(args map[string]any) map[string]any {
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = v
	}
	return out
}

func errorResult(err error) toolCallResult {
	return toolCallResult{Content: []contentItem{{Type: "text", Text: err.Error()}}, IsError: true}
}
