package audio

import (
	"fmt"
	"sync"

	ebaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100
const maxConcurrent = 8

// Category identifies a group of sounds for volume control.
type Category int

const (
	CategorySFX     Category = iota
	CategoryMusic
	CategoryAmbient
)

// Engine manages all audio playback.
type Engine struct {
	mu        sync.Mutex
	context   *ebaudio.Context
	sfxVol    float64
	musicVol  float64
	ambVol    float64
	activeSFX int // count of currently playing SFX (mock)
}

// NewEngine creates an audio engine. Returns nil if audio context creation fails.
func NewEngine() (e *Engine) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("audio: context creation panicked (%v) — running silent\n", r)
			e = nil
		}
	}()
	ctx := ebaudio.NewContext(sampleRate)
	if ctx == nil {
		fmt.Println("audio: context creation failed — running silent")
		return nil
	}
	return &Engine{
		context:  ctx,
		sfxVol:   1.0,
		musicVol: 0.8,
		ambVol:   0.5,
	}
}

// SetVolume adjusts volume for a category (0.0–1.0).
func (e *Engine) SetVolume(cat Category, vol float64) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	switch cat {
	case CategorySFX:
		e.sfxVol = vol
	case CategoryMusic:
		e.musicVol = vol
	case CategoryAmbient:
		e.ambVol = vol
	}
}

// PlaySFX plays a sound effect from the given asset path.
// No-ops silently if the file doesn't exist or the engine is nil.
func (e *Engine) PlaySFX(path string) {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.activeSFX >= maxConcurrent {
		return // drop excess
	}
	// Asset loading is deferred until real .ogg files exist.
	// TODO: load and play e.context.NewPlayer(ogg.Decode(file))
	_ = path
}

// PlayMusic starts looping background music, replacing any current track.
// No-ops if file doesn't exist or engine is nil.
func (e *Engine) PlayMusic(path string) {
	if e == nil {
		return
	}
	// TODO: load OGG, create looping player, fade in.
	_ = path
}

// StopMusic fades out and stops the current music track.
func (e *Engine) StopMusic() {
	if e == nil {
		return
	}
	// TODO: fade out current player.
}

// PlayAmbient starts a looping ambient sound (biome loop).
func (e *Engine) PlayAmbient(path string) {
	if e == nil {
		return
	}
	// TODO: crossfade from current ambient to new path.
	_ = path
}

// StopAmbient stops ambient playback.
func (e *Engine) StopAmbient() {
	if e == nil {
		return
	}
}

// SFX path constants — asset paths relative to working dir (assets not yet created).
const (
	SFXHit            = "assets/audio/sfx/hit.ogg"
	SFXMiss           = "assets/audio/sfx/miss.ogg"
	SFXSpellFireball  = "assets/audio/sfx/spell_fireball.ogg"
	SFXSpellLightning = "assets/audio/sfx/spell_lightning.ogg"
	SFXSpellChaos     = "assets/audio/sfx/spell_chaos.ogg"
	SFXSpellGeneric   = "assets/audio/sfx/spell_generic.ogg"
	SFXEnemyDeath     = "assets/audio/sfx/enemy_death.ogg"
	SFXPlayerDamage   = "assets/audio/sfx/player_damage.ogg"
	SFXMenuOpen       = "assets/audio/sfx/menu_open.ogg"
	SFXPurchase       = "assets/audio/sfx/purchase.ogg"
	SFXTypewriter     = "assets/audio/sfx/typewriter.ogg"

	MusicBoss    = "assets/audio/music/boss.ogg"
	AmbientCrypt = "assets/audio/ambient/crypt.ogg"
	AmbientMoss  = "assets/audio/ambient/moss.ogg"
	AmbientGallery = "assets/audio/ambient/gallery.ogg"
	AmbientBrick = "assets/audio/ambient/brick.ogg"
	AmbientHub   = "assets/audio/ambient/hub.ogg"
)

// BiomeAmbient returns the ambient track path for a biome name.
func BiomeAmbient(biomeName string) string {
	switch biomeName {
	case "moss":
		return AmbientMoss
	case "gallery":
		return AmbientGallery
	case "brick":
		return AmbientBrick
	case "catacomb":
		return AmbientCrypt
	default:
		return AmbientCrypt
	}
}
