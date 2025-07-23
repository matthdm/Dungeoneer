package constants

const (
	DEBUG_TEMPLATE = `DEBUG LOG
FPS      %0.0f
TPS      %0.0f
SCA      %0.2f
POS      %0.0f,%0.0f
CONTROLS F1
`
	DEBUG_BINDS_TEMPLATE = `BINDS
EDITOR  P
MOVE    WASD/M2
DASH    SHIFT
GRPPL   F
MENU    ESC
CAM     ARROWS
DRAG    M3
ZM      WHEEL+/-
SPELLS  1-6
LOS     9
FOW     0
ATK     M1
QSV/LD  F5/F6
LVL     N/M/ESC
`
)

// UI layout and scaling
const (
	MainMenuLabelScale       = 0.25
	MenuLabelHeightPixels    = 195
	MenuLabelVerticalPadding = 10
	MenuLabelOffsetX         = 35.0
)

// Animation timing and effects
const (
	GlowAlphaMin   = 0.3
	GlowAlphaRange = 0.7
)

// Visual effects
const (
	PathPreviewAlpha   = 0.4
	HostileTargetAlpha = 0.8
	ShadowAlpha        = 200
)

const (
	DefaultTileSize = 64
)

// Player dash mechanics
const (
	MaxDashCharges      = 5
	DashRecharge        = 5.0 // seconds per charge
	DashDuration        = 0.2 // seconds of dash movement
	DashSpeedMultiplier = 3.0 // multiple of normal move speed
)

// Grappling hook mechanics
const (
	GrappleMaxDistance = 12.0 // tiles
	// Increased speed for a snappier hook shot and pull
	GrappleSpeed = 30.0 // tiles per second for extension and pull
	GrappleDelay = 0.1  // delay before pulling starts
)

const (
	FireballHitRadius = 0.75
)

// Field-of-view rendering
const (
	// MaxFOVRays caps the number of rays cast when computing the
	// player's visible polygon. Large maps automatically reduce the
	// number of rays for performance.
	MaxFOVRays = 360
)

var DebugMode = true
