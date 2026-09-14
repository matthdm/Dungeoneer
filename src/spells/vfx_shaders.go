package spells

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// This file holds the shared Kage-shader VFX primitives new spell/artifact
// visuals build on: a radial glow (additive light/energy) and a procedural
// collapsing-rift pattern (swirling core + bright rim). Both are compiled
// once at package load — never per-frame — and driven by small Go wrappers
// that follow the same screen-space-quad convention as every other spell
// Draw method in this package.

// glowShaderSrc renders a soft additive radial glow: bright core fading to
// transparent at Radius, with a slow pulse driven by Time.
const glowShaderSrc = `//kage:unit pixels
package main

var Center vec2
var Radius float
var Time float
var Col vec3
var Intensity float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	dist := distance(dstPos.xy, Center)
	pulse := 0.85 + 0.15*sin(Time*1.6)
	falloff := 1.0 - smoothstep(0.0, Radius*pulse, dist)
	falloff *= falloff
	a := clamp(falloff*Intensity, 0.0, 1.0)
	return vec4(Col*a, a)
}
`

// riftShaderSrc procedurally synthesizes a collapsing-rift pattern: a dark
// swirling core (angle-based sine swirl, animated by Time) inside Radius*0.55,
// with a bright rim band between Radius*0.55 and Radius*0.85. Self-contained —
// it does not sample the background, so it needs no offscreen render target.
const riftShaderSrc = `//kage:unit pixels
package main

var Center vec2
var Radius float
var Time float
var CoreCol vec3
var RimCol vec3

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	offset := dstPos.xy - Center
	dist := length(offset)
	angle := atan2(offset.y, offset.x)
	mask := 1.0 - smoothstep(Radius*0.9, Radius, dist)

	swirl := sin(angle*6.0+Time*4.0-dist*0.3)*0.5 + 0.5
	core := 1.0 - smoothstep(0.0, Radius*0.55, dist)
	core = pow(core, 1.6)

	rimOuter := 1.0 - smoothstep(Radius*0.7, Radius*0.85, dist)
	rimInner := 1.0 - smoothstep(Radius*0.55, Radius*0.7, dist)
	rim := clamp(rimOuter-rimInner, 0.0, 1.0)

	col := CoreCol*core*(0.6+0.4*swirl) + RimCol*rim
	a := clamp(core+rim, 0.0, 1.0) * mask
	return vec4(col*a, a)
}
`

var (
	glowShader *ebiten.Shader
	riftShader *ebiten.Shader
)

func init() {
	var err error
	glowShader, err = ebiten.NewShader([]byte(glowShaderSrc))
	if err != nil {
		panic("spells: failed to compile glowShaderSrc: " + err.Error())
	}
	riftShader, err = ebiten.NewShader([]byte(riftShaderSrc))
	if err != nil {
		panic("spells: failed to compile riftShaderSrc: " + err.Error())
	}
}

// drawShaderQuad draws a single shader-filled quad covering a square of side
// 2*halfSize centered at (sx,sy) in screen pixels. Vertex colors are left at
// white/opaque — the shaders compute their own color and alpha, ignoring the
// vertex color input.
func drawShaderQuad(screen *ebiten.Image, shader *ebiten.Shader, sx, sy, halfSize float32, uniforms map[string]interface{}, blend ebiten.Blend) {
	x0, y0 := sx-halfSize, sy-halfSize
	x1, y1 := sx+halfSize, sy+halfSize
	vs := []ebiten.Vertex{
		{DstX: x0, DstY: y0, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x1, DstY: y0, SrcX: 1, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x1, DstY: y1, SrcX: 1, SrcY: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x0, DstY: y1, SrcX: 0, SrcY: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	is := []uint16{0, 1, 2, 0, 2, 3}
	op := &ebiten.DrawTrianglesShaderOptions{Uniforms: uniforms, Blend: blend}
	screen.DrawTrianglesShader(vs, is, shader, op)
}

// drawGlow renders an additive radial glow of the given screen-pixel radius
// centered at (sx,sy). col is linear RGB in [0,1]; intensity scales opacity.
func drawGlow(screen *ebiten.Image, sx, sy, radius float32, col [3]float32, intensity float32, t float64) {
	drawShaderQuad(screen, glowShader, sx, sy, radius*1.3, map[string]interface{}{
		"Center":    []float32{sx, sy},
		"Radius":    radius,
		"Time":      float32(t),
		"Col":       col[:],
		"Intensity": intensity,
	}, ebiten.BlendLighter)
}

// drawRift renders the procedural collapsing-rift pattern (see riftShaderSrc)
// of the given screen-pixel radius centered at (sx,sy).
func drawRift(screen *ebiten.Image, sx, sy, radius float32, coreCol, rimCol [3]float32, t float64) {
	drawShaderQuad(screen, riftShader, sx, sy, radius*1.1, map[string]interface{}{
		"Center":  []float32{sx, sy},
		"Radius":  radius,
		"Time":    float32(t),
		"CoreCol": coreCol[:],
		"RimCol":  rimCol[:],
	}, ebiten.BlendSourceOver)
}
