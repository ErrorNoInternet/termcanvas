package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

var (
	block   rune = '█'
	resumed bool
	colors  []string = []string{
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
	tools = map[string]int{
		"Pencil": 0,
		"Region": 8,
		"Border": 16,
		"Text":   24,
	}
	actions = map[string]int{
		"Save":  0,
		"Load":  6,
		"Clear": 12,
		"Exit":  19,
	}
	selectedColor string = "white"
	selectedTool  string = "Pencil"
)

func drawRegion(screen tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, borderStyle tcell.Style, letter rune, drawBorders bool) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

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
	var pressed, erase bool
	var startX, startY, lastX, lastY int
	var textX, textY int = 0, 4

	colorsLength := len(colors)
	toolsLength := 0
	for tool := range tools {
		toolsLength += len(tool) + 2
	}
	actionsLength := 0
	for action := range actions {
		actionsLength += len(action) + 2
	}
	colorsOffset := 7
	toolsOffset := colorsOffset + colorsLength + 2
	actionsOffset := toolsOffset + toolsLength + 2

	for {
		screen.Show()
		width, height := screen.Size()
		var event tcell.Event
		if !resumed {
			event = screen.PollEvent()
		} else {
			event = tcell.NewEventResize(width, height)
			resumed = false
		}

		drawRegion(screen, 0, 0, 5, 3, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), defaultStyle, block, true)
		drawRegion(screen, colorsOffset-1, 0, colorsLength+colorsOffset, 3, defaultStyle, defaultStyle, ' ', true)
		for index, color := range colors {
			drawRegion(screen, index+(colorsOffset-1), 0, index+(colorsOffset+1), 3, tcell.StyleDefault.Foreground(tcell.GetColor(color)), defaultStyle, block, false)
		}
		drawRegion(screen, toolsOffset-1, 0, toolsLength+toolsOffset-2, 3, defaultStyle, defaultStyle, ' ', true)
		for tool, offset := range tools {
			for letterOffset, letter := range tool {
				drawRegion(
					screen,
					toolsOffset+letterOffset+offset-1,
					0,
					toolsOffset+letterOffset+offset+1,
					2,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
					defaultStyle,
					letter,
					false,
				)
			}
		}
		selectedToolOffset := 0
		for tool, offset := range tools {
			if selectedTool == tool {
				selectedToolOffset = offset
				break
			}
		}
		for i := 0; i < len(selectedTool); i++ {
			drawRegion(
				screen,
				(selectedToolOffset)+(toolsOffset-1)+i,
				1,
				(selectedToolOffset)+(toolsOffset+1)+i,
				3,
				tcell.StyleDefault.Foreground(tcell.ColorWhite),
				defaultStyle,
				'^',
				false,
			)
		}
		drawRegion(screen, actionsOffset-3, 0, actionsLength+actionsOffset-4, 3, defaultStyle, defaultStyle, ' ', true)
		for action, offset := range actions {
			for letterOffset, letter := range action {
				drawRegion(
					screen,
					actionsOffset-2+letterOffset+offset-1,
					0,
					actionsOffset-2+letterOffset+offset+1,
					2,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
					defaultStyle,
					letter,
					false,
				)
			}
		}

		switch event := event.(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape {
				exit(screen)
			}
			if selectedTool == "Text" {
				if textX >= width || textX <= 0 {
					textX = 0
				}
				if textY >= height || textY <= 0 {
					textY = 4
				}

				if event.Key() == tcell.KeyEnter {
					textX = 0
					textY++
				} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
					textX--
					_, _, style, _ := screen.GetContent(textX, textY)
					_, backgroundColor, _ := style.Decompose()
					textColor := tcell.StyleDefault.
						Foreground(backgroundColor).
						Background(backgroundColor)
					screen.SetContent(textX, textY, ' ', nil, textColor)
				} else {
					_, _, style, _ := screen.GetContent(textX, textY)
					originalForegroundColor, originalBackgroundColor, _ := style.Decompose()
					foregroundColor, backgroundColor := tcell.GetColor(selectedColor), originalBackgroundColor
					if backgroundColor == 0 {
						backgroundColor = originalForegroundColor
					}
					textColor := tcell.StyleDefault.
						Foreground(foregroundColor).
						Background(backgroundColor)
					screen.SetContent(textX, textY, event.Rune(), nil, textColor)
					textX++
				}
			}
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventMouse:
			x, y := event.Position()
			button := event.Buttons()
			if button == 1 {
				if y == 1 || y == 2 {
					if x < colorsLength+colorsOffset && x-colorsOffset >= 0 {
						selectedColor = colors[x-colorsOffset]
					} else if x-toolsOffset < toolsLength-2 && x >= colorsLength+colorsOffset+2 {
						for tool, offset := range tools {
							if x-toolsOffset >= offset && x-toolsOffset <= (offset+len(tool)+1) {
								selectedTool = tool
								if selectedTool == "Text" {
									textX, textY = 0, 4
								}
							}
						}
					} else if x-actionsOffset < actionsLength-4 && x >= toolsLength {
						for action, offset := range actions {
							if x-actionsOffset+2 >= offset && x-actionsOffset+2 <= (offset+len(action)+1) {
								if action == "Exit" {
									exit(screen)
								} else if action == "Clear" {
									screen.Clear()
								} else if action == "Save" {
									data, _ := dumpData(screen)
									screen.Suspend()
									selectedTool = "Pencil"

									reader := bufio.NewScanner(os.Stdin)
									fmt.Print("(Save) File Path: ")
									reader.Scan()
									filePath := reader.Text()
									if strings.TrimSpace(filePath) == "" {
										screen.Resume()
										resumed = true
										break
									}
									err := ioutil.WriteFile(filePath, []byte(data), 0644)
									if err != nil {
										fmt.Printf("Unable to write to file: %v\n", err.Error())
									} else {
										fmt.Printf("Successfully saved to %v!\n", filePath)
									}

									fmt.Print("Press Enter to continue...")
									reader.Scan()
									screen.Resume()
									resumed = true
								} else if action == "Load" {
									screen.Suspend()
									selectedTool = "Pencil"

									reader := bufio.NewScanner(os.Stdin)
									fmt.Print("(Load) File Path: ")
									reader.Scan()
									filePath := reader.Text()
									if strings.TrimSpace(filePath) == "" {
										screen.Resume()
										resumed = true
										break
									}
									data, err := ioutil.ReadFile(filePath)
									if err != nil {
										fmt.Printf("Unable to load %v: %v\n", filePath, err.Error())
										fmt.Print("Press Enter to continue...")
										reader.Scan()
										screen.Resume()
										resumed = true
									} else {
										screen.Resume()
										readData(string(data), screen)
										resumed = true
									}
								}
							}
						}
					}
				} else {
					if selectedTool == "Pencil" {
						screen.SetContent(x, y, block, nil, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)))
					} else if selectedTool == "Region" {
						if !pressed {
							pressed = true
							startX = x
							startY = y
						}
						if lastX+lastY != 0 {
							drawRegion(screen, startX, startY, lastX, lastY, defaultStyle, defaultStyle, ' ', false)
						}
						lastX = x
						lastY = y
						drawRegion(screen, startX, startY, x, y, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), defaultStyle, block, false)
					} else if selectedTool == "Border" {
						if !pressed {
							pressed = true
							startX = x
							startY = y
						}
						if lastX+lastY != 0 {
							x1, y1, x2, y2 := startX, startY, lastX, lastY
							if y2 < y1 {
								y1, y2 = y2, y1
							}
							if x2 < x1 {
								x1, x2 = x2, x1
							}
							for row := y1; row <= y2; row++ {
								for col := x1; col <= x2; col++ {
									screen.SetContent(col, row, ' ', nil, defaultStyle)
								}
							}
						}
						lastX = x
						lastY = y
						drawRegion(screen, startX, startY, x, y, defaultStyle, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), ' ', true)
					} else if selectedTool == "Text" {
						textX, textY = x, y
					}
				}
			} else if button == 2 {
				if selectedTool == "Pencil" {
					screen.SetContent(x, y, ' ', nil, defaultStyle)
				} else if selectedTool == "Region" {
					if !pressed {
						pressed = true
						erase = true
						startX = x
						startY = y
					}
					drawRegion(screen, startX, startY, x, y, defaultStyle, defaultStyle, ' ', false)
				} else if selectedTool == "Border" {
					if !pressed {
						pressed = true
						erase = true
						startX = x
						startY = y
					}
					drawRegion(screen, startX, startY, x, y, defaultStyle, defaultStyle, ' ', false)
				}
			} else if button == 0 {
				if pressed {
					pressed = false
					lastX, lastY = 0, 0
					if !erase {
						if selectedTool == "Region" {
							drawRegion(screen, startX, startY, x, y, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), defaultStyle, block, false)
						} else if selectedTool == "Border" {
							drawRegion(screen, startX, startY, x, y, defaultStyle, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), ' ', true)
						}
					}
				}
			}
		default:
			screen.SetContent(width-1, height-1, 'X', nil, background)
		}
	}
}

func exit(screen tcell.Screen) {
	data, empty := dumpData(screen)
	screen.Fini()
	if empty {
		os.Exit(0)
	}

	reader := bufio.NewScanner(os.Stdin)
	save := ""
	for {
		if save == "y" || save == "n" {
			break
		}
		fmt.Print("Would you like to save your drawing? [Y]es/[N]o: ")
		reader.Scan()
		save = reader.Text()
		if len(save) > 0 {
			save = strings.ToLower(string(save[0]))
		}
	}

	if save == "y" {
		var saved bool
		for !saved {
			fmt.Print("(Save) File Path: ")
			reader.Scan()
			filePath := reader.Text()
			if strings.TrimSpace(filePath) == "" {
				continue
			}
			err := ioutil.WriteFile(filePath, []byte(data), 0644)
			if err != nil {
				fmt.Printf("Unable to write to file: %v\n", err.Error())
			} else {
				fmt.Printf("Successfully saved to %v!\n", filePath)
				saved = true
			}
		}
	}

	os.Exit(0)
}
