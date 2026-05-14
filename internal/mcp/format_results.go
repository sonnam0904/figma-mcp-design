package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatToolResult turns a raw plugin response into the MCP toolCallResult
// content blocks expected by clients.
// server's text output so model-side prompts continue to work unchanged.
func formatToolResult(name string, args map[string]any, result any) (toolCallResult, error) {
	switch name {
	case "join_channel":
		channel, _ := args["channel"].(string)
		return textResult(fmt.Sprintf("Successfully joined channel: %s", channel)), nil

	case "get_document_info", "get_selection", "read_my_design",
		"get_styles", "get_variables", "get_local_components", "get_annotations",
		"set_annotation", "create_component_instance",
		"delete_multiple_nodes":
		return jsonResult(result)

	case "get_node_info":
		return jsonResult(filterFigmaNode(result))

	case "get_nodes_info":
		items := asSlice(result)
		filtered := make([]any, 0, len(items))
		for _, item := range items {
			if v := filterFigmaNode(item); v != nil {
				filtered = append(filtered, v)
			}
		}
		return jsonResult(filtered)

	case "create_rectangle":
		raw, _ := json.Marshal(result)
		return textResult(fmt.Sprintf("Created rectangle \"%s\"", string(raw))), nil

	case "create_frame":
		m := asMap(result)
		return textResult(fmt.Sprintf("Created frame \"%s\" with ID: %s. Use the ID as the parentId to appendChild inside this frame.",
			asString(m["name"]), asString(m["id"]))), nil

	case "create_text":
		m := asMap(result)
		return textResult(fmt.Sprintf("Created text \"%s\" with ID: %s",
			asString(m["name"]), asString(m["id"]))), nil

	case "create_image":
		m := asMap(result)
		return textResult(fmt.Sprintf("Created image rectangle \"%s\" with ID: %s (imageHash: %s)",
			asString(m["name"]), asString(m["id"]), asString(m["imageHash"]))), nil

	case "apply_style":
		m := asMap(result)
		return textResult(fmt.Sprintf("Applied %s style %s to node \"%s\" (%s)",
			asString(m["styleType"]), asString(m["styleId"]), asString(m["name"]), asString(m["nodeId"]))), nil

	case "set_variable_binding":
		m := asMap(result)
		if !asBool(m["success"], false) {
			return jsonResult(result)
		}
		if asBool(m["unbind"], false) {
			return textResult(fmt.Sprintf("Unbound variable on field %s for node %s", asString(m["field"]), asString(m["nodeId"]))), nil
		}
		return textResult(fmt.Sprintf("Bound variable %s to field %s on node %s",
			asString(m["variableId"]), asString(m["field"]), asString(m["nodeId"]))), nil

	case "set_fill_color":
		m := asMap(result)
		return textResult(fmt.Sprintf("Set fill color of node \"%s\" to %s",
			asString(m["name"]), formatRGBA(args["r"], args["g"], args["b"], args["a"]))), nil

	case "set_stroke_color":
		m := asMap(result)
		weight := numberOrDefault(args["weight"], 1)
		return textResult(fmt.Sprintf("Set stroke color of node \"%s\" to %s with weight %g",
			asString(m["name"]), formatRGBA(args["r"], args["g"], args["b"], args["a"]), weight)), nil

	case "move_node":
		m := asMap(result)
		return textResult(fmt.Sprintf("Moved node \"%s\" to position (%v, %v)",
			asString(m["name"]), args["x"], args["y"])), nil

	case "clone_node":
		m := asMap(result)
		msg := fmt.Sprintf("Cloned node \"%s\" with new ID: %s", asString(m["name"]), asString(m["id"]))
		_, hasX := args["x"]
		_, hasY := args["y"]
		if hasX && hasY {
			msg += fmt.Sprintf(" at position (%v, %v)", args["x"], args["y"])
		}
		return textResult(msg), nil

	case "resize_node":
		m := asMap(result)
		return textResult(fmt.Sprintf("Resized node \"%s\" to width %v and height %v",
			asString(m["name"]), args["width"], args["height"])), nil

	case "delete_node":
		return textResult(fmt.Sprintf("Deleted node with ID: %v", args["nodeId"])), nil

	case "set_text_content":
		m := asMap(result)
		return textResult(fmt.Sprintf("Updated text content of node \"%s\" to \"%v\"",
			asString(m["name"]), args["text"])), nil

	case "set_corner_radius":
		m := asMap(result)
		return textResult(fmt.Sprintf("Set corner radius of node \"%s\" to %vpx",
			asString(m["name"]), args["radius"])), nil

	case "export_node_as_image":
		m := asMap(result)
		data := asString(m["imageData"])
		if data == "" {
			return jsonResult(result)
		}
		mime := asString(m["mimeType"])
		if mime == "" {
			mime = "image/png"
		}
		return toolCallResult{Content: []contentItem{{Type: "image", Data: data, MimeType: mime}}}, nil

	case "get_instance_overrides":
		m := asMap(result)
		success := asBool(m["success"], false)
		msg := asString(m["message"])
		if success {
			return textResult(fmt.Sprintf("Successfully got instance overrides: %s", msg)), nil
		}
		return textResult(fmt.Sprintf("Failed to get instance overrides: %s", msg)), nil

	case "set_instance_overrides":
		m := asMap(result)
		if !asBool(m["success"], false) {
			return textResult(fmt.Sprintf("Failed to set instance overrides: %s", asString(m["message"]))), nil
		}
		total := numberOrDefault(m["totalCount"], 0)
		successCount := 0
		for _, r := range asSlice(m["results"]) {
			if asBool(asMap(r)["success"], false) {
				successCount++
			}
		}
		return textResult(fmt.Sprintf("Successfully applied %g overrides to %d instances.", total, successCount)), nil

	case "scan_text_nodes":
		m := asMap(result)
		initial := contentItem{Type: "text", Text: "Starting text node scanning. This may take a moment for large designs..."}
		if _, hasChunks := m["chunks"]; hasChunks {
			summary := fmt.Sprintf("\n        Scan completed:\n        - Found %v text nodes\n        - Processed in %v chunks\n        ", m["totalNodes"], m["chunks"])
			text, _ := json.MarshalIndent(m["textNodes"], "", "  ")
			return toolCallResult{Content: []contentItem{
				initial,
				{Type: "text", Text: summary},
				{Type: "text", Text: string(text)},
			}}, nil
		}
		text, _ := json.MarshalIndent(result, "", "  ")
		return toolCallResult{Content: []contentItem{initial, {Type: "text", Text: string(text)}}}, nil

	case "scan_nodes_by_types":
		m := asMap(result)
		types, _ := args["types"].([]any)
		typeStrs := make([]string, 0, len(types))
		for _, t := range types {
			typeStrs = append(typeStrs, fmt.Sprint(t))
		}
		initial := contentItem{Type: "text", Text: fmt.Sprintf("Starting node type scanning for types: %s...", strings.Join(typeStrs, ", "))}
		if matching, ok := m["matchingNodes"]; ok {
			searched := m["searchedTypes"]
			summary := fmt.Sprintf("Scan completed: Found %v nodes matching types: %v", m["count"], searched)
			body, _ := json.MarshalIndent(matching, "", "  ")
			return toolCallResult{Content: []contentItem{
				initial,
				{Type: "text", Text: summary},
				{Type: "text", Text: string(body)},
			}}, nil
		}
		body, _ := json.MarshalIndent(result, "", "  ")
		return toolCallResult{Content: []contentItem{initial, {Type: "text", Text: string(body)}}}, nil

	case "set_multiple_text_contents":
		return formatBatchResult(result, args["text"],
			"Starting text replacement for %d nodes. This will be processed in batches of 5...",
			"replacementsApplied", "replacementsFailed",
			"Text replacement completed"), nil

	case "set_multiple_annotations":
		return formatBatchResult(result, args["annotations"],
			"Starting annotation process for %d nodes. This will be processed in batches of 5...",
			"annotationsApplied", "annotationsFailed",
			"Annotation process completed"), nil

	case "set_layout_mode":
		m := asMap(result)
		layoutMode, _ := args["layoutMode"].(string)
		wrapText := ""
		if w, ok := args["layoutWrap"].(string); ok && w != "" {
			wrapText = fmt.Sprintf(" with %s", w)
		}
		return textResult(fmt.Sprintf("Set layout mode of frame \"%s\" to %s%s",
			asString(m["name"]), layoutMode, wrapText)), nil

	case "set_padding":
		m := asMap(result)
		parts := []string{}
		for _, key := range []string{"paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
			if v, ok := args[key]; ok {
				parts = append(parts, fmt.Sprintf("%s: %v", strings.TrimPrefix(key, "padding"), v))
			}
		}
		text := "padding"
		if len(parts) > 0 {
			text = fmt.Sprintf("padding (%s)", strings.Join(parts, ", "))
		}
		return textResult(fmt.Sprintf("Set %s for frame \"%s\"", text, asString(m["name"]))), nil

	case "set_axis_align":
		m := asMap(result)
		parts := []string{}
		if v, ok := args["primaryAxisAlignItems"]; ok {
			parts = append(parts, fmt.Sprintf("primary: %v", v))
		}
		if v, ok := args["counterAxisAlignItems"]; ok {
			parts = append(parts, fmt.Sprintf("counter: %v", v))
		}
		text := "axis alignment"
		if len(parts) > 0 {
			text = fmt.Sprintf("axis alignment (%s)", strings.Join(parts, ", "))
		}
		return textResult(fmt.Sprintf("Set %s for frame \"%s\"", text, asString(m["name"]))), nil

	case "set_layout_sizing":
		m := asMap(result)
		parts := []string{}
		if v, ok := args["layoutSizingHorizontal"]; ok {
			parts = append(parts, fmt.Sprintf("horizontal: %v", v))
		}
		if v, ok := args["layoutSizingVertical"]; ok {
			parts = append(parts, fmt.Sprintf("vertical: %v", v))
		}
		text := "layout sizing"
		if len(parts) > 0 {
			text = fmt.Sprintf("layout sizing (%s)", strings.Join(parts, ", "))
		}
		return textResult(fmt.Sprintf("Set %s for frame \"%s\"", text, asString(m["name"]))), nil

	case "set_item_spacing":
		m := asMap(result)
		msg := fmt.Sprintf("Updated spacing for frame \"%s\":", asString(m["name"]))
		if v, ok := args["itemSpacing"]; ok {
			msg += fmt.Sprintf(" itemSpacing=%v", v)
		}
		if v, ok := args["counterAxisSpacing"]; ok {
			msg += fmt.Sprintf(" counterAxisSpacing=%v", v)
		}
		return textResult(msg), nil

	case "get_reactions":
		body, _ := json.Marshal(result)
		return toolCallResult{
			Content: []contentItem{
				{Type: "text", Text: string(body)},
				{Type: "text", Text: "IMPORTANT: You MUST now use the reaction data above and follow the `reaction_to_connector_strategy` prompt to prepare the parameters for the `create_connections` tool call. This is a required next step."},
			},
			FollowUp: map[string]any{
				"type":   "prompt",
				"prompt": "reaction_to_connector_strategy",
			},
		}, nil

	case "set_default_connector":
		body, _ := json.Marshal(result)
		return textResult(fmt.Sprintf("Default connector set: %s", string(body))), nil

	case "create_connections":
		conns, _ := args["connections"].([]any)
		body, _ := json.Marshal(result)
		return textResult(fmt.Sprintf("Created %d connections: %s", len(conns), string(body))), nil

	case "set_focus":
		m := asMap(result)
		return textResult(fmt.Sprintf("Focused on node \"%s\" (ID: %s)",
			asString(m["name"]), asString(m["id"]))), nil

	case "set_selections":
		m := asMap(result)
		nodes := asSlice(m["selectedNodes"])
		parts := make([]string, 0, len(nodes))
		for _, n := range nodes {
			nm := asMap(n)
			parts = append(parts, fmt.Sprintf("\"%s\" (%s)", asString(nm["name"]), asString(nm["id"])))
		}
		count := int(numberOrDefault(m["count"], float64(len(nodes))))
		return textResult(fmt.Sprintf("Selected %d nodes: %s", count, strings.Join(parts, ", "))), nil
	}

	return jsonResult(result)
}

