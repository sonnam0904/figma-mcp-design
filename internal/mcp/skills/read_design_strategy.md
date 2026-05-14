When reading Figma designs, follow this order and these best practices:

1. Session and channel:
   - Call join_channel() first so the server talks to the correct Figma session.

2. Document and selection context:
   - Use get_document_info() to learn pages, structure, and overall context.
   - Use get_selection() to see what is selected; it includes currentPage, documentId/documentName, selectionCount, and per-node id, name, type, visible, plus width/height and x/y when the node exposes them.
   - When you need **design tokens** (variable names, modes, COLOR/FLOAT values): call get_variables() (optionally with resolvedType such as COLOR or FLOAT).

3. Deep read of nodes:
   - When there is a useful selection, use read_my_design() for rich detail on the current selection (preferred for “what am I looking at?”).
   - For a known node id (from selection, scan results, or the user), use get_node_info({ nodeId }).
   - For several known ids in parallel, use get_nodes_info({ nodeIds }).

4. When there is no selection or read_my_design is empty or unhelpful:
   - Ask the user to select one or more nodes in Figma, **or**
   - Ask the user to provide specific nodeId value(s) so you can call get_node_info / get_nodes_info without relying on selection.

5. Before changing the file:
   - Re-read get_selection or the target node after context gathering so edits apply to the intended layers.
