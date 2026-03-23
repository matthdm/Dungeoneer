# Kage Language Specification

Complete reference for Kage, Ebitengine's shader language.

## Table of Contents

1. [Unit Mode](#unit-mode)
2. [Data Types](#data-types)
3. [Type Construction & Casting](#type-construction--casting)
4. [Swizzling](#swizzling)
5. [Operators](#operators)
6. [Control Flow](#control-flow)
7. [Functions](#functions)
8. [Entry Point Signatures](#entry-point-signatures)
9. [Uniforms](#uniforms)
10. [Texture Sampling Functions](#texture-sampling-functions)
11. [Built-in Math Functions](#built-in-math-functions)
12. [Built-in Vector/Matrix Functions](#built-in-vectormatrix-functions)
13. [Special Functions](#special-functions)
14. [Arrays](#arrays)
15. [Limitations vs Go](#limitations-vs-go)
16. [Uniform Type Mapping (Go ↔ Kage)](#uniform-type-mapping)

---

## Unit Mode

Declared at the top of the file:

```
//kage:unit pixels   // RECOMMENDED — all coordinates in pixels
//kage:unit texels   // Legacy — normalized 0-1 coordinates
```

In pixel mode, `srcPos`, image origins, and sizes are all in pixels. In texel mode, texture coordinates are normalized to the texture's dimensions. Always use pixel mode for new shaders.

---

## Data Types

### Scalars
| Kage Type | Description |
|-----------|-------------|
| `bool` | Boolean |
| `int` | Integer |
| `float` | Floating-point (no precision specifier) |

### Vectors (float)
| Type | Components |
|------|------------|
| `vec2` | 2 floats |
| `vec3` | 3 floats |
| `vec4` | 4 floats (commonly RGBA color or position) |

### Vectors (int)
| Type | Components |
|------|------------|
| `ivec2` | 2 ints |
| `ivec3` | 3 ints |
| `ivec4` | 4 ints |

### Matrices (float, column-major)
| Type | Dimensions |
|------|------------|
| `mat2` | 2×2 |
| `mat3` | 3×3 |
| `mat4` | 4×4 |

---

## Type Construction & Casting

Types double as constructors. Arguments can be scalars, smaller vectors, or a mix:

```
// Scalar fill
vec4(0)                // vec4(0, 0, 0, 0)
mat4(1)                // Identity matrix (1 on diagonal)

// Explicit components
vec2(1.0, 2.0)
vec3(0.5, 0.5, 1.0)
vec4(1, 2, 3, 4)

// Mixed — components unpacked left to right
vec4(vec2(1, 2), 3, 4)      // vec4(1, 2, 3, 4)
vec4(1, vec3(2, 3, 4))      // vec4(1, 2, 3, 4)
vec4(vec2(1, 2), vec2(3, 4)) // vec4(1, 2, 3, 4)

// Matrix from column vectors
mat4(col0, col1, col2, col3)  // Each col is a vec4

// Type casting
float(myInt)
int(myFloat)
vec2(myIvec2)
```

---

## Swizzling

Access vector components by name. Three interchangeable naming groups:

| Group | Components |
|-------|------------|
| Position | `x`, `y`, `z`, `w` |
| Color | `r`, `g`, `b`, `a` |
| Texture | `s`, `t`, `p`, `q` |

Rules:
- All components in a single swizzle must use the same group
- Can read 1–4 components: `v.x`, `v.xy`, `v.xyz`, `v.xyzw`
- Can repeat: `v.xxx`, `v.rrgg`
- Can reorder: `v.bgra`, `v.wzyx`
- Can write to swizzle targets: `v.xy = vec2(1, 2)`
- Cannot write with repeated components: `v.xx = vec2(1, 2)` is INVALID

```
clr := vec4(1, 0.5, 0.2, 1)
clr.rgb     // vec3(1, 0.5, 0.2)
clr.a       // 1.0
clr.bgr     // vec3(0.2, 0.5, 1)
clr.rrr     // vec3(1, 1, 1)
```

---

## Operators

### Arithmetic
`+`, `-`, `*`, `/`, `%` (mod, int only)

Vector-scalar operations apply component-wise:
```
vec4(1, 2, 3, 4) * 2    // vec4(2, 4, 6, 8)
vec4(1, 2, 3, 4) + vec4(1, 1, 1, 1)  // vec4(2, 3, 4, 5)
```

Matrix-vector multiplication:
```
mat4(...) * vec4(...)    // Standard matrix-vector product
```

### Comparison
`==`, `!=`, `<`, `>`, `<=`, `>=`

### Logical
`&&`, `||`, `!`

### Assignment
`=`, `+=`, `-=`, `*=`, `/=`

---

## Control Flow

### Supported
```
if condition { ... }
if condition { ... } else { ... }
if condition { ... } else if condition { ... }

for i := 0; i < N; i++ { ... }   // C-style for loop
for condition { ... }              // While-style
```

### NOT Supported
- `switch` / `case`
- `goto`
- `for range`
- Labels / `break` to label / `continue` to label

### `break` and `continue`
Standard `break` and `continue` work in for loops.

---

## Functions

User-defined functions follow Go syntax:

```
func myHelper(a vec2, b float) vec4 {
    return vec4(a, b, 1)
}
```

- No method definitions (no receiver)
- No closures or function types
- No variadic functions (except some built-ins like `min`/`max` in v2.9+)
- No recursion (shaders are inlined)
- Multiple return values supported: `func foo() (float, float) { return 1, 2 }`

---

## Entry Point Signatures

The shader must define exactly one function named `Fragment`:

```
func Fragment() vec4
func Fragment(dstPos vec4) vec4
func Fragment(dstPos vec4, srcPos vec2) vec4
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4
func Fragment(dstPos vec4, srcPos vec2, color vec4, custom vec4) vec4
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `dstPos` | `vec4` | Destination pixel. `.xy` = position, `.z` = 0, `.w` = 1 |
| `srcPos` | `vec2` | Source texture coordinate for image 0 |
| `color` | `vec4` | Vertex color from `ColorScale` (0–1 range) |
| `custom` | `vec4` | Custom per-vertex data |
| Return | `vec4` | Output color (premultiplied alpha, 0–1 range) |

---

## Uniforms

Exported (capitalized) global variables become uniforms. They are read-only within the shader and set from Go.

```
var Time float
var Resolution vec2
var Palette [8]vec4      // Arrays are supported
var Transform mat4
```

**Only uniform globals are allowed.** No unexported globals, no non-uniform globals.

---

## Texture Sampling Functions

For each source image N (0–3):

| Function | Returns | Description |
|----------|---------|-------------|
| `imageSrcNAt(pos vec2)` | `vec4` | Safe — transparent black outside bounds |
| `imageSrcNUnsafeAt(pos vec2)` | `vec4` | Fast — undefined outside bounds |
| `imageSrcNOrigin()` | `vec2` | Top-left corner of source region |
| `imageSrcNSize()` | `vec2` | Size of source region |

Replace `N` with `0`, `1`, `2`, or `3`.

Destination info:
| Function | Returns | Description |
|----------|---------|-------------|
| `imageDstOrigin()` | `vec2` | Top-left of destination region |
| `imageDstSize()` | `vec2` | Size of destination region |

**Note:** In texel mode, coordinates and sizes are normalized. In pixel mode, they're in pixels. Always use pixel mode.

---

## Built-in Math Functions

All operate on `float`, `vec2`, `vec3`, or `vec4` (component-wise for vectors).

### Trigonometry
| Function | Description |
|----------|-------------|
| `sin(x)` | Sine |
| `cos(x)` | Cosine |
| `tan(x)` | Tangent |
| `asin(x)` | Arcsine |
| `acos(x)` | Arccosine |
| `atan(y_over_x)` | Arctangent (single arg) |
| `atan2(y, x)` | Arctangent (two args) |

### Exponential
| Function | Description |
|----------|-------------|
| `pow(x, y)` | x^y |
| `exp(x)` | e^x |
| `exp2(x)` | 2^x |
| `log(x)` | Natural log |
| `log2(x)` | Base-2 log |
| `sqrt(x)` | Square root |
| `inversesqrt(x)` | 1/√x |

### Common
| Function | Description |
|----------|-------------|
| `abs(x)` | Absolute value |
| `sign(x)` | -1, 0, or 1 |
| `floor(x)` | Round down |
| `ceil(x)` | Round up |
| `fract(x)` | Fractional part (x - floor(x)) |
| `mod(x, y)` | Modulo |
| `min(x, y, ...)` | Minimum (variadic in v2.9+) |
| `max(x, y, ...)` | Maximum (variadic in v2.9+) |
| `clamp(x, low, high)` | Constrain to range |

### Interpolation
| Function | Description |
|----------|-------------|
| `mix(x, y, a)` | Linear interpolation: x·(1-a) + y·a |
| `step(edge, x)` | 0 if x < edge, else 1 |
| `smoothstep(e0, e1, x)` | Hermite interpolation, 0–1 curve |

---

## Built-in Vector/Matrix Functions

| Function | Input | Description |
|----------|-------|-------------|
| `length(v)` | vec | Magnitude |
| `distance(a, b)` | vec | Euclidean distance |
| `dot(a, b)` | vec | Dot product |
| `cross(a, b)` | vec3 | Cross product |
| `normalize(v)` | vec | Unit vector |
| `faceforward(n, i, nref)` | vec | Flip n if dot(nref, i) < 0 |
| `reflect(i, n)` | vec | Reflection of i around n |
| `refract(i, n, eta)` | vec | Refraction (v2.4+) |
| `transpose(m)` | mat | Matrix transpose |

---

## Special Functions

| Function | Description |
|----------|-------------|
| `discard()` | Skip this pixel (output nothing). v2.4+ |
| `dfdx(p)` | Partial derivative in x (nondeterministic) |
| `dfdy(p)` | Partial derivative in y (nondeterministic) |
| `fwidth(p)` | abs(dfdx(p)) + abs(dfdy(p)) (nondeterministic) |
| `frontfacing()` | Returns bool — is this a front-facing fragment? v2.9+ |
| `len(arr)` | Array length |
| `cap(arr)` | Array capacity (v2.1+) |

---

## Arrays

Fixed-size arrays are supported. Declared with Go syntax:

```
samples := [...]float{-8, -4, -2, 2, 4, 8}
palette := [4]vec4{
    vec4(0, 0, 0, 1),
    vec4(1, 0, 0, 1),
    vec4(0, 1, 0, 1),
    vec4(0, 0, 1, 1),
}
```

Access with `arr[i]`. Use `len(arr)` for length. Arrays can be uniform:

```
var Colors [8]vec4  // Set from Go as []float32 of length 8*4=32
```

---

## Limitations vs Go

| Go Feature | Kage Support |
|------------|-------------|
| `float32`, `float64` | NO — use `float` |
| `rune`, `string` | NO |
| `interface{}` | NO |
| Slices | NO — use fixed arrays |
| Pointers | NO |
| Structs | NO |
| Maps | NO |
| Channels | NO |
| Goroutines | NO |
| `defer` | NO |
| `switch` / `case` | NO — use if/else chains |
| `goto` | NO |
| `for range` | NO — use C-style for |
| `import` | NO |
| `type` definitions | NO |
| `new`, `make`, `panic` | NO |
| `init` functions | NO |
| Methods (receivers) | NO |
| Multiple packages | NO |
| Recursion | NO |

---

## Uniform Type Mapping

How Go types map to Kage types when setting uniforms:

| Kage Type | Go Type in Uniforms Map |
|-----------|------------------------|
| `float` | `float32` |
| `int` | `int` |
| `vec2` | `[]float32{x, y}` |
| `vec3` | `[]float32{x, y, z}` |
| `vec4` | `[]float32{r, g, b, a}` |
| `ivec2` | `[]int{x, y}` |
| `mat2` | `[]float32` (4 elements, column-major) |
| `mat3` | `[]float32` (9 elements, column-major) |
| `mat4` | `[]float32` (16 elements, column-major) |
| `[N]float` | `[]float32` (N elements) |
| `[N]vec4` | `[]float32` (N*4 elements) |

**Common mistake:** Using `float64` instead of `float32` — this will silently produce wrong values or panic.
