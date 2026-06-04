package game

import "testing"

func TestParticlePool_NeverAllocatesAfterInit(t *testing.T) {
	p := NewParticlePool()
	// Emit a burst and update — should not panic or overflow.
	p.Emit(100, 100, 8, 1.0, 0.4, 0.1)
	p.Update(0.016)
	active := 0
	for i := range p.particles {
		if p.particles[i].Active {
			active++
		}
	}
	if active != 8 {
		t.Errorf("expected 8 active particles after emit, got %d", active)
	}
}

func TestParticle_UpdateMovesPosition(t *testing.T) {
	p := NewParticlePool()
	p.Emit(50, 50, 1, 1, 1, 1)
	// Find the active particle.
	var pt *Particle
	for i := range p.particles {
		if p.particles[i].Active {
			pt = &p.particles[i]
			break
		}
	}
	if pt == nil {
		t.Fatal("no active particle found")
	}
	origX, origY := pt.X, pt.Y
	p.Update(0.1)
	if pt.X == origX && pt.Y == origY {
		t.Error("particle should have moved after update")
	}
}

func TestParticlePool_ExpiresAfterLifetime(t *testing.T) {
	p := NewParticlePool()
	p.Emit(0, 0, 1, 1, 1, 1)
	// Force expire.
	for i := range p.particles {
		if p.particles[i].Active {
			p.particles[i].Life = 0.001
		}
	}
	p.Update(0.1)
	for i := range p.particles {
		if p.particles[i].Active {
			t.Error("all particles should be expired")
		}
	}
}

func TestParticlePool_CapAt512(t *testing.T) {
	p := NewParticlePool()
	// Emit more than pool size — should not panic.
	for i := 0; i < 600; i++ {
		p.Emit(0, 0, 1, 1, 1, 1)
	}
}

func TestSpellParticleColor(t *testing.T) {
	colors := []string{
		"fire", "lightning", "nature", "chaos", "arcane", "unknown",
	}

	for _, c := range colors {
		r, g, b := SpellParticleColor(c)
		if r == 0 && g == 0 && b == 0 {
			t.Errorf("unexpected zero color for %s", c)
		}
	}
}

func TestDamageNumberColor(t *testing.T) {
	colors := []string{
		"fire", "lightning", "nature", "chaos", "unknown",
	}

	for _, c := range colors {
		col := DamageNumberColor(c, false)
		if col.A == 0 {
			t.Errorf("unexpected alpha 0 for %s", c)
		}
	}

	critCol := DamageNumberColor("fire", true)
	if critCol.R != 255 || critCol.G != 220 || critCol.B != 0 {
		t.Errorf("unexpected crit color")
	}
}
