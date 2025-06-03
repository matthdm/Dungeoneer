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

const (
	// Floor tiles (always bottom layer)
	DepthFloorBase   = 0.0
	DepthFloorOffset = 1.0

	// Wall sprites (in between floors and entities)
	DepthWallBase   = 10000.0
	DepthWallOffset = 1.0

	// Entities (always topmost among things at same y)
	DepthEntityBase = 20000.0
)

const (
	DepthTile  = 0.0
	DepthWall  = 0.1
	DepthActor = 0.2
)
