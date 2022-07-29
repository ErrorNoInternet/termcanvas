package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func dumpData(screen tcell.Screen) string {
	data := ""
	width, height := screen.Size()
	for x := 0; x <= width; x++ {
		for y := 0; y <= height; y++ {
			character, _, style, _ := screen.GetContent(x, y)
			color, _, _ := style.Decompose()
			data += fmt.Sprintf("%v,%v|%v|%v\n", x, y, color.Hex(), string(character))
		}
	}
	return data
}