func formatBatchResult(result any, items any, startMsg, appliedKey, failedKey, completedLabel string) toolCallResult {
	list, _ := items.([]any)
	initial := contentItem{Type: "text", Text: fmt.Sprintf(startMsg, len(list))}
	m := asMap(result)

	applied := int(numberOrDefault(m[appliedKey], 0))
	failed := int(numberOrDefault(m[failedKey], 0))
	chunks := int(numberOrDefault(m["completedInChunks"], 1))

	progress := fmt.Sprintf("\n      %s:\n      - %d of %d successfully %s\n      - %d failed\n      - Processed in %d batches\n      ",
		completedLabel, applied, len(list), pastTense(completedLabel), failed, chunks)

	detailed := ""
	for _, r := range asSlice(m["results"]) {
		entry := asMap(r)
		if asBool(entry["success"], false) {
			continue
		}
		errMsg := asString(entry["error"])
		if errMsg == "" {
			errMsg = "Unknown error"
		}
		detailed += fmt.Sprintf("\n- %s: %s", asString(entry["nodeId"]), errMsg)
	}
	if detailed != "" {
		detailed = "\n\nNodes that failed:" + detailed
	}

	return toolCallResult{Content: []contentItem{initial, {Type: "text", Text: progress + detailed}}}
}

func pastTense(label string) string {
	switch {
	case strings.Contains(label, "Text replacement"):
		return "updated"
	case strings.Contains(label, "Annotation"):
		return "applied"
	default:
		return "processed"
	}
}

func textResult(text string) toolCallResult {
	return toolCallResult{Content: []contentItem{{Type: "text", Text: text}}}
}

func jsonResult(value any) (toolCallResult, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return errorResult(err), err
	}
	return textResult(string(body)), nil
}
