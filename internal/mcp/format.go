package mcp

import (
	"fmt"
	"math"
	"strings"
)

// rgbaToHex converts a Figma RGBA color (0-1 floats) to a hex string.
// When the alpha channel is fully opaque the alpha component is omitted.
func rgbaToHex(color any) string {
	if s, ok := color.(string); ok {
		if strings.HasPrefix(s, "#") {
			return s
		}
	}
	m, ok := color.(map[string]any)
	if !ok {
		return ""
	}
	r := channelToByte(m["r"])
	g := channelToByte(m["g"])
	b := channelToByte(m["b"])
	a := 255
	if v, present := m["a"]; present {
		a = channelToByte(v)
	}
	if a == 255 {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

func channelToByte(v any) int {
	switch n := v.(type) {
	case float64:
		return clampInt(int(math.Round(n*255)), 0, 255)
	case float32:
		return clampInt(int(math.Round(float64(n)*255)), 0, 255)
	case int:
		if n <= 1 {
			return clampInt(int(math.Round(float64(n)*255)), 0, 255)
		}
		return clampInt(n, 0, 255)
	case int64:
		return channelToByte(float64(n))
	}
	return 0
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// filterFigmaNode trims raw node payloads coming back from the plugin so that
// model output stays compact: VECTOR nodes are removed, fills/strokes lose
// boundVariables/imageRef, gradient stops and solid colors are converted to
// hex, and only a small set of style fields is kept on TEXT nodes.
func filterFigmaNode(node any) any {
	m, ok := node.(map[string]any)
	if !ok {
		return node
	}
	if t, _ := m["type"].(string); t == "VECTOR" {
		return nil
	}

	filtered := map[string]any{
		"id":   m["id"],
		"name": m["name"],
		"type": m["type"],
	}

	if fills, ok := m["fills"].([]any); ok && len(fills) > 0 {
		filtered["fills"] = processPaintArray(fills)
	}
	if strokes, ok := m["strokes"].([]any); ok && len(strokes) > 0 {
		filtered["strokes"] = processPaintArray(strokes)
	}
	if v, ok := m["cornerRadius"]; ok {
		filtered["cornerRadius"] = v
	}
	if v, ok := m["absoluteBoundingBox"]; ok {
		filtered["absoluteBoundingBox"] = v
	}
	if v, ok := m["characters"]; ok {
		filtered["characters"] = v
	}
	if style, ok := m["style"].(map[string]any); ok {
		filtered["style"] = map[string]any{
			"fontFamily":          style["fontFamily"],
			"fontStyle":           style["fontStyle"],
			"fontWeight":          style["fontWeight"],
			"fontSize":            style["fontSize"],
			"textAlignHorizontal": style["textAlignHorizontal"],
			"letterSpacing":       style["letterSpacing"],
			"lineHeightPx":        style["lineHeightPx"],
		}
	}
	if children, ok := m["children"].([]any); ok {
		out := make([]any, 0, len(children))
		for _, child := range children {
			if c := filterFigmaNode(child); c != nil {
				out = append(out, c)
			}
		}
		filtered["children"] = out
	}
	return filtered
}

func processPaintArray(items []any) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		paint, ok := item.(map[string]any)
		if !ok {
			out = append(out, item)
			continue
		}
		copyMap := make(map[string]any, len(paint))
		for k, v := range paint {
			if k == "boundVariables" || k == "imageRef" {
				continue
			}
			copyMap[k] = v
		}
		if stops, ok := copyMap["gradientStops"].([]any); ok {
			processed := make([]any, 0, len(stops))
			for _, stop := range stops {
				sm, ok := stop.(map[string]any)
				if !ok {
					processed = append(processed, stop)
					continue
				}
				cp := make(map[string]any, len(sm))
				for k, v := range sm {
					if k == "boundVariables" {
						continue
					}
					cp[k] = v
				}
				if c, present := cp["color"]; present {
					cp["color"] = rgbaToHex(c)
				}
				processed = append(processed, cp)
			}
			copyMap["gradientStops"] = processed
		}
		if c, present := copyMap["color"]; present {
			copyMap["color"] = rgbaToHex(c)
		}
		out = append(out, copyMap)
	}
	return out
}

// asMap is a convenience accessor for plugin responses that come back as
// untyped JSON objects.
func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func asBool(v any, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
}

func numberOrDefault(v any, fallback float64) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return fallback
}

func nameOr(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func defaultColor(v any, fallback map[string]any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return fallback
}

func formatRGBA(r, g, b, a any) string {
	return fmt.Sprintf("RGBA(%v, %v, %v, %v)", r, g, b, numberOrDefault(a, 1))
}
