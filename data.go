package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func getColor(style tcell.Style) (string, string) {
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
		return "", ""
	}
	if foregroundColorName == "" {
		foregroundColorName = "reset"
	}
	if backgroundColorName == "" {
		backgroundColorName = "reset"
	}
	return foregroundColorName, backgroundColorName
}

func dumpData(screen tcell.Screen) (string, bool) {
	data := "x,y,foregroundColor,backgroundColor,character\n"
	empty := true
	width, height := screen.Size()
	for x := 0; x <= width; x++ {
		for y := 4; y <= height; y++ {
			character, _, style, _ := screen.GetContent(x, y)
			if character != ' ' && character != 0 {
				empty = false
			}
			foregroundColorName, backgroundColorName := getColor(style)
			if foregroundColorName == "" && backgroundColorName == "" {
				continue
			}
			data += fmt.Sprintf("%v,%v,%v,%v,%v\n", x, y, foregroundColorName, backgroundColorName, string(character))
		}
	}
	return data, empty
}

func drawData(data string, screen tcell.Screen) {
	for index, line := range strings.Split(data, "\n") {
		if index == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		segments := strings.Split(line, ",")
		x, err := strconv.Atoi(segments[0])
		if err != nil {
			screen.Fini()
			fmt.Printf("Invalid X coordinate at line %v\n", index+1)
			os.Exit(1)
		}
		y, err := strconv.Atoi(segments[1])
		if err != nil {
			screen.Fini()
			fmt.Printf("Invalid Y coordinate at line %v\n", index+1)
			os.Exit(1)
		}
		character := ' '
		if strings.HasSuffix(line, ",,") {
			character = ','
		} else {
			characters := []rune(segments[4])
			if len(characters) > 0 {
				character = characters[0]
			}
		}
		textColor := tcell.StyleDefault.
			Foreground(tcell.GetColor(segments[2])).
			Background(tcell.GetColor(segments[3]))
		screen.SetContent(x, y, character, nil, textColor)
	}
}
