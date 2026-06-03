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
