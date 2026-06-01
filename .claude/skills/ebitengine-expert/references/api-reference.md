# Ebitengine API Reference

Complete reference for ebiten/v2 v2.9+ and all sub-packages.

## Table of Contents

1. [Core Package (ebiten/v2)](#core-package)
2. [inpututil](#inpututil)
3. [text/v2](#textv2)
4. [audio](#audio)
5. [vector](#vector)
6. [ebitenutil](#ebitenutil)
7. [colorm](#colorm)
8. [Types & Constants](#types--constants)

---

## Core Package

### Game Interface (required)
```go
type Game interface {
    Update() error
    Draw(screen *ebiten.Image)
    Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}
```

### Optional Interfaces
```go
// Custom final screen rendering (v2.9+)
type FinalScreenDrawer interface {
    DrawFinalScreen(screen FinalScreen, offscreen *Image, geoM GeoM)
}

// Alternative layout with float precision
type LayoutFer interface {
    LayoutF(outsideWidth, outsideHeight float64) (screenWidth, screenHeight float64)
}
```

### Game Execution
```go
func RunGame(game Game) error
func RunGameWithOptions(game Game, options *RunGameOptions) error  // v2.5+

type RunGameOptions struct {
    InitUnfocused     bool
    ScreenTransparent bool
    GraphicsLibrary   GraphicsLibrary
    SingleThread      bool  // v2.8+
}
```

### Image
```go
// Creation
func NewImage(width, height int) *Image
func NewImageWithOptions(bounds image.Rectangle, options *NewImageOptions) *Image
func NewImageFromImage(source image.Image) *Image
func NewImageFromImageWithOptions(source image.Image, options *NewImageFromImageOptions) *Image

// Drawing
func (i *Image) DrawImage(img *Image, options *DrawImageOptions)
func (i *Image) DrawTriangles(vertices []Vertex, indices []uint16, img *Image, options *DrawTrianglesOptions)
func (i *Image) DrawTriangles32(vertices []Vertex, indices []uint32, img *Image, options *DrawTrianglesOptions)
func (i *Image) DrawTrianglesShader(vertices []Vertex, indices []uint16, shader *Shader, options *DrawTrianglesShaderOptions)
func (i *Image) DrawTrianglesShader32(vertices []Vertex, indices []uint32, shader *Shader, options *DrawTrianglesShaderOptions)
func (i *Image) DrawRectShader(width, height int, shader *Shader, options *DrawRectShaderOptions)

// Pixel operations
func (i *Image) Clear()
func (i *Image) Fill(clr color.Color)
func (i *Image) Set(x, y int, clr color.Color)
func (i *Image) At(x, y int) color.Color              // EXPENSIVE — forces GPU sync
func (i *Image) RGBA64At(x, y int) color.RGBA64
func (i *Image) ReadPixels(pixels []byte)
func (i *Image) WritePixels(pixels []byte)

// Properties
func (i *Image) Bounds() image.Rectangle
func (i *Image) SubImage(r image.Rectangle) image.Image  // Free — shares texture
func (i *Image) ColorModel() color.Model
func (i *Image) Deallocate()                              // Free GPU memory
```

### Draw Options
```go
type DrawImageOptions struct {
    GeoM       GeoM
    ColorScale ColorScale
    Filter     Filter
    Blend      Blend
}

type DrawTrianglesOptions struct {
    ColorScale    ColorScale
    Filter        Filter
    Address       Address
    Blend         Blend
    FillRule      FillRule
    AntiAlias     bool
}

type DrawTrianglesShaderOptions struct {
    Uniforms   map[string]any
    Images     [4]*Image
    ColorScale ColorScale
    Blend      Blend
    FillRule   FillRule
    AntiAlias  bool
}

type DrawRectShaderOptions struct {
    Uniforms   map[string]any
    Images     [4]*Image
    ColorScale ColorScale
    Blend      Blend
    GeoM       GeoM
}

type Vertex struct {
    DstX, DstY             float32
    SrcX, SrcY             float32
    ColorR, ColorG, ColorB, ColorA float32
}
```

### GeoM (Geometric Transform Matrix)
```go
type GeoM struct{}

func (g *GeoM) Reset()
func (g *GeoM) Translate(tx, ty float64)
func (g *GeoM) Scale(x, y float64)
func (g *GeoM) Rotate(theta float64)              // Radians
func (g *GeoM) Skew(skewX, skewY float64)
func (g *GeoM) Concat(other GeoM)
func (g *GeoM) Invert()
func (g *GeoM) IsInvertible() bool
func (g *GeoM) Apply(x, y float64) (float64, float64)
func (g *GeoM) Element(i, j int) float64
func (g *GeoM) SetElement(i, j int, element float64)
func (g *GeoM) String() string
```

### ColorScale
```go
type ColorScale struct{}

func (c *ColorScale) Reset()
func (c *ColorScale) Scale(r, g, b, a float32)
func (c *ColorScale) ScaleAlpha(a float32)
func (c *ColorScale) ScaleWithColor(clr color.Color)
func (c *ColorScale) ScaleWithColorScale(other ColorScale)
func (c *ColorScale) SetR(r float32)
func (c *ColorScale) SetG(g float32)
func (c *ColorScale) SetB(b float32)
func (c *ColorScale) SetA(a float32)
func (c *ColorScale) R() float32
func (c *ColorScale) G() float32
func (c *ColorScale) B() float32
func (c *ColorScale) A() float32
```

### Blend (Compositing)
```go
type Blend struct {
    BlendFactorSourceRGB      BlendFactor
    BlendFactorSourceAlpha    BlendFactor
    BlendFactorDestinationRGB BlendFactor
    BlendFactorDestinationAlpha BlendFactor
    BlendOperationRGB         BlendOperation
    BlendOperationAlpha       BlendOperation
}

// Presets (Porter-Duff)
var BlendSourceOver      Blend  // Default — standard alpha compositing
var BlendClear           Blend
var BlendCopy            Blend
var BlendDestination     Blend
var BlendDestinationOver Blend
var BlendSourceIn        Blend
var BlendDestinationIn   Blend
var BlendSourceOut       Blend  // Useful for shadow/light carving
var BlendDestinationOut  Blend
var BlendSourceAtop      Blend
var BlendDestinationAtop Blend
var BlendXor             Blend
var BlendLighter         Blend  // Additive blending
```

### Shader
```go
func NewShader(src []byte) (*Shader, error)
func (s *Shader) Deallocate()
```

### Window Management
```go
func WindowSize() (int, int)
func SetWindowSize(width, height int)
func WindowPosition() (int, int)
func SetWindowPosition(x, y int)
func SetWindowTitle(title string)
func SetWindowIcon(iconImages []image.Image)
func SetWindowDecorated(decorated bool)
func IsWindowDecorated() bool
func SetWindowResizingMode(mode WindowResizingModeType)
func WindowResizingMode() WindowResizingModeType
func SetWindowSizeLimits(minw, minh, maxw, maxh int)
func WindowSizeLimits() (minw, minh, maxw, maxh int)
func SetWindowFloating(float bool)
func IsWindowFloating() bool
func MaximizeWindow()
func IsWindowMaximized() bool
func MinimizeWindow()
func IsWindowMinimized() bool
func RestoreWindow()
func SetWindowClosingHandled(handled bool)  // v2.2+
func IsWindowBeingClosed() bool             // v2.2+
func SetFullscreen(fullscreen bool)
func IsFullscreen() bool
func SetWindowMousePassthrough(enabled bool) // v2.6+
func RequestAttention()                      // v2.8+
```

### Frame & Timing
```go
func ActualFPS() float64    // v2.4+
func ActualTPS() float64    // v2.4+
func TPS() int
func SetTPS(tps int)        // v2.4+ (default 60, use SyncWithFPS for uncapped)
func Tick() int64

const DefaultTPS = 60
const SyncWithFPS = -1     // Uncap TPS to match FPS
```

### Display
```go
func SetVsyncEnabled(enabled bool)
func IsVsyncEnabled() bool
func IsFocused() bool
func SetRunnableOnUnfocused(runnableOnUnfocused bool)
func IsRunnableOnUnfocused() bool
func SetScreenClearedEveryFrame(cleared bool)
func IsScreenClearedEveryFrame() bool

func Monitor() *MonitorType
func AppendMonitors(monitors []*MonitorType) []*MonitorType
func (m *MonitorType) Name() string
func (m *MonitorType) Size() (int, int)
func (m *MonitorType) DeviceScaleFactor() float64
```

### Keyboard Input
```go
func IsKeyPressed(key Key) bool
func AppendInputChars(runes []rune) []rune  // v2.2+ — typed characters
func KeyName(key Key) string                // v2.5+ — physical key name
```

### Mouse Input
```go
func CursorPosition() (x, y int)
func IsMouseButtonPressed(mouseButton MouseButton) bool
func Wheel() (xoff, yoff float64)
func SetCursorMode(mode CursorModeType)
func CursorMode() CursorModeType
func SetCursorShape(shape CursorShapeType)  // v2.1+
func CursorShape() CursorShapeType
```

### Gamepad Input
```go
func AppendGamepadIDs(gamepadIDs []GamepadID) []GamepadID
func GamepadName(id GamepadID) string
func GamepadSDLID(id GamepadID) string
func IsGamepadButtonPressed(id GamepadID, button GamepadButton) bool
func GamepadAxisValue(id GamepadID, axis GamepadAxisType) float64
func GamepadButtonCount(id GamepadID) int
func GamepadAxisCount(id GamepadID) int

// Standard gamepad layout (v2.2+)
func IsStandardGamepadLayoutAvailable(id GamepadID) bool
func IsStandardGamepadButtonPressed(id GamepadID, button StandardGamepadButton) bool
func StandardGamepadButtonValue(id GamepadID, button StandardGamepadButton) float64
func StandardGamepadAxisValue(id GamepadID, axis StandardGamepadAxis) float64
func UpdateStandardGamepadLayoutMappings(mappings string) (bool, error)
```

### Touch Input
```go
func AppendTouchIDs(touches []TouchID) []TouchID
func TouchPosition(id TouchID) (int, int)
```

### Haptics
```go
func Vibrate(options *VibrateOptions)
func VibrateGamepad(gamepadID GamepadID, options *VibrateGamepadOptions)
```

### File Drop
```go
func DroppedFiles() fs.FS  // v2.5+ — desktop/browser
```

### Debug
```go
func ReadDebugInfo(d *DebugInfo)

type DebugInfo struct {
    GraphicsLibrary GraphicsLibrary
}
```

### Graceful Termination
```go
var Termination error  // Return from Update() to quit cleanly
```

---

## inpututil

All functions must be called in `Update()`, not `Draw()`. All are concurrent-safe.

### Keyboard
```go
func IsKeyJustPressed(key ebiten.Key) bool
func IsKeyJustReleased(key ebiten.Key) bool
func KeyPressDuration(key ebiten.Key) int                    // Ticks held
func AppendPressedKeys(keys []ebiten.Key) []ebiten.Key
func AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key
func AppendJustReleasedKeys(keys []ebiten.Key) []ebiten.Key
```

### Mouse
```go
func IsMouseButtonJustPressed(button ebiten.MouseButton) bool
func IsMouseButtonJustReleased(button ebiten.MouseButton) bool
func MouseButtonPressDuration(button ebiten.MouseButton) int
```

### Gamepad
```go
func AppendJustConnectedGamepadIDs(ids []ebiten.GamepadID) []ebiten.GamepadID
func IsGamepadJustDisconnected(id ebiten.GamepadID) bool
func IsGamepadButtonJustPressed(id ebiten.GamepadID, button ebiten.GamepadButton) bool
func IsGamepadButtonJustReleased(id ebiten.GamepadID, button ebiten.GamepadButton) bool
func GamepadButtonPressDuration(id ebiten.GamepadID, button ebiten.GamepadButton) int
func AppendPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton
func AppendJustPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton
func AppendJustReleasedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton
```

### Standard Gamepad
```go
func IsStandardGamepadButtonJustPressed(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool
func IsStandardGamepadButtonJustReleased(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool
func StandardGamepadButtonPressDuration(id ebiten.GamepadID, button ebiten.StandardGamepadButton) int
func AppendPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton
func AppendJustPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton
func AppendJustReleasedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton
```

### Touch
```go
func AppendJustPressedTouchIDs(touchIDs []ebiten.TouchID) []ebiten.TouchID
func AppendJustReleasedTouchIDs(touchIDs []ebiten.TouchID) []ebiten.TouchID
func IsTouchJustReleased(id ebiten.TouchID) bool
func TouchPressDuration(id ebiten.TouchID) int
func TouchPositionInPreviousTick(id ebiten.TouchID) (int, int)
```

---

## text/v2

### Drawing Text
```go
func Draw(dst *ebiten.Image, text string, face Face, options *DrawOptions)
func Measure(text string, face Face, lineSpacingInPixels float64) (width, height float64)
func Advance(text string, face Face) float64
func CacheGlyphs(text string, face Face)
func AppendGlyphs(glyphs []Glyph, text string, face Face, options *LayoutOptions) []Glyph
func AppendVectorPath(path *vector.Path, text string, face Face, options *LayoutOptions)
```

### Face Interface
```go
type Face interface {
    Metrics() Metrics
}
```

### GoTextFace (primary — OpenType/TrueType)
```go
type GoTextFace struct {
    Source    *GoTextFaceSource
    Direction Direction
    Size      float64
    Language  language.Tag
}

func (g *GoTextFace) Metrics() Metrics
func (g *GoTextFace) SetVariation(tag Tag, value float32)
func (g *GoTextFace) RemoveVariation(tag Tag)
func (g *GoTextFace) SetFeature(tag Tag, value uint32)
func (g *GoTextFace) RemoveFeature(tag Tag)

// Source (shared glyph cache — create once, reuse)
func NewGoTextFaceSource(source io.Reader) (*GoTextFaceSource, error)
func NewGoTextFaceSourcesFromCollection(source io.Reader) ([]*GoTextFaceSource, error)
func (g *GoTextFaceSource) Metadata() Metadata
```

### GoXFace (legacy — golang.org/x/image/font)
```go
func NewGoXFace(face font.Face) *GoXFace
func (g *GoXFace) Metrics() Metrics
```

### Composite Faces
```go
func NewLimitedFace(face Face) *LimitedFace       // Unicode range filtering
func (l *LimitedFace) AddUnicodeRange(start, end rune)

func NewMultiFace(faces ...Face) (*MultiFace, error) // Fallback chain
```

### Options
```go
type DrawOptions struct {
    ebiten.DrawImageOptions
    LayoutOptions
}

type LayoutOptions struct {
    LineSpacing    float64  // Baseline-to-baseline distance (pixels)
    PrimaryAlign   Align    // Along text direction
    SecondaryAlign Align    // Perpendicular to text direction
}

const (
    AlignStart  Align = iota  // Left for LTR, Top for vertical
    AlignCenter
    AlignEnd                   // Right for LTR, Bottom for vertical
)
```

### Metrics
```go
type Metrics struct {
    HLineGap, HAscent, HDescent float64  // Horizontal text metrics
    VLineGap, VAscent, VDescent float64  // Vertical text metrics
    XHeight, CapHeight          float64
}
```

### Direction
```go
const (
    DirectionLeftToRight Direction = iota
    DirectionRightToLeft
    DirectionTopToBottomAndLeftToRight
    DirectionTopToBottomAndRightToLeft
)
```

---

## audio

### Context (singleton per app)
```go
func NewContext(sampleRate int) *Context
func CurrentContext() *Context

func (c *Context) IsReady() bool
func (c *Context) SampleRate() int
func (c *Context) NewPlayer(src io.Reader) (*Player, error)
func (c *Context) NewPlayerF32(src io.Reader) (*Player, error)
func (c *Context) NewPlayerFromBytes(src []byte) *Player
func (c *Context) NewPlayerF32FromBytes(src []byte) *Player
```

### Player
```go
func (p *Player) Play()
func (p *Player) Pause()
func (p *Player) Rewind() error
func (p *Player) IsPlaying() bool
func (p *Player) Close() error
func (p *Player) SetPosition(offset time.Duration) error  // v2.6+
func (p *Player) Position() time.Duration                 // v2.6+
func (p *Player) SetVolume(volume float64)                // 0.0 to 1.0+
func (p *Player) Volume() float64
func (p *Player) SetBufferSize(bufferSize time.Duration)  // v2.3+
```

### InfiniteLoop (for BGM)
```go
func NewInfiniteLoop(src io.ReadSeeker, length int64) *InfiniteLoop
func NewInfiniteLoopF32(src io.ReadSeeker, length int64) *InfiniteLoop
func NewInfiniteLoopWithIntro(src io.ReadSeeker, introLength, loopLength int64) *InfiniteLoop
func NewInfiniteLoopWithIntroF32(src io.ReadSeeker, introLength, loopLength int64) *InfiniteLoop

func (i *InfiniteLoop) Read(b []byte) (int, error)
func (i *InfiniteLoop) Seek(offset int64, whence int) (int64, error)
```

### Resampling
```go
func Resample(source io.ReadSeeker, size int64, from, to int) io.ReadSeeker
func ResampleF32(source io.ReadSeeker, size int64, from, to int) io.ReadSeeker
```

### Audio Format Decoders (sub-packages)
```go
// audio/mp3
func mp3.DecodeF32(r io.ReadSeeker) (*mp3.Stream, error)

// audio/vorbis
func vorbis.DecodeF32(r io.ReadSeeker) (*vorbis.Stream, error)

// audio/wav
func wav.DecodeF32(r io.ReadSeeker) (*wav.Stream, error)

// All streams provide: Read, Seek, Length, SampleRate
```

### Key Audio Patterns
- **SFX**: `ctx.NewPlayerFromBytes(data)` per sound — cheap, shareable source
- **BGM**: Reuse one `Player`, call `Rewind()` to replay
- **Looping**: Wrap source in `NewInfiniteLoop` before creating player
- **F32 preferred**: Use F32 variants for all new code
- **Loop seams**: Add ~0.1s extra data past loop point to avoid clicks (lossy formats)

---

## vector

### Simple Drawing (convenience functions)
```go
func StrokeLine(dst *ebiten.Image, x0, y0, x1, y1 float32, strokeWidth float32, clr color.Color, antialias bool)
func FillRect(dst *ebiten.Image, x, y, width, height float32, clr color.Color, antialias bool)
func StrokeRect(dst *ebiten.Image, x, y, width, height float32, strokeWidth float32, clr color.Color, antialias bool)
func FillCircle(dst *ebiten.Image, cx, cy, r float32, clr color.Color, antialias bool)
func StrokeCircle(dst *ebiten.Image, cx, cy, r float32, strokeWidth float32, clr color.Color, antialias bool)
```

### Path (complex shapes)
```go
type Path struct{}

func (p *Path) MoveTo(x, y float32)
func (p *Path) LineTo(x, y float32)
func (p *Path) QuadTo(x1, y1, x2, y2 float32)
func (p *Path) CubicTo(x1, y1, x2, y2, x3, y3 float32)
func (p *Path) Arc(x, y, radius, startAngle, endAngle float32, dir Direction)
func (p *Path) ArcTo(x1, y1, x2, y2, radius float32)
func (p *Path) Close()
func (p *Path) Reset()
func (p *Path) Bounds() image.Rectangle
func (p *Path) AddPath(src *Path, options *AddPathOptions)

func FillPath(dst *ebiten.Image, path *Path, fillOpts *FillOptions, drawOpts *DrawPathOptions)
func StrokePath(dst *ebiten.Image, path *Path, strokeOpts *StrokeOptions, drawOpts *DrawPathOptions)
```

### Stroke/Fill Options
```go
type StrokeOptions struct {
    Width      float32
    LineCap    LineCap   // LineCapButt, LineCapRound, LineCapSquare
    LineJoin   LineJoin  // LineJoinMiter, LineJoinBevel, LineJoinRound
    MiterLimit float32
}

type FillOptions struct {
    FillRule FillRule  // FillRuleNonZero, FillRuleEvenOdd
}

type DrawPathOptions struct {
    AntiAlias  bool
    ColorScale ebiten.ColorScale
    Blend      ebiten.Blend
}
```

### Deprecated (still works but prefer the above)
```go
func DrawFilledRect(...)   // Use FillRect
func DrawFilledCircle(...) // Use FillCircle
```

---

## ebitenutil

```go
func DebugPrint(image *ebiten.Image, str string)
func DebugPrintAt(image *ebiten.Image, str string, x, y int)
func NewImageFromFile(path string) (*ebiten.Image, image.Image, error)
func NewImageFromURL(url string) (*ebiten.Image, error)  // WebAssembly only
```

---

## colorm (v2.5+ — replaces deprecated ColorM)

```go
type ColorM struct{}

func (c *ColorM) Reset()
func (c *ColorM) Scale(r, g, b, a float64)
func (c *ColorM) Translate(r, g, b, a float64)
func (c *ColorM) RotateHue(theta float64)
func (c *ColorM) ChangeHSV(hueTheta, saturationScale, valueScale float64)
func (c *ColorM) Concat(other ColorM)
func (c *ColorM) Invert()
func (c *ColorM) IsInvertible() bool
func (c *ColorM) Apply(clr color.Color) color.Color

func DrawImage(dst, src *ebiten.Image, colorM ColorM, options *DrawImageOptions)
func DrawTriangles(dst *ebiten.Image, vertices []ebiten.Vertex, indices []uint16, src *ebiten.Image, colorM ColorM, options *DrawTrianglesOptions)
```

Use `colorm` when you need full 5x4 color matrix operations (hue rotation, saturation adjustment, color inversion). For simple tinting/alpha, `ColorScale` on `DrawImageOptions` is simpler and faster.

---

## Types & Constants

### Filter
```go
const (
    FilterNearest Filter = iota  // Pixelated (good for pixel art)
    FilterLinear                  // Smooth (good for scaled photos/UI)
)
```

### Address (texture wrapping)
```go
const (
    AddressUnsafe         Address = iota  // Undefined outside bounds
    AddressClampToZero                     // Transparent outside bounds
    AddressRepeat                          // Tile
    AddressMirroredRepeat                  // Tile with mirroring (v2.9+)
)
```

### Key Constants
All keyboard keys: `KeyA`–`KeyZ`, `Key0`–`Key9`, `KeySpace`, `KeyEnter`, `KeyEscape`, `KeyArrowUp/Down/Left/Right`, `KeyShiftLeft/Right`, `KeyControlLeft/Right`, `KeyAltLeft/Right`, `KeyTab`, `KeyBackspace`, `KeyDelete`, `KeyHome`, `KeyEnd`, `KeyPageUp/Down`, `KeyF1`–`KeyF12`, etc.

### Mouse Buttons
```go
const (
    MouseButtonLeft   MouseButton = iota
    MouseButtonMiddle
    MouseButtonRight
)
```

### Window Resizing
```go
const (
    WindowResizingModeDisabled               WindowResizingModeType = iota
    WindowResizingModeOnlyFullscreenEnabled
    WindowResizingModeEnabled
)
```

### Cursor Modes
```go
const (
    CursorModeVisible  CursorModeType = iota
    CursorModeHidden
    CursorModeCaptured  // FPS-style mouse lock
)
```

### Cursor Shapes
```go
const (
    CursorShapeDefault CursorShapeType = iota
    CursorShapeText
    CursorShapePointer
    CursorShapeMove
    // ... resize shapes, not-allowed, etc.
)
```

### Graphics Libraries
```go
const (
    GraphicsLibraryUnknown     GraphicsLibrary = iota
    GraphicsLibraryOpenGL
    GraphicsLibraryDirectX
    GraphicsLibraryMetal
    GraphicsLibraryPlayStation5
)
```

### Build Tags
- `ebitenginedebug` — Log all graphics commands
- `ebitenginegldebug` — OpenGL debug mode
- `microsoftgdk` — Xbox support
- `nintendosdk` — Nintendo Switch support

### Environment Variables
- `EBITENGINE_SCREENSHOT_KEY` — Key to capture screenshot (e.g., "q")
- `EBITENGINE_GRAPHICS_LIBRARY` — Force graphics backend (auto/opengl/directx/metal)
- `EBITENGINE_DIRECTX` — DirectX options (version, debug, warp)
