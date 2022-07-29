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
			color, _, _ := style.Decompose()
			data += fmt.Sprintf("%v,%v|%v|%v\n", x, y, color.Hex(), string(character))
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
		color, err := strconv.Atoi(segments[1])
		if err != nil {
			panic("invalid color")
		}
		character := []rune(segments[2])[0]
		screen.SetContent(x, y, character, nil, tcell.StyleDefault.Foreground(tcell.NewHexColor(int32(color))))
	}
}
