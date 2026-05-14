# Strategy: Convert Figma Prototype Reactions to Connector Lines

## Goal
Process the JSON output from the `get_reactions` tool to generate an array of connection objects suitable for the `create_connections` tool. This visually represents prototype flows as connector lines on the Figma canvas.

## Input Data
You will receive JSON data from the `get_reactions` tool. This data contains an array of nodes, each with potential reactions. A typical reaction object looks like this:
```json
{
  "trigger": { "type": "ON_CLICK" },
  "action": {
    "type": "NAVIGATE",
    "destinationId": "destination-node-id",
    "navigationTransition": { ... },
    "preserveScrollPosition": false
  }
}
```

## Step-by-Step Process

### 1. Preparation & Context Gathering
   - **Action:** Call `read_my_design` on the relevant node(s) to get context about the nodes involved (names, types, etc.). This helps in generating meaningful connector labels later.
   - **Action:** Call `set_default_connector` **without** the `connectorId` parameter.
   - **Check Result:** Analyze the response from `set_default_connector`.
     - If it confirms a default connector is already set (e.g., "Default connector is already set"), proceed to Step 2.
     - If it indicates no default connector is set (e.g., "No default connector set..."), you **cannot** proceed with `create_connections` yet. Inform the user they need to manually copy a connector from FigJam, paste it onto the current page, select it, and then you can run `set_default_connector({ connectorId: "SELECTED_NODE_ID" })` before attempting `create_connections`. **Do not proceed to Step 2 until a default connector is confirmed.**

### 2. Filter and Transform Reactions from `get_reactions` Output
   - **Iterate:** Go through the JSON array provided by `get_reactions`. For each node in the array:
     - Iterate through its `reactions` array.
   - **Filter:** Keep only reactions where the `action` meets these criteria:
     - Has a `type` that implies a connection (e.g., `NAVIGATE`, `OPEN_OVERLAY`, `SWAP_OVERLAY`). **Ignore** types like `CHANGE_TO`, `CLOSE_OVERLAY`, etc.
     - Has a valid `destinationId` property.
   - **Extract:** For each valid reaction, extract the following information:
     - `sourceNodeId`: The ID of the node the reaction belongs to (from the outer loop).
     - `destinationNodeId`: The value of `action.destinationId`.
     - `actionType`: The value of `action.type`.
     - `triggerType`: The value of `trigger.type`.

### 3. Generate Connector Text Labels
   - **For each extracted connection:** Create a concise, descriptive text label string.
   - **Combine Information:** Use the `actionType`, `triggerType`, and potentially the names of the source/destination nodes (obtained from Step 1's `read_my_design` or by calling `get_node_info` if necessary) to generate the label.
   - **Example Labels:**
     - If `triggerType` is "ON\_CLICK" and `actionType` is "NAVIGATE": "On click, navigate to [Destination Node Name]"
     - If `triggerType` is "ON\_DRAG" and `actionType` is "OPEN\_OVERLAY": "On drag, open [Destination Node Name] overlay"
   - **Keep it brief and informative.** Let this generated string be `generatedText`.

### 4. Prepare the `connections` Array for `create_connections`
   - **Structure:** Create a JSON array where each element is an object representing a connection.
   - **Format:** Each object in the array must have the following structure:
     ```json
     {
       "startNodeId": "sourceNodeId_from_step_2",
       "endNodeId": "destinationNodeId_from_step_2",
       "text": "generatedText_from_step_3"
     }
     ```
   - **Result:** This final array is the value you will pass to the `connections` parameter when calling the `create_connections` tool.

### 5. Execute Connection Creation
   - **Action:** Call the `create_connections` tool, passing the array generated in Step 4 as the `connections` argument.
   - **Verify:** Check the response from `create_connections` to confirm success or failure.

This detailed process ensures you correctly interpret the reaction data, prepare the necessary information, and use the appropriate tools to create the connector lines.