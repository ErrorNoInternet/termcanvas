package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func dumpData(screen tcell.Screen) string {
	data := ""
	width, height := screen.Size()
	for x := 0; x <= width; x++ {
		for y := 4; y <= height; y++ {
			character, _, style, _ := screen.GetContent(x, y)
			foregroundColor, backgroundColor, _ := style.Decompose()
			data += fmt.Sprintf("%v,%v|%v|%v|%v\n", x, y, foregroundColor.Hex(), backgroundColor.Hex(), string(character))
		}
	}
	return data
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
		foregroundColor, err := strconv.Atoi(segments[1])
		if err != nil {
			panic("invalid color")
		}
		backgroundColor, err := strconv.Atoi(segments[2])
		if err != nil {
			panic("invalid color")
		}
		character := []rune(segments[3])[0]
		textColor := tcell.StyleDefault.
			Foreground(tcell.NewHexColor(int32(foregroundColor))).
			Background(tcell.NewHexColor(int32(backgroundColor)))
		screen.SetContent(x, y, character, nil, textColor)
	}
}
