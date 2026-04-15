package main

import (
	"machine"

	"image/color"
	"time"

	"tinygo.org/x/drivers/unoqmatrix"
)

var (
	displayBuffer *DisplayBuffer
	lifegame      *LifeGame

	textWhite = color.RGBA{255, 255, 255, 255}
	textBlack = color.RGBA{0, 0, 0, 255}
)

func main() {
	display := unoqmatrix.NewFromBasePin(machine.PF0)
	display.ClearDisplay()
	displayBuffer = NewDisplayBuffer(display.Size())

	var err error
	lifegame, err = NewLifeGame(13, 8)
	if err != nil {
		return
	}
	lifegame.InitRandom()

	for {
		playLife()

		for y := int16(0); y < displayBuffer.height; y++ {
			for x := int16(0); x < displayBuffer.width; x++ {
				if displayBuffer.GetPixel(x, y) {
					display.SetPixel(x, y, textWhite)
				} else {
					display.SetPixel(x, y, textBlack)
				}
			}
		}
		display.Display()
		time.Sleep(5 * time.Millisecond)
	}
}

func playLife() {
	lifegame.Update()
	cells := lifegame.GetCells()
	for y := range cells {
		for x := range cells[y] {
			color := textBlack
			if cells[y][x] {
				color = textWhite
			}

			displayBuffer.SetPixel(int16(x), int16(y), color)
		}
	}
}
