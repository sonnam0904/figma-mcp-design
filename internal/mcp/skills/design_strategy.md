When working with Figma designs, follow these best practices:

1. Start with Document Structure:
   - First use get_document_info() to understand the current document
   - Plan your layout hierarchy before creating elements
   - Create a main container frame for each screen/section

2. Naming Conventions:
   - Use descriptive, semantic names for all elements
   - Follow a consistent naming pattern (e.g., "Login Screen", "Logo Container", "Email Input")
   - Group related elements with meaningful names

3. Layout Hierarchy:
   - Create parent frames first, then add child elements
   - For forms/login screens:
     * Start with the main screen container frame
     * Create a logo container at the top
     * Group input fields in their own containers
     * Place action buttons (login, submit) after inputs
     * Add secondary elements (forgot password, signup links) last

4. Input Fields Structure:
   - Create a container frame for each input field
   - Include a label text above or inside the input
   - Group related inputs (e.g., username/password) together

5. Element Creation:
   - Use create_frame() for containers and input fields
   - Use create_text() for labels, buttons text, and links
   - Use create_image() when you have PNG or JPEG bytes as base64 (e.g. logos); set parentId to keep hierarchy; use scaleMode FIT or FILL as needed
   - For design-system colors and typography from the same file: call get_styles(), then apply_style({ nodeId, styleId, styleType }) with styleType TEXT for text nodes (ids from texts) or FILL for fills (ids from colors)
   - For **Figma Variables** (tokens): call get_variables(), then set_variable_binding({ nodeId, field, variableId }) — use field fillPaintColor / strokePaintColor for COLOR on solid paints, or node fields such as width, itemSpacing, paddingTop for FLOAT variables. Follow the variable_strategy prompt for details.
   - Set appropriate raw colors when no matching local style exists:
     * Use fillColor for backgrounds
     * Use strokeColor for borders
     * Set proper fontWeight for different text elements

6. Modifying existing elements:
   - Use set_text_content() to modify text content.

7. Visual Hierarchy:
   - Position elements in logical reading order (top to bottom)
   - Maintain consistent spacing between elements
   - Use appropriate font sizes for different text types:
     * Larger for headings/welcome text
     * Medium for input labels
     * Standard for button text
     * Smaller for helper text/links

8. Best Practices:
   - Verify each creation with get_node_info()
   - Use parentId to maintain proper hierarchy
   - Group related elements together in frames
   - Keep consistent spacing and alignment

Example Login Screen Structure:
- Login Screen (main frame)
  - Logo Container (frame)
    - Logo: use create_image() with imageBase64 (PNG/JPEG) when you have file bytes; otherwise create_text() for a wordmark or create_frame() as a named placeholder
  - Welcome Text (text)
  - Input Container (frame)
    - Email Input (frame)
      - Email Label (text)
      - Email Field (frame)
    - Password Input (frame)
      - Password Label (text)
      - Password Field (frame)
  - Login Button (frame)
    - Button Text (text)
  - Helper Links (frame)
    - Forgot Password (text)
    - Don't have account (text)

9. Current tool limitations (do not assume capabilities beyond the exposed MCP tools):
   - create_image accepts base64 PNG/JPEG only; very large payloads may fail (the plugin enforces a decoded size cap). There is no URL fetch or SVG-as-image import in this tool.
   - apply_style only maps to local document **styles** from get_styles. Variables use get_variables + set_variable_binding instead.
   - set_variable_binding supports local variables and common fields; library-variable import, extended enterprise collections, and arbitrary vector/boolean or full prototype authoring are not covered end-to-end here.
   - Prefer frames, rectangles, text, component instances (from get_local_components), styles, variables, and auto-layout for structure.