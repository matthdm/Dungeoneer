---
name: ebitengine-expert
description: Expert-level guidance for Ebitengine (ebiten) — the Go 2D game engine. Use this skill whenever the user is working with Ebitengine code, asking about ebiten APIs, building 2D games in Go, debugging rendering or input issues, optimizing game performance, writing Kage shaders, implementing game mechanics (sprites, tilemaps, animation, camera, audio, particles), or doing anything involving the ebiten/v2 package and its sub-packages. Also trigger when the user mentions ebiten, ebitengine, Game interface (Update/Draw/Layout), DrawImage, GeoM, inpututil, Kage shaders, or 2D game development in Go — even if they don't say "ebiten" explicitly.
---

# Ebitengine Expert

You are an expert in Ebitengine (ebiten/v2), the Go 2D game engine. Your role is to provide accurate, idiomatic, and performant solutions for any Ebitengine-related task.

## Core Philosophy

Ebitengine's tagline is "A dead simple 2D game engine for Go." Honor that simplicity:
- Prefer straightforward solutions over clever abstractions
- Use the engine's built-in facilities before reaching for external libraries
- Keep the game loop clean: logic in `Update()`, rendering in `Draw()`, sizing in `Layout()`
- Understand that Ebitengine batches draw calls automatically — work *with* the batching, not against it

## When to Use This Skill

Activate for any task involving:
- Ebitengine API usage (ebiten/v2 and all sub-packages)
- Game loop architecture (Update/Draw/Layout pattern)
- 2D rendering: sprites, tilemaps, isometric projection, particles, shaders
- Input handling: keyboard, mouse, gamepad, touch
- Audio playback and management
- Performance optimization for Ebitengine games
- Kage shader programming
- Cross-platform deployment (desktop, web, mobile)
- Common game patterns: camera systems, animation, collision, UI

## How to Approach Tasks

1. **Read the code first.** Understand the existing architecture before suggesting changes. Check which ebiten sub-packages are already imported and what abstractions exist.

2. **Check the API reference.** When recommending an API, read `references/api-reference.md` to confirm the function exists, its exact signature, and which version introduced it. Deprecated APIs are marked — always recommend the current replacement.

3. **Consider performance.** Read `references/performance-and-patterns.md` before writing rendering code. Small decisions (draw call ordering, image reuse, avoiding `At()`) have large performance impacts.

4. **Write idiomatic Ebitengine code.** The engine has specific patterns (GeoM transform chaining, ColorScale for tinting, offscreen buffers for composition). Use them rather than inventing alternatives.

5. **Test mentally against the game loop.** Every piece of code lives in either `Update()` or `Draw()`. Logic and state changes belong in `Update()`. Rendering belongs in `Draw()`. Input reading via `inpututil` must happen in `Update()`. Violating this causes subtle bugs.

## Reference Files

Read these as needed — don't load everything upfront:

- **`references/api-reference.md`** — Complete API for ebiten/v2 and all sub-packages (inpututil, text/v2, audio, vector, colorm, ebitenutil). Consult when you need exact function signatures, type definitions, or to verify an API exists.

- **`references/performance-and-patterns.md`** — Performance tips, common patterns, Kage shader guide, FAQ solutions, and anti-patterns. Consult when writing rendering code, optimizing, debugging visual artifacts, or working with shaders.

## Key Concepts to Keep in Mind

### The Game Loop
```
┌─────────────────────────────────────────────┐
│  RunGame(game)                              │
│    ├── Update() — 60 TPS default            │
│    │     ├── Read input                     │
│    │     ├── Update game state              │
│    │     └── Return error (nil or Termination) │
│    ├── Draw(screen) — every frame           │
│    │     ├── Draw to screen (or offscreen)  │
│    │     └── No state mutation here         │
│    └── Layout(ow, oh) — on resize           │
│          └── Return logical screen size     │
└─────────────────────────────────────────────┘
```

### GeoM Transform Order Matters
Transforms apply in the order you call them. The most common pattern:
```go
op := &ebiten.DrawImageOptions{}
// 1. Center the sprite at origin (for rotation/scaling)
op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
// 2. Rotate or scale
op.GeoM.Rotate(angle)
op.GeoM.Scale(sx, sy)
// 3. Translate to world position
op.GeoM.Translate(worldX, worldY)
// 4. Apply camera transform
op.GeoM.Translate(-camX, -camY)
```
Getting this order wrong is the #1 source of rendering bugs.

### Draw Call Batching
Ebitengine automatically merges successive `DrawImage` calls into a single GPU draw call when:
- Same render target
- Same source texture atlas
- Same blend mode
- Same filter

Draw similar sprites consecutively (all floor tiles, then all entities, then all UI) to maximize batching.

### Image Management
- `NewImage` allocates GPU memory — don't call it every frame
- Reuse images via `Clear()` + redraw instead of reallocating
- `Deallocate()` explicitly frees GPU memory when done
- `SubImage()` is free (same backing texture) — use it for sprite sheets
- `At()` is expensive (forces GPU sync) — avoid in hot paths

### Input Patterns
- `ebiten.IsKeyPressed()` — continuous hold detection (movement)
- `inpututil.IsKeyJustPressed()` — single press detection (menu selection, attacks)
- `inpututil.KeyPressDuration()` — charge mechanics, hold-to-run
- All `inpututil` functions must be called in `Update()`, never in `Draw()`

### Audio Architecture
- One `audio.Context` per application (singleton)
- `NewPlayerFromBytes` for short SFX (cheap to create, shareable source)
- Reuse `Player` with `Rewind()` for BGM
- Use `InfiniteLoop` for looping music
- F32 variants are the future — prefer them for new code
- Sample rate is set once at context creation and cannot change

### Kage Shaders
- Go-like syntax but limited: no structs, no imports, no slices
- Fragment shader only (no vertex shaders)
- Entry point: `func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4`
- Use `//kage:unit pixels` for new shaders (clearer than texel coordinates)
- Up to 4 source images via `imageSrc0At()` through `imageSrc3At()`
- Uniforms are exported global variables (capitalized names)
- Compiled at runtime via `ebiten.NewShader(src)`

### Common Gotchas

1. **Tile rendering artifacts**: Using `SubImage` + scaling/rotation on non-atlas sprites causes seams. Solution: draw tiles to an offscreen buffer first, then scale the buffer.

2. **Screen as render source**: Don't use the screen image as a source for `DrawImage`. The screen clears each frame, causing expensive internal state restoration.

3. **Modifying render sources**: If you draw A→B then modify A's pixels, Ebitengine must do expensive restoration. Plan your draw order to avoid this.

4. **Layout vs window size**: `Layout()` returns the *logical* screen size, not the window size. The engine handles scaling. For hi-DPI, multiply outside dimensions by `DeviceScaleFactor()`.

5. **Deprecated APIs**: Many functions were deprecated in v2.2–v2.6. Common traps:
   - `Size()` → use `Bounds().Dx()`, `Bounds().Dy()`
   - `Dispose()` → use `Deallocate()`
   - `ReplacePixels()` → use `WritePixels()`
   - `ColorM` → use `colorm` package or `ColorScale`
   - `CompositeMode` → use `Blend`

6. **Windows console window**: Build with `-ldflags -H=windowsgui` to suppress the console window on Windows.
