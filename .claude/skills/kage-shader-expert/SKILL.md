---
name: kage-shader-expert
description: Expert-level guidance for writing Kage shaders — Ebitengine's custom fragment shader language. Use this skill whenever the user wants to write, debug, or optimize shaders for Ebitengine games. Trigger when the user mentions Kage, shader effects, visual effects, post-processing, fragment shaders, GLSL-like code for ebiten, screen effects (blur, glow, dissolve, outline, lighting, water, distortion, color grading, CRT, pixelation, etc.), DrawRectShader, DrawTrianglesShader, or asks how to achieve any visual effect in Ebitengine that goes beyond basic DrawImage. Also trigger when the user is porting GLSL/HLSL shaders to Kage or asking about Ebitengine's shader compilation, uniforms, or texture sampling.
---

# Kage Shader Expert

You are an expert in **Kage**, Ebitengine's custom shading language. Kage has Go-like syntax but is a specialized fragment shader language that cross-compiles at runtime to OpenGL, Metal, DirectX, and other backends. Your job is to help write correct, performant, and visually compelling Kage shaders.

## What Kage Is (and Isn't)

Kage is **not** GLSL, HLSL, or WGSL — it's Ebitengine's own language. Key differences from Go and other shader languages:

- **Fragment shaders only** — no vertex shaders, no compute shaders
- **Go-like syntax** but with GPU-oriented types (`vec2`, `vec3`, `vec4`, `mat2`, `mat3`, `mat4`, `ivec2`, etc.)
- **No structs, interfaces, slices, pointers, maps, channels, goroutines, defer, switch, goto, for-range, imports, or type definitions**
- **Only `int`, `float`, and `bool` scalar types** — no `float32`, `float64`, `rune`, `string`
- **Arrays are supported**, structs are not (yet)
- **Uniforms are exported globals** (capitalized names) — read-only, set from Go
- **No other globals allowed** besides uniforms
- **Entry point must be named `Fragment`**

Understanding these constraints prevents frustration. When a user wants something Kage can't do, explain why and suggest the workaround (often: do the computation in Go and pass results as uniforms or pre-rendered textures).

## How to Approach Shader Tasks

1. **Understand the visual goal.** Ask what effect the user wants to achieve. Many effects have well-known shader implementations — identify which pattern applies.

2. **Check the reference.** Read `references/kage-language-spec.md` for exact syntax, types, and built-in functions. Read `references/shader-recipes.md` for ready-made effect implementations.

3. **Start with `//kage:unit pixels`.** Always use pixel-mode for new shaders — it's clearer and recommended by the Ebitengine documentation. Only use texel mode if the user specifically needs it or is maintaining legacy code.

4. **Write the shader, then the Go integration.** Show both the Kage source and the Go code that compiles and uses it (`NewShader`, `DrawRectShader` or `DrawTrianglesShader`, uniform setup).

5. **Test mentally.** Walk through what happens for a few representative pixels. Shader bugs are hard to debug because there's no printf — careful reasoning about coordinates and math is the primary debugging tool.

6. **Optimize thoughtfully.** GPU shaders run per-pixel. Avoid unnecessary texture samples, minimize branching, and prefer math over conditionals. But don't micro-optimize at the cost of readability — Kage compiles to efficient backend code automatically.

## Reference Files

- **`references/kage-language-spec.md`** — Complete language specification: types, operators, swizzling, control flow, built-in functions, entry point signatures, uniform system, texture sampling, coordinate modes. Consult for any syntax or API question.

- **`references/shader-recipes.md`** — Ready-to-use shader implementations for common game effects: grayscale, flash/hit, outline, dissolve, chromatic aberration, radial blur, lighting/normal maps, water reflection, CRT scanlines, vignette, color grading, noise generation, and more. Also covers Go-side integration patterns. Consult when the user wants a specific visual effect.

## Key Concepts

### The Fragment Function

Every Kage shader must define a `Fragment` function. It runs once per pixel and returns the output color as `vec4` (RGBA, premultiplied alpha, 0.0–1.0 range).

```
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4
```

| Parameter | Meaning |
|-----------|---------|
| `dstPos` | Destination pixel position. `dstPos.xy` is the pixel coordinate. `.z` = 0, `.w` = 1. |
| `srcPos` | Source texture coordinate for the primary image (in pixels or texels depending on unit mode). |
| `color` | Vertex color from `ColorScale` — multiply your output by this for proper tinting. |
| Return | Output color as `vec4`. Components must be 0.0–1.0. Alpha is premultiplied. |

Shorter signatures are also valid if you don't need all parameters:
```
func Fragment() vec4
func Fragment(dstPos vec4) vec4
func Fragment(dstPos vec4, srcPos vec2) vec4
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4
```

### Coordinate Systems

**Always start with `//kage:unit pixels`** (recommended):
- All coordinates and sizes are in pixels
- `imageSrc0At(srcPos)` samples at a pixel position
- `imageDstSize()` returns size in pixels

Legacy `//kage:unit texels`:
- Coordinates are normalized 0.0–1.0 across the texture
- Requires manual conversion: `pixels = texels × textureSize`
- Confusing for most use cases

### Uniforms — Passing Data from Go

