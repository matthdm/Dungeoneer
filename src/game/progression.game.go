package game

import (
	"dungeoneer/entities"
	"dungeoneer/progression"
)

func (g *Game) awardEXP(m *entities.Monster) {
	if g.player == nil || m == nil {
		return
	}
	exp := progression.CalculateEXPReward(m.Level, g.player.Level)
	g.player.AddEXP(exp)
}
