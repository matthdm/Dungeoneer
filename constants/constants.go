package constants

const (
	DEBUG_TEMPLATE = `DEBUG LOG
CAM  WASD 
ZM   EC 
LVL  R T Q
FPS  %0.0f
TPS  %0.0f
SCA  %0.2f
POS  %0.0f,%0.0f
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
