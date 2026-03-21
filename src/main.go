package main

import (
	"dungeoneer/game"
	"dungeoneer/images"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("Dungeoneer")
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	images.SetDefaultWindowIcon()

	g, err := game.NewGame()
	if err != nil {
		log.Fatal(err)
	}

	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
