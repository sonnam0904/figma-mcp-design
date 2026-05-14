# Swap Component Instance and Override Strategy

## Overview
This strategy enables transferring content and property overrides from a source instance to one or more target instances in Figma, maintaining design consistency while reducing manual work.

## Step-by-Step Process

### 1. Selection Analysis
- Use `get_selection()` to identify the parent component or selected instances
- For parent components, scan for instances with `scan_nodes_by_types({ nodeId: "parent-id", types: ["INSTANCE"] })`
- Identify custom slots by name patterns (e.g. "Custom Slot*" or "Instance Slot") or by examining text content
- Determine which is the source instance (with content to copy) and which are targets (where to apply content)

### 2. Extract Source Overrides
- Use `get_instance_overrides()` to extract customizations from the source instance
- This captures text content, property values, and style overrides
- Command syntax: `get_instance_overrides({ nodeId: "source-instance-id" })`
- Look for successful response like "Got component information from [instance name]"

### 3. Apply Overrides to Targets
- Apply captured overrides using `set_instance_overrides()`
- Command syntax:
  ```
  set_instance_overrides({
    sourceInstanceId: "source-instance-id", 
    targetNodeIds: ["target-id-1", "target-id-2", ...]
  })
  ```

### 4. Verification
- Verify results with `get_node_info()` or `read_my_design()`
- Confirm text content and style overrides have transferred successfully

## Key Tips
- Always join the appropriate channel first with `join_channel()`
- When working with multiple targets, check the full selection with `get_selection()`
- Preserve component relationships by using instance overrides rather than direct text manipulation