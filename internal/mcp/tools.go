package mcp

func tools() []tool {
	noArgs := objectSchema(nil, map[string]any{})
	nodeID := map[string]any{"nodeId": prop("string", "Figma node ID")}
	nodeIDs := map[string]any{"nodeIds": arrayOf("Array of Figma node IDs", prop("string", "Figma node ID"))}

	textReplacement := objectSchema([]string{"nodeId", "text"}, map[string]any{
		"nodeId": prop("string", "The ID of the text node"),
		"text":   prop("string", "Replacement text"),
	})

	annotationItem := objectSchema([]string{"nodeId", "labelMarkdown"}, map[string]any{
		"nodeId":        prop("string", "The ID of the node to annotate"),
		"labelMarkdown": prop("string", "The annotation text in markdown format"),
		"categoryId":    prop("string", "The ID of the annotation category"),
		"annotationId":  prop("string", "The ID of the annotation to update (if updating existing annotation)"),
		"properties": arrayOf("Additional properties for the annotation", objectSchema([]string{"type"}, map[string]any{
			"type": prop("string", "Property type"),
		})),
	})

	connectionItem := objectSchema([]string{"startNodeId", "endNodeId"}, map[string]any{
		"startNodeId": prop("string", "ID of the starting node"),
		"endNodeId":   prop("string", "ID of the ending node"),
		"text":        prop("string", "Optional text to display on the connector"),
	})

	rgbaColor := func(description string) map[string]any {
		return objectSchema([]string{"r", "g", "b"}, map[string]any{
			"r": prop("number", "Red component (0-1)"),
			"g": prop("number", "Green component (0-1)"),
			"b": prop("number", "Blue component (0-1)"),
			"a": prop("number", "Alpha component (0-1)"),
		})
	}

	return []tool{
		{"get_document_info", "Get detailed information about the current Figma document", noArgs},
		{"get_selection", "Get information about the current selection in Figma (includes current page and document ids, selection count, and per-node id, name, type, visibility, and size/position when available)", noArgs},
		{"read_my_design", "Get detailed information about the current selection in Figma, including all node details", noArgs},
		{"get_node_info", "Get detailed information about a specific node in Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to get information about"),
		})},
		{"get_nodes_info", "Get detailed information about multiple nodes in Figma", objectSchema([]string{"nodeIds"}, map[string]any{
			"nodeIds": arrayOf("Array of node IDs to get information about", prop("string", "Figma node ID")),
		})},

		{"create_rectangle", "Create a new rectangle in Figma", objectSchema([]string{"x", "y", "width", "height"}, map[string]any{
			"x":        prop("number", "X position"),
			"y":        prop("number", "Y position"),
			"width":    prop("number", "Width of the rectangle"),
			"height":   prop("number", "Height of the rectangle"),
			"name":     prop("string", "Optional name for the rectangle"),
			"parentId": prop("string", "Optional parent node ID to append the rectangle to"),
		})},

		{"create_frame", "Create a new frame in Figma", objectSchema([]string{"x", "y", "width", "height"}, map[string]any{
			"x":            prop("number", "X position"),
			"y":            prop("number", "Y position"),
			"width":        prop("number", "Width of the frame"),
			"height":       prop("number", "Height of the frame"),
			"name":         prop("string", "Optional name for the frame"),
			"parentId":     prop("string", "Optional parent node ID to append the frame to"),
			"fillColor":    rgbaColor("Fill color in RGBA format"),
			"strokeColor":  rgbaColor("Stroke color in RGBA format"),
			"strokeWeight": prop("number", "Stroke weight"),
			"layoutMode":   stringEnum("Auto-layout mode for the frame", "NONE", "HORIZONTAL", "VERTICAL"),
			"layoutWrap":   stringEnum("Whether the auto-layout frame wraps its children", "NO_WRAP", "WRAP"),
			"paddingTop":   prop("number", "Top padding for auto-layout frame"),
			"paddingRight": prop("number", "Right padding for auto-layout frame"),
			"paddingBottom": prop("number", "Bottom padding for auto-layout frame"),
			"paddingLeft":   prop("number", "Left padding for auto-layout frame"),
			"primaryAxisAlignItems":  stringEnum("Primary axis alignment. SPACE_BETWEEN makes itemSpacing inert.", "MIN", "MAX", "CENTER", "SPACE_BETWEEN"),
			"counterAxisAlignItems":  stringEnum("Counter axis alignment", "MIN", "MAX", "CENTER", "BASELINE"),
			"layoutSizingHorizontal": stringEnum("Horizontal sizing mode", "FIXED", "HUG", "FILL"),
			"layoutSizingVertical":   stringEnum("Vertical sizing mode", "FIXED", "HUG", "FILL"),
			"itemSpacing":            prop("number", "Distance between children. Ignored when primaryAxisAlignItems is SPACE_BETWEEN."),
		})},

		{"create_text", "Create a new text element in Figma", objectSchema([]string{"x", "y", "text"}, map[string]any{
			"x":          prop("number", "X position"),
			"y":          prop("number", "Y position"),
			"text":       prop("string", "Text content"),
			"fontSize":   prop("number", "Font size (default: 14)"),
			"fontWeight": prop("number", "Font weight (e.g., 400 for Regular, 700 for Bold)"),
			"fontColor":  rgbaColor("Font color in RGBA format"),
			"name":       prop("string", "Semantic layer name for the text node"),
			"parentId":   prop("string", "Optional parent node ID to append the text to"),
		})},

		{"create_image", "Create a rectangle with a raster image fill from base64-encoded PNG or JPEG bytes (max decoded size 15MB in the plugin). Use for logos and bitmap assets.", objectSchema([]string{"x", "y", "width", "height", "imageBase64"}, map[string]any{
			"x":           prop("number", "X position"),
			"y":           prop("number", "Y position"),
			"width":       prop("number", "Width of the image bounds"),
			"height":      prop("number", "Height of the image bounds"),
			"imageBase64": prop("string", "Base64-encoded PNG or JPEG file bytes; optional data URL prefix (e.g. data:image/png;base64,) is stripped"),
			"name":        prop("string", "Optional layer name (default: Image)"),
			"parentId":    prop("string", "Optional parent node ID to append the rectangle to"),
			"scaleMode":   stringEnum("Image scale inside the rectangle", "FIT", "FILL", "CROP", "TILE"),
		})},

		{"set_fill_color", "Set the fill color of a node in Figma. Works on TextNode or FrameNode and similar shape nodes.", objectSchema([]string{"nodeId", "r", "g", "b"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to modify"),
			"r":      prop("number", "Red component (0-1)"),
			"g":      prop("number", "Green component (0-1)"),
			"b":      prop("number", "Blue component (0-1)"),
			"a":      prop("number", "Alpha component (0-1)"),
		})},
		{"set_stroke_color", "Set the stroke color of a node in Figma", objectSchema([]string{"nodeId", "r", "g", "b"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to modify"),
			"r":      prop("number", "Red component (0-1)"),
			"g":      prop("number", "Green component (0-1)"),
			"b":      prop("number", "Blue component (0-1)"),
			"a":      prop("number", "Alpha component (0-1)"),
			"weight": prop("number", "Stroke weight"),
		})},

		{"move_node", "Move a node to a new position in Figma", objectSchema([]string{"nodeId", "x", "y"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to move"),
			"x":      prop("number", "New X position"),
			"y":      prop("number", "New Y position"),
		})},
		{"clone_node", "Clone an existing node in Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to clone"),
			"x":      prop("number", "New X position for the clone"),
			"y":      prop("number", "New Y position for the clone"),
		})},
		{"resize_node", "Resize a node in Figma", objectSchema([]string{"nodeId", "width", "height"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to resize"),
			"width":  prop("number", "New width"),
			"height": prop("number", "New height"),
		})},
		{"delete_node", "Delete a node from Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to delete"),
		})},
		{"delete_multiple_nodes", "Delete multiple nodes from Figma at once", objectSchema([]string{"nodeIds"}, map[string]any{
			"nodeIds": arrayOf("Array of node IDs to delete", prop("string", "Figma node ID")),
		})},

		{"export_node_as_image", "Export a node as an image from Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to export"),
			"format": stringEnum("Export format", "PNG", "JPG", "SVG", "PDF"),
			"scale":  prop("number", "Export scale"),
		})},

		{"set_text_content", "Set the text content of an existing text node in Figma", objectSchema([]string{"nodeId", "text"}, map[string]any{
			"nodeId": prop("string", "The ID of the text node to modify"),
			"text":   prop("string", "New text content"),
		})},

		{"get_styles", "Get all styles from the current Figma document", noArgs},
		{"apply_style", "Apply a local document style to a node by id. Use ids from get_styles: styleType FILL=colors (paint/fill style), STROKE=colors as stroke style where supported, TEXT=texts, EFFECT=effects, GRID=grids.", objectSchema([]string{"nodeId", "styleId", "styleType"}, map[string]any{
			"nodeId":    prop("string", "Target node id"),
			"styleId":   prop("string", "Local style id from get_styles (colors[].id, texts[].id, effects[].id, or grids[].id)"),
			"styleType": stringEnum("Which style slot to set on the node", "FILL", "PAINT", "STROKE", "TEXT", "EFFECT", "GRID"),
		})},
		{"get_variables", "List local Figma variable collections and variables in the current file (modes, ids, resolvedType, valuesByMode). Use before set_variable_binding.", objectSchema(nil, map[string]any{
			"resolvedType": stringEnum("Optional: only return variables of this resolved type", "COLOR", "FLOAT", "STRING", "BOOLEAN", "ALIAS"),
		})},
		{"set_variable_binding", "Bind a Variable to a node property, or unbind. For COLOR on fills/strokes use field fillPaintColor or strokePaintColor (solid paint at fillIndex/strokeIndex). Otherwise use a setBoundVariable field name supported by the node (e.g. width, height, cornerRadius, opacity, itemSpacing, paddingTop, strokeWeight, fontFamily, fontWeight, lineHeight, letterSpacing on TEXT).", objectSchema([]string{"nodeId", "field"}, map[string]any{
			"nodeId":      prop("string", "Target node id"),
			"field":       prop("string", "fillPaintColor | strokePaintColor | or a VariableBindable field name for this node type"),
			"variableId":  prop("string", "Variable id from get_variables (required unless unbind is true)"),
			"fillIndex":   prop("number", "Fill index when field is fillPaintColor (default 0)"),
			"strokeIndex": prop("number", "Stroke index when field is strokePaintColor (default 0)"),
			"unbind":      prop("boolean", "If true, removes the binding (variableId ignored)"),
		})},
		{"get_local_components", "Get all local components from the Figma document", noArgs},

		{"get_annotations", "Get all annotations in the current document or specific node", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId":            prop("string", "Node ID to get annotations for specific node"),
			"includeCategories": prop("boolean", "Whether to include category information (default true)"),
		})},
		{"set_annotation", "Create or update an annotation", annotationItem},
		{"set_multiple_annotations", "Set multiple annotations parallelly in a node", objectSchema([]string{"nodeId", "annotations"}, map[string]any{
			"nodeId":      prop("string", "The ID of the node containing the elements to annotate"),
			"annotations": arrayOf("Array of annotations to apply", annotationItem),
		})},

		{"create_component_instance", "Create an instance of a component in Figma. For LOCAL components (from get_local_components), use componentId with the id field. For published LIBRARY components, use componentKey with the publishedKey field.", objectSchema([]string{"x", "y"}, map[string]any{
			"componentId":  prop("string", "ID of a local component. Use the id field from get_local_components result."),
			"componentKey": prop("string", "Key of a published library component (publishedKey from get_local_components)."),
			"x":            prop("number", "X position"),
			"y":            prop("number", "Y position"),
			"parentId":     prop("string", "Optional parent node ID to place the instance into"),
		})},
		{"get_instance_overrides", "Get all override properties from a selected component instance. These overrides can be applied to other instances, which will swap them to match the source component.", objectSchema(nil, map[string]any{
			"nodeId": prop("string", "Optional ID of the component instance to get overrides from. If not provided, currently selected instance will be used."),
		})},
		{"set_instance_overrides", "Apply previously copied overrides to selected component instances. Target instances will be swapped to the source component and all copied override properties will be applied.", objectSchema([]string{"sourceInstanceId", "targetNodeIds"}, map[string]any{
			"sourceInstanceId": prop("string", "ID of the source component instance"),
			"targetNodeIds":    arrayOf("Array of target instance IDs", prop("string", "Target instance ID")),
		})},

		{"set_corner_radius", "Set the corner radius of a node in Figma", objectSchema([]string{"nodeId", "radius"}, map[string]any{
			"nodeId":  prop("string", "The ID of the node to modify"),
			"radius":  prop("number", "Corner radius value"),
			"corners": arrayOf("Optional array of 4 booleans to specify which corners to round [topLeft, topRight, bottomRight, bottomLeft]", prop("boolean", "Whether to round this corner")),
		})},

		{"scan_text_nodes", "Scan all text nodes in the selected Figma node", objectSchema([]string{"nodeId"}, nodeID)},
		{"set_multiple_text_contents", "Set multiple text contents parallelly in a node", objectSchema([]string{"nodeId", "text"}, map[string]any{
			"nodeId": prop("string", "The ID of the node containing the text nodes to replace"),
			"text":   arrayOf("Array of text node IDs and their replacement texts", textReplacement),
		})},

		{"scan_nodes_by_types", "Scan for child nodes with specific types in the selected Figma node", objectSchema([]string{"nodeId", "types"}, map[string]any{
			"nodeId": prop("string", "ID of the node to scan"),
			"types":  arrayOf("Array of node types to find (e.g. ['COMPONENT', 'FRAME'])", prop("string", "Figma node type")),
		})},

		{"set_layout_mode", "Set the layout mode and wrap behavior of a frame in Figma", objectSchema([]string{"nodeId", "layoutMode"}, map[string]any{
			"nodeId":     prop("string", "The ID of the frame to modify"),
			"layoutMode": stringEnum("Layout mode for the frame", "NONE", "HORIZONTAL", "VERTICAL"),
			"layoutWrap": stringEnum("Whether the auto-layout frame wraps its children", "NO_WRAP", "WRAP"),
		})},
		{"set_padding", "Set padding values for an auto-layout frame in Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId":        prop("string", "The ID of the frame to modify"),
			"paddingTop":    prop("number", "Top padding value"),
			"paddingRight":  prop("number", "Right padding value"),
			"paddingBottom": prop("number", "Bottom padding value"),
			"paddingLeft":   prop("number", "Left padding value"),
		})},
		{"set_axis_align", "Set primary and counter axis alignment for an auto-layout frame in Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId":                prop("string", "The ID of the frame to modify"),
			"primaryAxisAlignItems": stringEnum("Primary axis alignment (MIN/MAX = left/right in horizontal, top/bottom in vertical). SPACE_BETWEEN ignores itemSpacing.", "MIN", "MAX", "CENTER", "SPACE_BETWEEN"),
			"counterAxisAlignItems": stringEnum("Counter axis alignment", "MIN", "MAX", "CENTER", "BASELINE"),
		})},
		{"set_layout_sizing", "Set horizontal and vertical sizing modes for an auto-layout frame in Figma", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId":                 prop("string", "The ID of the frame to modify"),
			"layoutSizingHorizontal": stringEnum("Horizontal sizing mode (HUG for frames/text only, FILL for auto-layout children only)", "FIXED", "HUG", "FILL"),
			"layoutSizingVertical":   stringEnum("Vertical sizing mode (HUG for frames/text only, FILL for auto-layout children only)", "FIXED", "HUG", "FILL"),
		})},
		{"set_item_spacing", "Set distance between children in an auto-layout frame", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId":             prop("string", "The ID of the frame to modify"),
			"itemSpacing":        prop("number", "Distance between children. Ignored when primaryAxisAlignItems is SPACE_BETWEEN."),
			"counterAxisSpacing": prop("number", "Distance between wrapped rows/columns. Only works when layoutWrap=WRAP."),
		})},

		{"get_reactions", "Get Figma Prototyping Reactions from multiple nodes. CRITICAL: The output MUST be processed using the 'reaction_to_connector_strategy' prompt IMMEDIATELY to generate parameters for connector lines via the 'create_connections' tool.", objectSchema([]string{"nodeIds"}, nodeIDs)},
		{"set_default_connector", "Set a copied connector node as the default connector", objectSchema(nil, map[string]any{
			"connectorId": prop("string", "The ID of the connector node to set as default"),
		})},
		{"create_connections", "Create connections between nodes using the default connector style", objectSchema([]string{"connections"}, map[string]any{
			"connections": arrayOf("Array of node connections to create", connectionItem),
		})},

		{"set_focus", "Set focus on a specific node in Figma by selecting it and scrolling viewport to it", objectSchema([]string{"nodeId"}, map[string]any{
			"nodeId": prop("string", "The ID of the node to focus on"),
		})},
		{"set_selections", "Set selection to multiple nodes in Figma and scroll viewport to show them", objectSchema([]string{"nodeIds"}, map[string]any{
			"nodeIds": arrayOf("Array of node IDs to select", prop("string", "Node ID")),
		})},

		{"join_channel", "Join a specific channel to communicate with Figma", objectSchema([]string{"channel"}, map[string]any{
			"channel": prop("string", "The name of the channel to join"),
		})},
	}
}
