# Kage Shader Recipes

Ready-to-use shader implementations for common game effects. Each recipe includes the Kage source and Go integration code.

## Table of Contents

1. [Passthrough (identity)](#passthrough)
2. [Grayscale](#grayscale)
3. [Flash White (hit effect)](#flash-white)
4. [Tint / Color Overlay](#tint--color-overlay)
5. [Outline / Stroke](#outline)
6. [Dissolve / Disintegrate](#dissolve)
7. [Chromatic Aberration](#chromatic-aberration)
8. [Radial Blur](#radial-blur)
9. [Gaussian Blur (single-pass)](#gaussian-blur)
10. [Normal Map Lighting](#normal-map-lighting)
11. [Water Reflection](#water-reflection)
12. [CRT Scanlines](#crt-scanlines)
13. [Vignette](#vignette)
14. [Pixelation / Mosaic](#pixelation)
15. [Wave Distortion](#wave-distortion)
16. [Color Grading (brightness/contrast/saturation)](#color-grading)
17. [Procedural Noise](#procedural-noise)
18. [Screen Shake Distortion](#screen-shake-distortion)
19. [Palette Swap](#palette-swap)
20. [Silhouette / Shadow](#silhouette)
21. [Go Integration Patterns](#go-integration-patterns)

---

## Passthrough

The simplest possible shader — outputs the source pixel unchanged.

```
//kage:unit pixels

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    return imageSrc0UnsafeAt(srcPos)
}
```

---

## Grayscale

Convert to grayscale using luminance weights (ITU-R BT.601).

```
//kage:unit pixels

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    gray := dot(clr.rgb, vec3(0.299, 0.587, 0.114))
    return vec4(gray, gray, gray, clr.a)
}
```

**Variation — partial desaturation:**
```
var Amount float  // 0 = full color, 1 = full grayscale

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    gray := dot(clr.rgb, vec3(0.299, 0.587, 0.114))
    rgb := mix(clr.rgb, vec3(gray), Amount)
    return vec4(rgb, clr.a)
}
```

---

## Flash White

Flash a sprite white on hit. Classic action game feedback.

```
//kage:unit pixels

var FlashAmount float  // 0.0 = normal, 1.0 = full white

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    // White = (alpha, alpha, alpha, alpha) in premultiplied space
    white := vec4(clr.a, clr.a, clr.a, clr.a)
    return mix(clr, white, FlashAmount)
}
```

**Go side:**
```go
// In Update(), on hit:
g.flashTimer = 0.15 // seconds

// Each Update():
if g.flashTimer > 0 {
    g.flashTimer -= 1.0 / float64(ebiten.TPS())
}

// In Draw():
op.Uniforms = map[string]any{
    "FlashAmount": float32(math.Max(0, g.flashTimer) / 0.15),
}
```

---

## Tint / Color Overlay

Apply a colored tint over the sprite.

```
//kage:unit pixels

var TintColor vec4   // e.g., vec4(1, 0, 0, 0.5) for red 50%
var TintAmount float // 0 = original, 1 = full tint

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    tinted := vec4(mix(clr.rgb, TintColor.rgb * clr.a, TintAmount), clr.a)
    return tinted
}
```

---

## Outline

Draw a colored outline around non-transparent pixels. Checks 4 neighbors.

```
//kage:unit pixels

var OutlineColor vec4  // e.g., vec4(0, 0, 0, 1) for black outline
var OutlineWidth float // typically 1.0 or 2.0

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0At(srcPos)
    if clr.a > 0.01 {
        return clr  // Already a visible pixel — return as-is
    }

    // Check cardinal neighbors
    offsets := [4]vec2{
        vec2(OutlineWidth, 0),
        vec2(-OutlineWidth, 0),
        vec2(0, OutlineWidth),
        vec2(0, -OutlineWidth),
    }

    for i := 0; i < 4; i++ {
        neighbor := imageSrc0At(srcPos + offsets[i])
        if neighbor.a > 0.01 {
            return OutlineColor
        }
    }

    return vec4(0)
}
```

**8-direction outline (thicker, higher quality):**
```
//kage:unit pixels

var OutlineColor vec4
var OutlineWidth float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0At(srcPos)
    if clr.a > 0.01 {
        return clr
    }

    offsets := [8]vec2{
        vec2(OutlineWidth, 0),
        vec2(-OutlineWidth, 0),
        vec2(0, OutlineWidth),
        vec2(0, -OutlineWidth),
        vec2(OutlineWidth, OutlineWidth),
        vec2(-OutlineWidth, OutlineWidth),
        vec2(OutlineWidth, -OutlineWidth),
        vec2(-OutlineWidth, -OutlineWidth),
    }

    for i := 0; i < 8; i++ {
        if imageSrc0At(srcPos + offsets[i]).a > 0.01 {
            return OutlineColor
        }
    }

    return vec4(0)
}
```

---

## Dissolve

Dissolve a sprite using a noise texture threshold.

```
//kage:unit pixels

var Threshold float  // 0.0 = fully visible, 1.0 = fully dissolved

func hash(p vec2) float {
    h := dot(p, vec2(127.1, 311.7))
    return fract(sin(h) * 43758.5453)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    pos := dstPos.xy - imageDstOrigin()
    noise := hash(pos)

    if noise < Threshold {
        discard()
    }

    // Optional: glow at dissolve edge
    edgeWidth := 0.05
    if noise < Threshold + edgeWidth {
        t := (noise - Threshold) / edgeWidth
        clr.rgb = mix(vec3(1, 0.5, 0), clr.rgb, t)  // Orange edge glow
    }

    return clr
}
```

**With a noise texture (smoother):**
```
//kage:unit pixels

var Threshold float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    noise := imageSrc1At(srcPos).r  // Noise texture in image slot 1

    if noise < Threshold {
        discard()
    }

    return clr
}
```

---

## Chromatic Aberration

Split RGB channels with offset. Great for damage/glitch effects.

```
//kage:unit pixels

var Amount vec2  // Offset amount in pixels, e.g., vec2(3, 1)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    var clr vec3
    clr.r = imageSrc0At(srcPos + Amount).r
    clr.g = imageSrc0UnsafeAt(srcPos).g
    clr.b = imageSrc0At(srcPos - Amount).b
    return vec4(clr, 1)
}
```

**Cursor-relative version (from Ebitengine examples):**
```
//kage:unit pixels

var Cursor vec2  // Mouse position

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    center := imageDstSize() / 2
    amount := (center - Cursor) / 10
    var clr vec3
    clr.r = imageSrc0At(srcPos + amount).r
    clr.g = imageSrc0UnsafeAt(srcPos).g
    clr.b = imageSrc0At(srcPos - amount).b
    return vec4(clr, 1)
}
```

---

## Radial Blur

Blur radiating outward from a center point.

```
//kage:unit pixels

var Center vec2   // Center point in destination pixels
var Strength float // Blur intensity (0-1)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    pos := dstPos.xy - imageDstOrigin()
    dir := normalize(pos - Center)

    samples := [...]float{-22, -14, -8, -4, -2, 2, 4, 8, 14, 22}
    sum := imageSrc0UnsafeAt(srcPos)
    for i := 0; i < len(samples); i++ {
        sum += imageSrc0At(srcPos + dir*samples[i]*Strength)
    }
    sum /= float(len(samples)) + 1

    original := imageSrc0UnsafeAt(srcPos)
    dist := distance(pos, Center)
    t := clamp(dist / 256, 0, 1)
    return mix(original, sum, t)
}
```

---

## Gaussian Blur

Single-pass approximation. For true Gaussian, use two passes (horizontal + vertical) with offscreen buffers.

```
//kage:unit pixels

var Direction vec2  // vec2(1, 0) for horizontal, vec2(0, 1) for vertical
var Radius float    // Blur radius in pixels

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    // 9-tap Gaussian weights
    weights := [...]float{
        0.0162, 0.0540, 0.1218, 0.1944, 0.2270,
        0.1944, 0.1218, 0.0540, 0.0162,
    }
    offsets := [...]float{-4, -3, -2, -1, 0, 1, 2, 3, 4}

    sum := vec4(0)
    for i := 0; i < 9; i++ {
        offset := Direction * offsets[i] * Radius / 4
        sum += imageSrc0At(srcPos + offset) * weights[i]
    }
    return sum
}
```

**Two-pass Go integration:**
```go
// Pass 1: horizontal blur to offscreen
hOp := &ebiten.DrawRectShaderOptions{}
hOp.Images[0] = source
hOp.Uniforms = map[string]any{
    "Direction": []float32{1, 0},
    "Radius":    float32(radius),
}
offscreen.DrawRectShader(w, h, blurShader, hOp)

// Pass 2: vertical blur to screen
vOp := &ebiten.DrawRectShaderOptions{}
vOp.Images[0] = offscreen
vOp.Uniforms = map[string]any{
    "Direction": []float32{0, 1},
    "Radius":    float32(radius),
}
screen.DrawRectShader(w, h, blurShader, vOp)
```

---

## Normal Map Lighting

Simple directional lighting using a normal map texture.

```
//kage:unit pixels

var LightDir vec2  // Light direction (normalized), e.g., from cursor

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)

    // Normal map in image slot 1 (RGB → XYZ normals)
    normal := imageSrc1UnsafeAt(srcPos).xyz
    normal = normal*2 - 1  // Unpack from 0-1 to -1 to 1

    // Light calculation
    lightDir3 := normalize(vec3(LightDir, 1))
    diffuse := max(dot(normal, lightDir3), 0.0)

    ambient := 0.25
    lighting := ambient + diffuse * 0.75

    return vec4(clr.rgb * lighting, clr.a)
}
```

---

## Water Reflection

Horizontal flip with wave distortion. Classic water surface effect.

```
//kage:unit pixels

var Time float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    pos := dstPos.xy - imageDstOrigin()
    size := imageDstSize()
    border := size.y*0.6 + 4*cos(Time*3 + pos.y/10)

    if pos.y < border {
        return imageSrc0UnsafeAt(srcPos)  // Above water line — normal
    }

    // Below water: sample from reflected position with wave distortion
    reflectY := border - (pos.y - border)
    offsetX := 4 * cos(Time*2 + pos.y/8)
    offsetY := 2 * sin(Time*3 + pos.x/12)

    reflectedPos := srcPos + vec2(offsetX, reflectY - pos.y + offsetY)
    clr := imageSrc0At(reflectedPos)

    // Tint water
    waterTint := vec4(0.2, 0.4, 0.6, 0.5)
    return mix(clr, waterTint, 0.3)
}
```

---

## CRT Scanlines

Retro CRT monitor effect with scanlines and slight curvature.

```
//kage:unit pixels

var ScanlineIntensity float  // 0.0–1.0, typically 0.3
var Time float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    pos := dstPos.xy - imageDstOrigin()

    // Scanline effect (every other row is darker)
    scanline := sin(pos.y * 3.14159) * 0.5 + 0.5
    scanline = 1.0 - ScanlineIntensity * (1.0 - scanline)

    // Slight RGB shift for CRT color fringing
    clr.r *= scanline * 1.05
    clr.g *= scanline
    clr.b *= scanline * 0.95

    return clr
}
```

---

## Vignette

Darken screen edges for cinematic focus.

```
//kage:unit pixels

var Intensity float  // 0.0–1.0, typically 0.5
var Softness float   // Edge softness, typically 0.4

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)

    // Normalize position to -1..1
    uv := (dstPos.xy - imageDstOrigin()) / imageDstSize() * 2 - 1

    // Vignette factor: distance from center
    dist := length(uv)
    vignette := 1.0 - smoothstep(1.0 - Softness, 1.0, dist) * Intensity

    return vec4(clr.rgb * vignette, clr.a)
}
```

---

## Pixelation

Snap pixels to a grid for retro/mosaic effect.

```
//kage:unit pixels

var PixelSize float  // Block size in pixels, e.g., 4.0 or 8.0

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    // Snap source coordinates to pixel grid
    snapped := floor(srcPos / PixelSize) * PixelSize + PixelSize/2
    return imageSrc0At(snapped)
}
```

---

## Wave Distortion

Sine-wave distortion for underwater, heat haze, or dream sequences.

```
//kage:unit pixels

var Time float
var Amplitude float  // Wave height in pixels (e.g., 3.0)
var Frequency float  // Wave frequency (e.g., 0.1)
var Speed float      // Animation speed (e.g., 2.0)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    offset := vec2(
        sin(srcPos.y*Frequency + Time*Speed) * Amplitude,
        cos(srcPos.x*Frequency + Time*Speed) * Amplitude * 0.5,
    )
    return imageSrc0At(srcPos + offset)
}
```

---

## Color Grading

Adjust brightness, contrast, and saturation.

```
//kage:unit pixels

var Brightness float  // -1 to 1 (0 = unchanged)
var Contrast float    // 0 to 2 (1 = unchanged)
var Saturation float  // 0 to 2 (1 = unchanged)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)

    // Brightness
    rgb := clr.rgb + Brightness * clr.a

    // Contrast (around midpoint)
    rgb = (rgb - 0.5*clr.a) * Contrast + 0.5*clr.a

    // Saturation
    gray := dot(rgb, vec3(0.299, 0.587, 0.114))
    rgb = mix(vec3(gray), rgb, Saturation)

    return vec4(clamp(rgb, 0, clr.a), clr.a)
}
```

---

## Procedural Noise

Hash-based noise for use in other effects. No texture needed.

```
//kage:unit pixels

// Simple hash function — fast, decent quality
func hash(p vec2) float {
    h := dot(p, vec2(127.1, 311.7))
    return fract(sin(h) * 43758.5453)
}

// Value noise with interpolation
func noise(p vec2) float {
    i := floor(p)
    f := fract(p)

    // Smooth interpolation
    u := f * f * (3.0 - 2.0*f)

    a := hash(i)
    b := hash(i + vec2(1, 0))
    c := hash(i + vec2(0, 1))
    d := hash(i + vec2(1, 1))

    return mix(mix(a, b, u.x), mix(c, d, u.x), u.y)
}

// Fractal Brownian Motion — layered noise
func fbm(p vec2) float {
    value := 0.0
    amplitude := 0.5
    pos := p
    for i := 0; i < 5; i++ {
        value += amplitude * noise(pos)
        pos *= 2.0
        amplitude *= 0.5
    }
    return value
}

var Time float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    pos := (dstPos.xy - imageDstOrigin()) * 0.01
    n := fbm(pos + Time*0.2)
    return vec4(n, n, n, 1)
}
```

---

## Screen Shake Distortion

Add per-pixel jitter during screen shake for extra impact.

```
//kage:unit pixels

var ShakeAmount float  // 0 = none, increases with shake intensity
var Time float

func hash(p vec2) float {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    jitter := vec2(
        hash(dstPos.xy + Time) - 0.5,
        hash(dstPos.xy + Time + 100) - 0.5,
    ) * ShakeAmount
    return imageSrc0At(srcPos + jitter)
}
```

---

## Palette Swap

Remap colors using a palette lookup.

```
//kage:unit pixels

var Palette [8]vec4    // Up to 8 target colors
var PaletteSize int    // How many palette entries are active

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    if clr.a < 0.01 {
        return clr
    }

    // Use grayscale value to index into palette
    gray := dot(clr.rgb / clr.a, vec3(0.299, 0.587, 0.114))
    idx := int(gray * float(PaletteSize - 1) + 0.5)
    idx = clamp(idx, 0, PaletteSize - 1)

    result := Palette[idx]
    return vec4(result.rgb * clr.a, clr.a)
}
```

---

## Silhouette

Render sprite as a solid color (for shadows, selections, etc.)

```
//kage:unit pixels

var SilhouetteColor vec4  // e.g., vec4(0, 0, 0, 0.5) for shadow

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0UnsafeAt(srcPos)
    if clr.a < 0.01 {
        return vec4(0)
    }
    return SilhouetteColor * clr.a
}
```

---

## Go Integration Patterns

### Embedding Shader Source

```go
import _ "embed"

//go:embed shaders/dissolve.kage
var dissolveShaderSrc []byte

func init() {
    var err error
    dissolveShader, err = ebiten.NewShader(dissolveShaderSrc)
    if err != nil {
        log.Fatal(err)
    }
}
```

### DrawRectShader (most common)

```go
func (g *Game) Draw(screen *ebiten.Image) {
    w, h := g.sprite.Bounds().Dx(), g.sprite.Bounds().Dy()

    op := &ebiten.DrawRectShaderOptions{}
    op.Images[0] = g.sprite
    op.Uniforms = map[string]any{
        "Time":      float32(g.tick) / 60.0,
        "Threshold": float32(g.dissolveProgress),
    }
    op.GeoM.Translate(g.x, g.y)

    screen.DrawRectShader(w, h, dissolveShader, op)
}
```

### Multiple Source Images

```go
op := &ebiten.DrawRectShaderOptions{}
op.Images[0] = sprite        // imageSrc0 in shader
op.Images[1] = normalMap     // imageSrc1 in shader
op.Images[2] = noiseTexture  // imageSrc2 in shader
// op.Images[3] = ...        // imageSrc3 in shader
```

All images should be the same size as the rect being drawn.

### Two-Pass Effects (e.g., Blur)

```go
// Allocate offscreen once
if g.blurBuffer == nil {
    g.blurBuffer = ebiten.NewImage(screenW, screenH)
}

// Pass 1: source → offscreen (horizontal)
g.blurBuffer.Clear()
op1 := &ebiten.DrawRectShaderOptions{}
op1.Images[0] = source
op1.Uniforms = map[string]any{
    "Direction": []float32{1, 0},
    "Radius":    float32(4),
}
g.blurBuffer.DrawRectShader(w, h, blurShader, op1)

// Pass 2: offscreen → screen (vertical)
op2 := &ebiten.DrawRectShaderOptions{}
op2.Images[0] = g.blurBuffer
op2.Uniforms = map[string]any{
    "Direction": []float32{0, 1},
    "Radius":    float32(4),
}
screen.DrawRectShader(w, h, blurShader, op2)
```

### Animated Uniforms

```go
func (g *Game) Update() error {
    g.tick++
    // Animate dissolve
    g.dissolveProgress += 0.005
    if g.dissolveProgress > 1 {
        g.dissolveProgress = 0
    }
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    op.Uniforms = map[string]any{
        "Time":      float32(g.tick) / float32(ebiten.TPS()),
        "Threshold": float32(g.dissolveProgress),
    }
}
```

### Error Handling on Shader Compile

```go
shader, err := ebiten.NewShader(src)
if err != nil {
    // err contains line numbers and descriptions
    // Common errors:
    //   - "unexpected token" → syntax error
    //   - "undeclared variable" → typo in variable name
    //   - "type mismatch" → wrong argument types
    //   - "cannot use ... as ..." → type casting needed
    log.Fatalf("shader compile error:\n%v", err)
}
```
