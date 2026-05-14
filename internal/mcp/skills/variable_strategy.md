# Figma Variables (read + bind)

Use this flow after `join_channel` when the file uses **Variables** (tokens) instead of only raw styles.

## 1. Discover tokens

```typescript
get_variables({}) // all collections + variables
// or narrow:
get_variables({ resolvedType: "COLOR" })
```

- Read `collections[]` for **modeId** / mode names (Light/Dark, etc.).
- Read `variables[]`: each entry has `id`, `name`, `resolvedType` (`COLOR`, `FLOAT`, `STRING`, `BOOLEAN`, `ALIAS`), `variableCollectionId`, and `valuesByMode` (per-mode values).

Pick a variable whose **resolvedType** matches what you want to bind (e.g. `COLOR` for fill color, `FLOAT` for width or spacing).

## 2. Bind to a node

Call `set_variable_binding` with:

- **nodeId**: target layer.
- **variableId**: from `get_variables.variables[].id` (omit when unbinding).
- **field**: how to apply the variable:

### Fill / stroke color (COLOR variables)

- Use **fillPaintColor** to bind to the **solid** fill at index `fillIndex` (default `0`). The fill must already be type `SOLID` (or the plugin will add a black solid at that index first).
- Use **strokePaintColor** similarly with `strokeIndex`.

```typescript
set_variable_binding({
  nodeId: "1:234",
  field: "fillPaintColor",
  variableId: "VariableID:...",
  fillIndex: 0,
})
```

### Layout / geometry / typography (direct fields)

Use any **setBoundVariable** field supported by that node type, for example:

- Frames / auto-layout: `width`, `height`, `itemSpacing`, `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft`, `cornerRadius`, `opacity`, `strokeWeight`, …
- Text: `fontFamily`, `fontStyle`, `fontWeight`, `lineHeight`, `letterSpacing`, … (load fonts first if changing family/style.)

```typescript
set_variable_binding({
  nodeId: "1:234",
  field: "itemSpacing",
  variableId: "VariableID:...",
})
```

The variable’s **resolvedType** must match the property (e.g. FLOAT for `itemSpacing`, COLOR for `fillPaintColor`).

## 3. Unbind

```typescript
set_variable_binding({
  nodeId: "1:234",
  field: "fillPaintColor",
  fillIndex: 0,
  unbind: true,
})
```

## 4. Verification

Use `get_node_info` / `read_my_design` and check `boundVariables` on the node when the payload includes them.

## Notes

- **Library variables**: binding usually requires variables **local to the file** or already linked; importing library variables may need user action in Figma.
- **Mixed / unsupported nodes**: if `setBoundVariable` is not supported on the node, the command fails with a clear error—use a compatible node type or `fillPaintColor` / `strokePaintColor` for paints.
- Prefer **get_variables** + **set_variable_binding** for tokenized spacing and colors; keep **get_styles** + **apply_style** for classic local **styles** (paint/text/effect/grid styles).
