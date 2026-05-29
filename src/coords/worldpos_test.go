package coords

import "testing"

func TestWorldPosCoordinateConversions(t *testing.T) {
	pos := WorldPos{X: 3.5, Y: 7.25}

	if got := pos.TileX(); got != 3 {
		t.Fatalf("TileX() = %d, want 3", got)
	}
	if got := pos.TileY(); got != 7 {
		t.Fatalf("TileY() = %d, want 7", got)
	}

	body := pos.BodyCenter()
	if body.X != 4.5 || body.Y != 7.55 {
		t.Fatalf("BodyCenter() = {%g, %g}, want {4.5, 7.55}", body.X, body.Y)
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
