package entities

import "testing"

// Taunt takes hold on ordinary monsters, is ignored by bosses and echoes, and
// a shorter taunt never cuts a longer one short.
func TestMonsterTaunt(t *testing.T) {
	m := &Monster{HP: 10, Behavior: NewRangedBehavior(6)}
	if !m.Taunt(120) || m.TauntTicks != 120 {
		t.Fatalf("ranged monster not taunted: ticks=%d", m.TauntTicks)
	}
	if !m.Taunt(30) || m.TauntTicks != 120 {
		t.Fatalf("shorter taunt changed remaining time to %d, want 120", m.TauntTicks)
	}

	boss := &Monster{HP: 10, Behavior: NewBossBehavior(&Boss{})}
	if boss.Taunt(120) || boss.TauntTicks != 0 {
		t.Fatal("boss was taunted")
	}
	echo := &Monster{HP: 10, IsEcho: true}
	if echo.Taunt(120) || echo.TauntTicks != 0 {
		t.Fatal("echo was taunted")
	}
	dead := &Monster{IsDead: true}
	if dead.Taunt(120) {
		t.Fatal("dead monster was taunted")
	}
}