Uniforms are exported (capitalized) global variables. They're set from Go via the `Uniforms` map and are constant across all pixels in a single draw call.

**In Kage:**
```
var Time float
var Resolution vec2
var FlashColor vec4
var Enabled int  // Use int for bools (0 or 1)
```

**In Go:**
```go
op.Uniforms = map[string]any{
    "Time":       float32(elapsed),
    "Resolution": []float32{float32(w), float32(h)},
    "FlashColor": []float32{1, 1, 1, 1},
    "Enabled":    1,
}
```

**Important:** Go types must match. `float` → `float32`. `vec2/3/4` → `[]float32` of matching length. `int` → `int` (not `int32`). `mat4` → `[]float32` of length 16 (column-major).

### Texture Sampling

Access up to 4 source images (set via `options.Images[0..3]`):

```
imageSrc0At(pos vec2) vec4        // Safe — returns vec4(0) outside bounds
imageSrc0UnsafeAt(pos vec2) vec4  // Fast — undefined outside bounds

imageSrc0Origin() vec2            // Top-left corner of the source region
imageSrc0Size() vec2              // Width and height of the source
imageDstOrigin() vec2             // Top-left of the destination region
imageDstSize() vec2               // Width and height of the destination
```

Replace `0` with `1`, `2`, or `3` for additional images.

**When to use safe vs unsafe:** Use `UnsafeAt` when you know the coordinate is in-bounds (sampling at `srcPos` without offset). Use `At` when you're offsetting coordinates (blur, outline, distortion) since out-of-bounds access returns transparent black, which is usually what you want.

### Swizzling

Vector components can be accessed and rearranged using `.xyzw`, `.rgba`, or `.stpq`:

```
v := vec4(1, 2, 3, 4)
v.xy    // vec2(1, 2)
v.rgb   // vec3(1, 2, 3)
v.rrr   // vec3(1, 1, 1)  — repeat components
v.bgra  // vec4(3, 2, 1, 4) — reorder
v.xy = vec2(5, 6)  // Write to specific components
```

Components within a single swizzle must come from the same naming group (can't mix `.x` with `.r`).

### Type Construction

Types double as constructors with flexible arguments:
```
vec4(0)              // All zeros
vec4(1, 2, 3, 4)     // Explicit components
vec4(v2, 0, 1)       // Expand vec2 + scalars
vec4(v3, 1)           // Expand vec3 + scalar
mat4(1)               // Identity matrix (1 on diagonal)
```

### Common Patterns

**Normalizing position to 0–1 range:**
```
uv := (dstPos.xy - imageDstOrigin()) / imageDstSize()
```

**Distance-based effects:**
```
dist := distance(dstPos.xy - imageDstOrigin(), center)
t := clamp(dist / radius, 0, 1)
```

**Time-based animation:**
```
var Time float
// In Fragment:
wave := sin(Time * speed + pos.x * frequency) * amplitude
```

**Mixing/lerping between states:**
```
result := mix(colorA, colorB, t)  // t=0 → colorA, t=1 → colorB
```

**Discarding pixels:**
```
if alpha < 0.01 {
    discard()
}
```

## Common Mistakes

1. **Forgetting `//kage:unit pixels`** — Without this directive, you're in texel mode and coordinates are normalized, which causes confusing behavior.

2. **Using Go types** — Writing `float64` or `float32` instead of just `float`. Kage only has `float`.

3. **Trying to use structs** — Kage doesn't support them. Use multiple variables or vec4 to pack data.

4. **Modifying uniforms** — Uniforms are read-only. You can't assign to exported globals.

5. **Non-premultiplied alpha** — Kage works with premultiplied alpha. If your output color is `(r, g, b, a)`, then `r`, `g`, `b` must each be ≤ `a`. To premultiply: `return vec4(color.rgb * color.a, color.a)`.

6. **Wrong uniform types from Go** — A `float` uniform needs `float32` in Go, not `float64`. A `vec2` needs `[]float32{x, y}`, not `[2]float32{x, y}`.

7. **Sampling outside bounds without `At`** — Using `UnsafeAt` with offset coordinates can read garbage pixels. Use `At` (the safe version) whenever you offset from `srcPos`.

8. **Image size mismatch** — All images in `DrawRectShaderOptions.Images` should match the rect size. Mismatches cause unexpected sampling.

## Go Integration Pattern

The complete pattern for using a Kage shader in Ebitengine:

```go
// At init / load time:
shader, err := ebiten.NewShader([]byte(kageSource))
if err != nil {
    log.Fatal("shader compile error:", err)
}

// In Draw():
op := &ebiten.DrawRectShaderOptions{}
op.Images[0] = sourceImage
op.Uniforms = map[string]any{
    "Time": float32(g.tick) / 60.0,
}
// GeoM positions the shader output on screen
op.GeoM.Translate(x, y)

w, h := sourceImage.Bounds().Dx(), sourceImage.Bounds().Dy()
screen.DrawRectShader(w, h, shader, op)
```

For triangle-based rendering (more control over vertices):
```go
op := &ebiten.DrawTrianglesShaderOptions{}
op.Images[0] = sourceImage
op.Uniforms = map[string]any{"Time": float32(t)}
screen.DrawTrianglesShader(vertices, indices, shader, op)
```
