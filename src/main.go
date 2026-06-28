package main

import (
	"dungeoneer/game"
	"dungeoneer/images"
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	testLevel := flag.String("test-level", "", "Load a specific level for testing")
	screenshotFile := flag.String("screenshot", "", "File path to save the first frame as a PNG and exit")
	flag.Parse()

	ebiten.SetWindowTitle("Dungeoneer")
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	images.SetDefaultWindowIcon()

	g, err := game.NewGame(*testLevel, *screenshotFile)
	if err != nil {
		log.Fatal(err)
	}

	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
