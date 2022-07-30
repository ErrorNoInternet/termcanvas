package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func dumpData(screen tcell.Screen) (string, bool) {
	data := ""
	empty := true
	width, height := screen.Size()
	for x := 0; x <= width; x++ {
		for y := 4; y <= height; y++ {
			character, _, style, _ := screen.GetContent(x, y)
			if character != ' ' && character != 0 {
				empty = false
			}
			foregroundColor, backgroundColor, _ := style.Decompose()
			var foregroundColorName, backgroundColorName string
			for _, existingColor := range colors {
				if tcell.GetColor(existingColor) == foregroundColor {
					foregroundColorName = existingColor
				}
			}
			for _, existingColor := range colors {
				if tcell.GetColor(existingColor) == backgroundColor {
					backgroundColorName = existingColor
				}
			}
			if foregroundColorName == "" && backgroundColorName == "" {
				continue
			}
			data += fmt.Sprintf("%v,%v|%v|%v|%v\n", x, y, foregroundColorName, backgroundColorName, string(character))
		}
	}
	return data, empty
}

func readData(data string, screen tcell.Screen) {
	for _, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		segments := strings.Split(line, "|")
		x, err := strconv.Atoi(strings.Split(segments[0], ",")[0])
		if err != nil {
			panic("invalid x coordinate")
		}
		y, err := strconv.Atoi(strings.Split(segments[0], ",")[1])
		if err != nil {
			panic("invalid y coordinate")
		}
		foregroundColorName := segments[1]
		backgroundColorName := segments[2]
		character := []rune(segments[3])[0]
		textColor := tcell.StyleDefault.
			Foreground(tcell.GetColor(foregroundColorName)).
			Background(tcell.GetColor(backgroundColorName))
		screen.SetContent(x, y, character, nil, textColor)
	}
}
