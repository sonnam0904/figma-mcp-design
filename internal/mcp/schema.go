package mcp

func objectSchema(required []string, props map[string]any) map[string]any {
	s := map[string]any{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": true,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func prop(kind, description string) map[string]any {
	return map[string]any{"type": kind, "description": description}
}

func stringEnum(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": values}
}

func arrayOf(description string, items map[string]any) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": items}
}

func colorSchema(description string) map[string]any {
	return objectSchema([]string{"r", "g", "b"}, map[string]any{
		"r": prop("number", "Red component (0-1)"),
		"g": prop("number", "Green component (0-1)"),
		"b": prop("number", "Blue component (0-1)"),
		"a": prop("number", "Alpha component (0-1)"),
	})
}
