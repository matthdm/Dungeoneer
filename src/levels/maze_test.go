package levels

import (
	"dungeoneer/sprites"
	"testing"
)

func TestMazeGeneration(t *testing.T) {
	ss, err := sprites.LoadSpriteSheet(64)
	if err != nil {
		t.Skip("spritesheet missing")
	}
	
	cfgs := []MazeConfig{
		{Width: 20, Height: 20, Tessellation: "ortho", Routing: "prim"},
		{Width: 20, Height: 20, Tessellation: "ortho", Routing: "eller"},
		{Width: 20, Height: 20, Tessellation: "fractal", Routing: "prim"}, // triggers fractalMaze
		{Width: 20, Height: 20, Tessellation: "ortho", Routing: "unicursal"},
		{Width: 20, Height: 20, Tessellation: "ortho", Routing: "braid"}, // triggers braidify
	}
	
	for _, cfg := range cfgs {
		l := GenerateMaze(cfg, ss)
		if l == nil {
			t.Errorf("failed to generate maze for routing %s", cfg.Routing)
		}
		if l.W != 20 || l.H != 20 {
			t.Errorf("maze size incorrect")
		}
	}
}
