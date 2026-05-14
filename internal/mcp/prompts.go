package mcp

import _ "embed"

// The prompts below are loaded verbatim from internal/mcp/skills/*.md
// at build time via //go:embed, so the canonical source of each instruction
// prompt is its Markdown file.

//go:embed skills/design_strategy.md
var designStrategyText string

//go:embed skills/read_design_strategy.md
var readDesignStrategyText string

//go:embed skills/text_replacement_strategy.md
var textReplacementStrategyText string

//go:embed skills/annotation_conversion_strategy.md
var annotationConversionStrategyText string

//go:embed skills/swap_overrides_instances.md
var swapOverridesInstancesText string

//go:embed skills/reaction_to_connector_strategy.md
var reactionToConnectorStrategyText string

//go:embed skills/variable_strategy.md
var variableStrategyText string

func prompts() []prompt {
	return []prompt{
		{
			Name:        "design_strategy",
			Description: "Best practices for working with Figma designs",
			Text:        designStrategyText,
		},
		{
			Name:        "read_design_strategy",
			Description: "Best practices for reading Figma designs",
			Text:        readDesignStrategyText,
		},
		{
			Name:        "text_replacement_strategy",
			Description: "Systematic approach for replacing text in Figma designs",
			Text:        textReplacementStrategyText,
		},
		{
			Name:        "annotation_conversion_strategy",
			Description: "Strategy for converting manual annotations to Figma's native annotations",
			Text:        annotationConversionStrategyText,
		},
		{
			Name:        "swap_overrides_instances",
			Description: "Strategy for transferring overrides between component instances in Figma",
			Text:        swapOverridesInstancesText,
		},
		{
			Name:        "reaction_to_connector_strategy",
			Description: "Strategy for converting Figma prototype reactions to connector lines using the output of 'get_reactions'",
			Text:        reactionToConnectorStrategyText,
		},
		{
			Name:        "variable_strategy",
			Description: "How to read local Figma variables and bind them to nodes with get_variables and set_variable_binding",
			Text:        variableStrategyText,
		},
	}
}
