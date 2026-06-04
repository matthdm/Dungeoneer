package game

import (
	"dungeoneer/sprites"
	"testing"
)

func TestBiomeEnemyByRole(t *testing.T) {
	bc := BiomeConfigs[BiomeCrypt]
	if bc == nil {
		t.Fatalf("expected crypt biome config")
	}

	enemy := bc.EnemyByRole("melee")
	if enemy == nil {
		t.Fatalf("expected melee enemy")
	}
	if enemy.ID != "crypt_melee" {
		t.Errorf("expected crypt_melee, got %s", enemy.ID)
	}

	enemyMissing := bc.EnemyByRole("non_existent_role")
	if enemyMissing != nil {
		t.Errorf("expected nil for missing role")
	}
}

func TestBuildSpriteMap(t *testing.T) {
	ss := &sprites.SpriteSheet{} // empty is fine for keys check
	m := BuildSpriteMap(ss)

	if len(m) == 0 {
		t.Errorf("expected non-empty sprite map")
	}

	if _, ok := m["GreyKnight"]; !ok {
		t.Errorf("expected GreyKnight in map")
	}
	if _, ok := m["Cyclops"]; !ok {
		t.Errorf("expected Cyclops in map")
	}
}
