# Ebitengine Performance Tips & Common Patterns

## Table of Contents

1. [Performance Tips](#performance-tips)
2. [Common Game Patterns](#common-game-patterns)
3. [Kage Shader Guide](#kage-shader-guide)
4. [FAQ & Troubleshooting](#faq--troubleshooting)
5. [Deprecated API Migration](#deprecated-api-migration)

---

## Performance Tips

### Draw Call Batching

Ebitengine automatically merges successive `DrawImage` calls into one GPU draw call when all conditions match:
- Same render target
- Same source texture (or same atlas)
- Same blend mode
- Same filter

**What to do:** Group similar draws together. Draw all floor tiles, then all entities, then all UI. Don't interleave different sprite sources.

**Debug it:** Use the `ebitenginedebug` build tag to log actual draw commands:
```bash
go run -tags ebitenginedebug .
```

### Avoid Modifying Render Sources

When you draw from image A to image B, then modify A's pixels, Ebitengine must perform expensive internal restoration. This also applies to cyclic draws (A→B then B→A).

**What to do:** Plan draw order so sources aren't modified after being read. Use separate offscreen buffers for intermediate results.

### Never Use Screen as a Render Source

The screen clears each frame. Using it as a source for `DrawImage` triggers expensive state restoration.

**What to do:** Draw to an offscreen `*ebiten.Image`, then draw that offscreen to the screen.

### Minimize At() and ReadPixels() Calls

`At()` and `ReadPixels()` force GPU synchronization — they flush all queued draw commands and wait for the GPU. This stalls the entire pipeline.

**What to do:** Avoid per-pixel reads in game logic. If you need collision data, maintain a separate data structure (tile map, spatial hash) rather than reading pixels.

### Minimize WritePixels() Calls

`WritePixels()` (formerly `ReplacePixels`) uploads data to the GPU. It's much cheaper than reading but still not free.

**What to do:** Batch pixel writes. Prepare your byte slice fully, then call `WritePixels()` once.

### Image Allocation

`NewImage()` allocates GPU memory. Creating images every frame is a major performance drain.

**What to do:**
- Allocate images once at init or on resize
- Reuse via `Clear()` + redraw instead of reallocating
- Use `Deallocate()` to free images you're truly done with
- For dynamic UI (health bars, etc.), consider reusing a single scratch image

### SubImage is Free

`SubImage()` returns a view into the same backing texture — no allocation, no copy. Use it liberally for sprite sheet indexing.

```go
// At init time, compute sub-images for each sprite
sprites := make([]*ebiten.Image, frameCount)
for i := 0; i < frameCount; i++ {
    r := image.Rect(i*w, 0, (i+1)*w, h)
    sprites[i] = sheet.SubImage(r).(*ebiten.Image)
}
```

### Discrete GPU on Windows

Windows laptops may default to integrated graphics. Export these symbols to prefer the discrete GPU:

```go
// +build windows
package main

// Force discrete GPU on Windows laptops
var (
    _ = [1]byte{0} // Ensure the variables are referenced
)

//go:linkname _ NvOptimusEnablement
var NvOptimusEnablement uint32 = 1

//go:linkname _ AmdPowerXpressRequestHighPerformance
var AmdPowerXpressRequestHighPerformance int32 = 1
```

Or use a third-party package like `github.com/AJ-325/preferdiscretegpu`.

### Audio Performance

- **SFX:** `NewPlayerFromBytes()` is cheap — create a new player per sound effect. The byte data is shared, not copied.
- **BGM:** Reuse a single `Player` and call `Rewind()` to replay. Creating players from large streams is expensive.
- **Context creation:** `NewContext()` is expensive — do it once at startup.

---

## Common Game Patterns

### Camera System
```go
type Camera struct {
    X, Y   float64 // World position (center of view)
    Zoom   float64 // Scale factor (1.0 = normal)
}

func (c *Camera) ApplyToOp(op *ebiten.DrawImageOptions, screenW, screenH int) {
    op.GeoM.Translate(-c.X, -c.Y)
    op.GeoM.Scale(c.Zoom, c.Zoom)
    op.GeoM.Translate(float64(screenW)/2, float64(screenH)/2)
}

func (c *Camera) ScreenToWorld(screenX, screenY int, screenW, screenH int) (float64, float64) {
    worldX := (float64(screenX) - float64(screenW)/2) / c.Zoom + c.X
    worldY := (float64(screenY) - float64(screenH)/2) / c.Zoom + c.Y
    return worldX, worldY
}
```

### Sprite Animation
```go
type Animation struct {
    Frames    []*ebiten.Image
    FrameTime int   // Ticks per frame
    tick      int
}

func (a *Animation) Update() {
    a.tick++
}

func (a *Animation) Frame() *ebiten.Image {
    idx := (a.tick / a.FrameTime) % len(a.Frames)
    return a.Frames[idx]
}
```

### Isometric Projection
```go
const TileWidth = 64
const TileHeight = 32

func CartToIso(x, y float64) (float64, float64) {
    isoX := (x - y) * float64(TileWidth) / 2
    isoY := (x + y) * float64(TileHeight) / 2
    return isoX, isoY
}

func IsoToCart(isoX, isoY float64) (float64, float64) {
    x := (isoX/float64(TileWidth/2) + isoY/float64(TileHeight/2)) / 2
    y := (isoY/float64(TileHeight/2) - isoX/float64(TileWidth/2)) / 2
    return x, y
}
```

### Depth Sorting (for isometric/top-down)
```go
type Renderable struct {
    Image  *ebiten.Image
    Op     *ebiten.DrawImageOptions
    Depth  float64 // Higher = drawn later (on top)
}

// Sort before drawing
sort.Slice(renderables, func(i, j int) bool {
    return renderables[i].Depth < renderables[j].Depth
})
for _, r := range renderables {
    screen.DrawImage(r.Image, r.Op)
}
```

### Offscreen Buffer Pattern
```go
var offscreen *ebiten.Image

func (g *Game) Draw(screen *ebiten.Image) {
    sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

    // Reallocate only on size change
    if offscreen == nil || offscreen.Bounds().Dx() != sw || offscreen.Bounds().Dy() != sh {
        if offscreen != nil {
            offscreen.Deallocate()
        }
        offscreen = ebiten.NewImage(sw, sh)
    }

    offscreen.Clear()
    // Draw game world to offscreen...

    // Draw offscreen to screen (with any post-processing)
    screen.DrawImage(offscreen, nil)
}
```

### Tilemap Rendering
```go
func DrawTilemap(screen *ebiten.Image, tiles [][]int, tileSheet *ebiten.Image, tileSize int, cam Camera) {
    tilesPerRow := tileSheet.Bounds().Dx() / tileSize

    for y, row := range tiles {
        for x, tileID := range row {
            if tileID == 0 { continue } // Empty tile

            // Source rect from sprite sheet
            sx := (tileID % tilesPerRow) * tileSize
            sy := (tileID / tilesPerRow) * tileSize
            src := tileSheet.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image)

            op := &ebiten.DrawImageOptions{}
            op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
            cam.ApplyToOp(op, screenW, screenH)

            screen.DrawImage(src, op)
        }
    }
}
```

### FOV / Shadow Carving with DrawTriangles
```go
// Use BlendSourceOut to "carve" light from a shadow overlay
func DrawLightWedge(shadow *ebiten.Image, vertices []ebiten.Vertex, indices []uint16) {
    white := ebiten.NewImage(1, 1)
    white.Fill(color.White)

    op := &ebiten.DrawTrianglesOptions{}
    op.Blend = ebiten.BlendSourceOut  // Subtracts from destination
    shadow.DrawTriangles(vertices, indices, white, op)
}
```

### Input Action Binding
```go
type Action int

const (
    ActionMoveUp Action = iota
    ActionMoveDown
    ActionAttack
    // ...
)

type Binding struct {
    Primary   ebiten.Key
    Secondary ebiten.Key
}

var bindings = map[Action]Binding{
    ActionMoveUp:   {Primary: ebiten.KeyW, Secondary: ebiten.KeyArrowUp},
    ActionMoveDown: {Primary: ebiten.KeyS, Secondary: ebiten.KeyArrowDown},
    ActionAttack:   {Primary: ebiten.KeySpace},
}

func IsActionJustPressed(action Action) bool {
    b := bindings[action]
    return inpututil.IsKeyJustPressed(b.Primary) ||
        (b.Secondary != 0 && inpututil.IsKeyJustPressed(b.Secondary))
}
```

### Text Rendering with text/v2
```go
import textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"

// Load font once at init
var fontSource *textv2.GoTextFaceSource

func init() {
    f, _ := os.Open("assets/font.ttf")
    fontSource, _ = textv2.NewGoTextFaceSource(f)
}

func DrawText(screen *ebiten.Image, str string, x, y, size float64) {
    face := &textv2.GoTextFace{
        Source: fontSource,
        Size:   size,
    }
    op := &textv2.DrawOptions{}
    op.GeoM.Translate(x, y)
    op.LineSpacing = size * 1.5
    textv2.Draw(screen, str, face, op)
}

func DrawCenteredText(screen *ebiten.Image, str string, cx, cy, size float64) {
    face := &textv2.GoTextFace{Source: fontSource, Size: size}
    op := &textv2.DrawOptions{}
    op.GeoM.Translate(cx, cy)
    op.PrimaryAlign = textv2.AlignCenter
    op.SecondaryAlign = textv2.AlignCenter
    op.LineSpacing = size * 1.5
    textv2.Draw(screen, str, face, op)
}
```

### Audio Setup
```go
import (
    "github.com/hajimehoshi/ebiten/v2/audio"
    "github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

var audioCtx *audio.Context
var bgmPlayer *audio.Player

func initAudio() {
    audioCtx = audio.NewContext(44100)
}

func playBGM(data []byte) {
    stream, _ := vorbis.DecodeF32(bytes.NewReader(data))
    loop := audio.NewInfiniteLoopF32(stream, stream.Length())
    bgmPlayer, _ = audioCtx.NewPlayerF32(loop)
    bgmPlayer.SetVolume(0.5)
    bgmPlayer.Play()
}

func playSFX(data []byte) {
    player := audioCtx.NewPlayerF32FromBytes(data)
    player.Play()
    // Player will be GC'd after playback finishes
}
```

---

## Kage Shader Guide

### Basics

Kage is Ebitengine's shader language. Go-like syntax, but severely limited:
- No structs, interfaces, slices, pointers, maps
- No imports, no type definitions
- No goroutines, defer, switch, goto, for-range
- Only `int`, `float`, `bool`, and vec/mat types
- Fragment shaders only (no vertex shaders)

### Entry Point Signatures
```go
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4
// dstPos: destination pixel position (xy = position, zw = unused)
// srcPos: source texture coordinates
// color: vertex color (from ColorScale)
// returns: output color (premultiplied alpha, 0-1)
```

### Unit Mode
```go
//kage:unit pixels   // Recommended — coordinates are in pixels
//kage:unit texels   // Legacy default — normalized 0-1 coordinates
```

### Uniforms
Exported (capitalized) global variables become uniforms. Set them via the `Uniforms` map in draw options.

```go
// In Kage:
var Time float
var Resolution vec2

// In Go:
op.Uniforms = map[string]any{
    "Time":       float32(time),
    "Resolution": []float32{float32(w), float32(h)},
}
```

### Source Image Access
```go
imageSrc0At(pos vec2) vec4          // Safe — returns transparent black outside bounds
imageSrc0UnsafeAt(pos vec2) vec4    // Fast — undefined outside bounds

imageSrc0Origin() vec2              // Top-left of source region
imageSrc0Size() vec2                // Size of source region
imageDstOrigin() vec2               // Top-left of destination
imageDstSize() vec2                 // Size of destination
```

Up to 4 source images: `imageSrc0` through `imageSrc3`.

### Built-in Functions

**Math:** `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2`, `pow`, `exp`, `exp2`, `log`, `log2`, `sqrt`, `inversesqrt`, `abs`, `sign`, `floor`, `ceil`, `fract`, `mod`, `min`, `max`, `clamp`, `mix`, `step`, `smoothstep`

**Vector:** `length`, `distance`, `dot`, `cross`, `normalize`, `faceforward`, `reflect`

**Matrix:** `transpose`

**Special:** `discard()` — skip this pixel, `dfdx`, `dfdy`, `fwidth` — derivatives

### Example: Grayscale Shader
```go
//kage:unit pixels

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0At(srcPos)
    gray := dot(clr.rgb, vec3(0.299, 0.587, 0.114))
    return vec4(gray, gray, gray, clr.a)
}
```

### Example: Flash White (hit effect)
```go
//kage:unit pixels

var FlashAmount float  // 0.0 = normal, 1.0 = full white

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0At(srcPos)
    white := vec4(clr.a, clr.a, clr.a, clr.a)
    return mix(clr, white, FlashAmount)
}
```

### Example: Outline Shader
```go
//kage:unit pixels

var OutlineColor vec4
var TexelSize vec2  // vec2(1.0/width, 1.0/height)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0At(srcPos)
    if clr.a > 0.0 {
        return clr
    }

    // Check neighbors
    offsets := [4]vec2{
        vec2(TexelSize.x, 0),
        vec2(-TexelSize.x, 0),
        vec2(0, TexelSize.y),
        vec2(0, -TexelSize.y),
    }

    for i := 0; i < 4; i++ {
        neighbor := imageSrc0At(srcPos + offsets[i])
        if neighbor.a > 0.0 {
            return OutlineColor
        }
    }

    return vec4(0)
}
```

### Compiling and Using Shaders
```go
shader, err := ebiten.NewShader([]byte(shaderSource))
if err != nil {
    log.Fatal(err) // Compilation error — check syntax
}

// Use with DrawRectShader
op := &ebiten.DrawRectShaderOptions{}
op.Images[0] = sourceImage
op.Uniforms = map[string]any{
    "Time": float32(g.tick) / 60.0,
}
screen.DrawRectShader(w, h, shader, op)
```

---

## FAQ & Troubleshooting

### Tile Rendering Artifacts (Seams/Holes)
**Problem:** Gaps between tiles when scaling or rotating.
**Cause:** `SubImage` + scaling on non-contiguous graphics.
**Solution:** Render tiles to an offscreen buffer at 1:1 scale, then scale/rotate the buffer.

### Console Window on Windows
**Problem:** A console window appears alongside the game window.
**Solution:** `go build -ldflags -H=windowsgui ./`

### Hi-DPI Scaling
**Problem:** Game looks blurry or wrong size on hi-DPI displays.
**Solution:** In `Layout()`, multiply outside dimensions:
```go
func (g *Game) Layout(ow, oh int) (int, int) {
    s := ebiten.Monitor().DeviceScaleFactor()
    return int(float64(ow) * s), int(float64(oh) * s)
}
```

### WebAssembly Deployment
Build with: `GOOS=js GOARCH=wasm go build -o game.wasm .`

Use the Ebitengine-specific `wasm_exec.js` from your Go installation.

### Cross-Compilation
- Windows → works (GOOS=windows)
- WebAssembly → works (GOOS=js GOARCH=wasm)
- macOS/Linux cross-compile → difficult due to Cgo

### Screen Flickering
**Common causes:**
- Drawing in `Update()` instead of `Draw()`
- Using screen as both source and destination
- Not clearing offscreen buffers between frames

### Slow Performance
**Check in order:**
1. Are you allocating `NewImage` every frame? → Reuse images
2. Are you calling `At()` or `ReadPixels()` in the game loop? → Maintain separate data
3. Are draw calls interleaved across different textures? → Group by source
4. Is `ebitenginedebug` showing excessive draw commands? → Batch better
5. Are you on integrated GPU (Windows laptop)? → Force discrete GPU

---

## Deprecated API Migration

| Deprecated | Replacement | Since |
|---|---|---|
| `Size()` | `Bounds().Dx()`, `Bounds().Dy()` | v2.1 |
| `Dispose()` | `Deallocate()` | v2.7 |
| `ReplacePixels()` | `WritePixels()` | v2.4 |
| `ColorM` (on DrawImageOptions) | `colorm` package or `ColorScale` | v2.5 |
| `CompositeMode` | `Blend` | v2.5 |
| `CurrentFPS()` | `ActualFPS()` | v2.4 |
| `CurrentTPS()` | `ActualTPS()` | v2.4 |
| `SetMaxTPS()` | `SetTPS()` | v2.4 |
| `MaxTPS()` | `TPS()` | v2.4 |
| `InputChars()` | `AppendInputChars()` | v2.2 |
| `GamepadIDs()` | `AppendGamepadIDs()` | v2.2 |
| `TouchIDs()` | `AppendTouchIDs()` | v2.2 |
| `PressedKeys()` | `AppendPressedKeys()` | v2.2 |
| `SetScreenFilterEnabled()` | `FinalScreenDrawer` interface | v2.3 |
| `SetFPSMode()` | `SetVsyncEnabled()` | v2.2 |
| `DeviceScaleFactor()` | `Monitor().DeviceScaleFactor()` | v2.6 |
| `ScreenSizeInFullscreen()` | `Monitor().Size()` | v2.6 |
| `SetWindowResizable()` | `SetWindowResizingMode()` | v2.3 |
| `IsWindowResizable()` | `WindowResizingMode()` | v2.3 |
| `SetInitFocused()` | `RunGameWithOptions` | v2.5 |
| `SetScreenTransparent()` | `RunGameWithOptions` | v2.5 |
| `DrawFilledRect()` | `FillRect()` | v2.6 |
| `DrawFilledCircle()` | `FillCircle()` | v2.6 |
| `Seek()` (audio Player) | `SetPosition()` | v2.6 |
| `Current()` (audio Player) | `Position()` | v2.6 |
