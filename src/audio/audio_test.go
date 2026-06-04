package audio

import "testing"

func TestEngine_NilSafeAllMethods(t *testing.T) {
	var e *Engine
	e.SetVolume(CategorySFX, 0.5)
	e.PlaySFX(SFXHit)
	e.PlayMusic(MusicBoss)
	e.StopMusic()
	e.PlayAmbient(AmbientCrypt)
	e.StopAmbient()
}

func TestEngine_SetVolume_StoresValue(t *testing.T) {
	e := &Engine{sfxVol: 1.0, musicVol: 1.0, ambVol: 1.0}
	e.SetVolume(CategorySFX, 0.5)
	if e.sfxVol != 0.5 {
		t.Errorf("sfxVol want 0.5, got %.2f", e.sfxVol)
	}
	e.SetVolume(CategoryMusic, 0.3)
	if e.musicVol != 0.3 {
		t.Errorf("musicVol want 0.3, got %.2f", e.musicVol)
	}
}

func TestBiomeAmbient_ReturnsKnownPaths(t *testing.T) {
	cases := []struct {
		biome string
		want  string
	}{
		{"moss", AmbientMoss},
		{"gallery", AmbientGallery},
		{"brick", AmbientBrick},
		{"crypt", AmbientCrypt},
		{"unknown", AmbientCrypt},
	}
	for _, c := range cases {
		got := BiomeAmbient(c.biome)
		if got != c.want {
			t.Errorf("BiomeAmbient(%q) want %q, got %q", c.biome, c.want, got)
		}
	}
}

func TestMaxConcurrentSFX(t *testing.T) {
	if maxConcurrent != 8 {
		t.Errorf("maxConcurrent want 8, got %d", maxConcurrent)
	}
}

func TestNewEngine(t *testing.T) {
	// ebiten/audio Context creation might fail without an actual audio device in tests,
	// but we should still call it to cover the function and its recovery.
	e := NewEngine()
	// Just verify it doesn't crash the test.
	_ = e
}

func TestEngine_SetVolume_Ambient(t *testing.T) {
	e := &Engine{ambVol: 1.0}
	e.SetVolume(CategoryAmbient, 0.4)
	if e.ambVol != 0.4 {
		t.Errorf("ambVol want 0.4, got %.2f", e.ambVol)
	}
}

func TestEngine_PlayMethods(t *testing.T) {
	e := &Engine{} // mock engine without context
	
	// Test normal play
	e.PlaySFX(SFXHit)
	
	// Test max concurrent
	e.activeSFX = maxConcurrent
	e.PlaySFX(SFXHit) // Should drop silently
	
	e.PlayMusic(MusicBoss)
	e.StopMusic()
	
	e.PlayAmbient(AmbientCrypt)
	e.StopAmbient()
}

func TestBiomeAmbient_Catacomb(t *testing.T) {
	got := BiomeAmbient("catacomb")
	if got != AmbientCrypt {
		t.Errorf("BiomeAmbient('catacomb') want %q, got %q", AmbientCrypt, got)
	}
}
