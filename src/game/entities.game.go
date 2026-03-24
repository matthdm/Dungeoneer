package game

import (
	"dungeoneer/entities"
	"dungeoneer/spells"

	"github.com/hajimehoshi/ebiten/v2"
)

// updateMonsterProjectiles drains pending projectiles from monsters into the
// game-level slice, updates each one, and checks for player collision.
func (g *Game) updateMonsterProjectiles() {
	// Drain pending projectiles from each monster.
	for _, m := range g.Monsters {
		if len(m.PendingProjectiles) > 0 {
			g.MonsterProjectiles = append(g.MonsterProjectiles, m.PendingProjectiles...)
			m.PendingProjectiles = m.PendingProjectiles[:0]
		}
		// Drain pending spell casts (caster monsters).
		for _, sc := range m.PendingSpells {
			g.spawnMonsterSpell(sc)
		}
		m.PendingSpells = m.PendingSpells[:0]
	}

	// Update and check collisions.
	alive := g.MonsterProjectiles[:0]
	for _, p := range g.MonsterProjectiles {
		p.Update(g.currentLevel)
		if !p.Finished && !g.player.IsDead && p.HitsPlayer(g.player.TileX, g.player.TileY) {
			g.player.TakeDamage(p.Damage)
			p.Finished = true
		}
		if !p.Finished {
			alive = append(alive, p)
		}
	}
	// Clear dangling pointers to help GC.
	for i := len(alive); i < len(g.MonsterProjectiles); i++ {
		g.MonsterProjectiles[i] = nil
	}
	g.MonsterProjectiles = alive
}

func (g *Game) drawMonsterProjectiles(target *ebiten.Image, scale, cx, cy float64) {
	tileSize := g.currentLevel.TileSize
	for _, p := range g.MonsterProjectiles {
		p.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
	}
}

// spawnMonsterSpell converts a pending monster spell cast into an active spell.
func (g *Game) spawnMonsterSpell(sc entities.PendingSpellCast) {
	switch sc.SpellName {
	case "fireball":
		info := spells.SpellInfo{Name: "fireball", Level: 1, Cooldown: 0, Damage: sc.Damage}
		fb := spells.NewFireball(info, sc.OriginX, sc.OriginY, sc.TargetX, sc.TargetY, g.fireballSprites, g.spriteSheet.FireBurst)
		fb.MonsterCast = true
		g.ActiveSpells = append(g.ActiveSpells, fb)
	}
}

func (g *Game) drawPlayer(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || g.player.IsDead {
		return
	}

	tileSize := g.currentLevel.TileSize
	g.player.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
}

func (g *Game) drawMonsters(target *ebiten.Image, scale, cx, cy float64) {
	tileSize := g.currentLevel.TileSize
	for _, m := range g.Monsters {
		if m.TileX < 0 || m.TileY < 0 || m.TileX >= g.currentLevel.W || m.TileY >= g.currentLevel.H {
			continue
		}
		if g.isTileVisible(m.TileX, m.TileY) {
			m.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
		} else if g.SeenTiles[m.TileY][m.TileX] {
			// optionally show faded sprite or placeholder
		} else {
			continue
		}

	}
}
