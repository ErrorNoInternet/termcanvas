package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

var (
	block  rune     = '█'
	colors []string = []string{
		"black",
		"maroon",
		"green",
		"olive",
		"navy",
		"purple",
		"teal",
		"silver",
		"grey",
		"red",
		"lime",
		"yellow",
		"blue",
		"fuchsia",
		"aqua",
		"white",
	}
	tools []string = []string{
		"Pencil",
		"Square",
	}
	selectedColor string = "white"
	selectedTool  string = "Pencil"
)

func drawRegion(screen tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, letter rune, drawBorders bool) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	if drawBorders {
		for col := x1; col <= x2; col++ {
			screen.SetContent(col, y1, tcell.RuneHLine, nil, borderStyle)
			screen.SetContent(col, y2, tcell.RuneHLine, nil, borderStyle)
		}
		for row := y1 + 1; row < y2; row++ {
			screen.SetContent(x1, row, tcell.RuneVLine, nil, borderStyle)
			screen.SetContent(x2, row, tcell.RuneVLine, nil, borderStyle)
		}
		if y1 != y2 && x1 != x2 {
			screen.SetContent(x1, y1, tcell.RuneULCorner, nil, borderStyle)
			screen.SetContent(x2, y1, tcell.RuneURCorner, nil, borderStyle)
			screen.SetContent(x1, y2, tcell.RuneLLCorner, nil, borderStyle)
			screen.SetContent(x2, y2, tcell.RuneLRCorner, nil, borderStyle)
		}
	}
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			screen.SetContent(col, row, letter, nil, style)
		}
	}
}

func main() {
	encoding.Register()
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Printf("Unable to create screen: %v\n", err.Error())
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Printf("Unable to create screen: %v\n", err.Error())
		os.Exit(1)
	}
	background := tcell.StyleDefault.Background(tcell.ColorBlack)
	defaultStyle := tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorReset)
	screen.SetStyle(defaultStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()
	width, height := screen.Size()
	pressed := false
	erase := false
	startX, startY := 0, 0
	colorsOffset := 7
	toolsOffset := colorsOffset + len(colors) + 3

	for {
		screen.Show()
		event := screen.PollEvent()
		width, height = screen.Size()

		drawRegion(screen, 0, 0, 5, 3, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), block, true)
		drawRegion(screen, colorsOffset-1, 0, len(colors)+colorsOffset, 3, defaultStyle, ' ', true)
		for index, color := range colors {
			drawRegion(screen, index+(colorsOffset-1), 0, index+(colorsOffset+1), 3, tcell.StyleDefault.Foreground(tcell.GetColor(color)), block, false)
		}
		drawRegion(screen, toolsOffset-1, 0, (len(tools)*8)+toolsOffset-2, 3, defaultStyle, ' ', true)
		for index, tool := range tools {
			for letterIndex, letter := range tool {
				drawRegion(
					screen,
					(index*8)+(toolsOffset-1)+letterIndex,
					0,
					(index*8)+(toolsOffset+1)+letterIndex,
					2,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
					letter,
					false,
				)
			}
		}
		selectedToolIndex := 0
		for index, tool := range tools {
			if selectedTool == tool {
				selectedToolIndex = index
				break
			}
		}
		for i := 0; i < 6; i++ {
			drawRegion(
				screen,
				(selectedToolIndex*8)+(toolsOffset-1)+i,
				1,
				(selectedToolIndex*8)+(toolsOffset+1)+i,
				3,
				tcell.StyleDefault.Foreground(tcell.ColorWhite),
				'^',
				false,
			)
		}

		switch event := event.(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape {
				screen.Fini()
				os.Exit(0)
			}
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventMouse:
			x, y := event.Position()
			button := event.Buttons()
			if button == 1 {
				if y == 1 || y == 2 {
					if x-colorsOffset < len(colors) && x-colorsOffset >= 0 {
						selectedColor = colors[x-colorsOffset]
					}
					if x-toolsOffset < (len(tools)*8)-2 && x-toolsOffset >= 0 {
						selectedTool = tools[(x-toolsOffset)/8]
					}
				}
				if selectedTool == "Pencil" {
					screen.SetContent(x, y, block, nil, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)))
				} else if selectedTool == "Square" {
					if !pressed {
						pressed = true
						startX = x
						startY = y
					}
				}
			} else if button == 2 {
				if selectedTool == "Pencil" {
					screen.SetContent(x, y, ' ', nil, defaultStyle)
				} else if selectedTool == "Square" {
					if !pressed {
						pressed = true
						erase = true
						startX = x
						startY = y
					}
				}
			} else if button == 0 {
				if pressed {
					pressed = false
					if erase {
						erase = false
						drawRegion(screen, startX, startY, x, y, defaultStyle, ' ', false)
					} else {
						drawRegion(screen, startX, startY, x, y, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), block, false)
					}
				}
			}
		default:
			screen.SetContent(width-1, height-1, 'X', nil, background)
		}
	}
}
