package leveleditor

import "dungeoneer/levels"

// Generate64x64 replaces the current level with a procedurally generated one.
// TODO: expose parameters via editor UI.
func (e *Editor) Generate64x64(p levels.GenParams) {
	lvl := levels.Generate64x64(p)
	if e.layered != nil {
		e.layered.Layers[e.layerIndex] = lvl
		e.level = lvl
		if e.OnLayerChange != nil {
			e.OnLayerChange(lvl)
		}
	} else {
		e.level = lvl
		if e.OnLayerChange != nil {
			e.OnLayerChange(lvl)
		}
	}
}
