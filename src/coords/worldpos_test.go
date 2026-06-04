package coords

import (
	"math"
	"testing"
)

func TestWorldPosCoordinateConversions(t *testing.T) {
	pos := WorldPos{X: 3.5, Y: 7.25}

	if got := pos.TileX(); got != 3 {
		t.Fatalf("TileX() = %d, want 3", got)
	}
	if got := pos.TileY(); got != 7 {
		t.Fatalf("TileY() = %d, want 7", got)
	}

	body := pos.BodyCenter()
	if body.X != 4.75 || body.Y != 7.50 {
		t.Fatalf("BodyCenter() = {%g, %g}, want {4.75, 7.50}", body.X, body.Y)
	}

	isoX, isoY := pos.ToIso(64)
	if isoX != -120 || isoY != 172 {
		t.Fatalf("ToIso(64) = {%g, %g}, want {-120, 172}", isoX, isoY)
	}

	centerX, centerY := pos.TileCenterIso(64)
	if centerX != -88 || centerY != 188 {
		t.Fatalf("TileCenterIso(64) = {%g, %g}, want {-88, 188}", centerX, centerY)
	}

	renderX, renderY := pos.RenderIso(64, 0.25)
	if renderX != -120 || renderY != 171.25 {
		t.Fatalf("RenderIso(64, 0.25) = {%g, %g}, want {-120, 171.25}", renderX, renderY)
	}
}

func TestFromIsoRoundTrip(t *testing.T) {
	tests := []WorldPos{
		{X: 0, Y: 0},
		{X: 2, Y: 1},
		{X: 2.5, Y: 1.5},
		{X: 12.25, Y: 4.75},
	}

	for _, want := range tests {
		isoX, isoY := want.ToIso(64)
		got := FromIso(isoX, isoY, 64)
		if got.X != want.X || got.Y != want.Y {
			t.Fatalf("FromIso(ToIso(%v)) = %v, want %v", want, got, want)
		}
	}
}

func TestFromIsoMatchesReferenceExample(t *testing.T) {
	got := FromIso(64, 96, 128)
	if got.X != 2 || got.Y != 1 {
		t.Fatalf("FromIso(64, 96, 128) = {%g, %g}, want {2, 1}", got.X, got.Y)
	}
}

func TestWorldPosTileFloorsNegativeCoordinates(t *testing.T) {
	pos := WorldPos{X: -1.2, Y: -0.01}

	if got := pos.TileX(); got != -2 {
		t.Fatalf("TileX() = %d, want -2", got)
	}
	if got := pos.TileY(); got != -1 {
		t.Fatalf("TileY() = %d, want -1", got)
	}
}

func TestWorldPosDistTo(t *testing.T) {
	tests := []struct {
		name  string
		p1    WorldPos
		p2    WorldPos
		want  float64
	}{
		{
			name: "same position",
			p1:   WorldPos{X: 1.5, Y: 2.5},
			p2:   WorldPos{X: 1.5, Y: 2.5},
			want: 0.0,
		},
		{
			name: "horizontal distance",
			p1:   WorldPos{X: 1.0, Y: 0.0},
			p2:   WorldPos{X: 4.0, Y: 0.0},
			want: 3.0,
		},
		{
			name: "vertical distance",
			p1:   WorldPos{X: 0.0, Y: -1.0},
			p2:   WorldPos{X: 0.0, Y: 4.0},
			want: 5.0,
		},
		{
			name: "diagonal 3-4-5 triangle",
			p1:   WorldPos{X: 0.0, Y: 0.0},
			p2:   WorldPos{X: 3.0, Y: 4.0},
			want: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p1.DistTo(tt.p2)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("DistTo() = %g, want %g", got, tt.want)
			}
		})
	}
}

func TestPackageToIso(t *testing.T) {
	tests := []struct {
		x, y     float64
		tileSize int
		wantX    float64
		wantY    float64
	}{
		{x: 0, y: 0, tileSize: 64, wantX: 0, wantY: 0},
		{x: 3.5, y: 7.25, tileSize: 64, wantX: -120, wantY: 172},
		{x: 2, y: 1, tileSize: 128, wantX: 64, wantY: 96},
	}

	for _, tt := range tests {
		gotX, gotY := ToIso(tt.x, tt.y, tt.tileSize)
		if gotX != tt.wantX || gotY != tt.wantY {
			t.Errorf("ToIso(%g, %g, %d) = (%g, %g), want (%g, %g)",
				tt.x, tt.y, tt.tileSize, gotX, gotY, tt.wantX, tt.wantY)
		}
	}
}

