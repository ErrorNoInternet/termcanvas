package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

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

	hostServer     bool
	connectAddress string
	connections    []net.Conn
)

func setContent(screen tcell.Screen, x, y int, letter rune, style tcell.Style, send bool) {
	screen.SetContent(x, y, letter, nil, style)

	if len(connections) > 0 && y >= 4 && send {
		for _, connection := range connections {
			foregroundColorName, backgroundColorName := getColor(style)
			if foregroundColorName == "" && backgroundColorName == "" {
				foregroundColorName = "reset"
				backgroundColorName = "reset"
			}
			go fmt.Fprintf(connection, fmt.Sprintf("set:%v,%v,%v,%v,%v\n", x, y, foregroundColorName, backgroundColorName, string(letter)))
		}
	}
}

func drawRegion(
	screen tcell.Screen,
	x1, y1, x2, y2 int,
	style tcell.Style,
	borderStyle tcell.Style,
	letter rune,
	drawBorders bool,
	send bool,
) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	if drawBorders {
		for col := x1; col <= x2; col++ {
			setContent(screen, col, y1, tcell.RuneHLine, borderStyle, false)
			setContent(screen, col, y2, tcell.RuneHLine, borderStyle, false)
		}
		for row := y1 + 1; row < y2; row++ {
			setContent(screen, x1, row, tcell.RuneVLine, borderStyle, false)
			setContent(screen, x2, row, tcell.RuneVLine, borderStyle, false)
		}
		if y1 != y2 && x1 != x2 {
			setContent(screen, x1, y1, tcell.RuneULCorner, borderStyle, false)
			setContent(screen, x2, y1, tcell.RuneURCorner, borderStyle, false)
			setContent(screen, x1, y2, tcell.RuneLLCorner, borderStyle, false)
			setContent(screen, x2, y2, tcell.RuneLRCorner, borderStyle, false)
		}
	}
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			setContent(screen, col, row, letter, style, false)
		}
	}
	if len(connections) > 0 && y1 >= 4 && send {
		for _, connection := range connections {
			foregroundColorName, backgroundColorName := getColor(style)
			if foregroundColorName == "" && backgroundColorName == "" {
				foregroundColorName = "reset"
				backgroundColorName = "reset"
			}
			borderForegroundColorName, borderBackgroundColorName := getColor(borderStyle)
			if borderForegroundColorName == "" && borderBackgroundColorName == "" {
				borderForegroundColorName = "reset"
				borderBackgroundColorName = "reset"
			}
			go fmt.Fprintf(connection, fmt.Sprintf(
				"region:%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
				x1,
				y1,
				x2,
				y2,
				foregroundColorName,
				backgroundColorName,
				borderForegroundColorName,
				borderBackgroundColorName,
				string(letter),
				drawBorders,
			))
		}
	}
}

func clearRegion(screen tcell.Screen, x1, y1, x2, y2 int, send bool) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	defaultStyle := tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorReset)
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			setContent(screen, col, row, ' ', defaultStyle, false)
		}
	}
	if len(connections) > 0 && y1 >= 4 && send {
		for _, connection := range connections {
			go fmt.Fprintf(connection, fmt.Sprintf(
				"clearRegion:%v,%v,%v,%v\n",
				x1,
				y1,
				x2,
				y2,
			))
		}
	}
}

func main() {
	flag.BoolVar(&hostServer, "host", false, "Host a termcanvas server")
	flag.StringVar(&connectAddress, "connect", "", "Connect to a termcanvas server")
	flag.Parse()

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

	if hostServer {
		listener, err := net.Listen("tcp", ":55055")
		if err != nil {
			screen.Fini()
			fmt.Printf("Unable to listen for connections: %v\n", err.Error())
			os.Exit(1)
		}
		go handleConnections(listener, screen)
	} else if connectAddress != "" {
		connection, err := net.Dial("tcp", connectAddress+":55055")
		if err != nil {
			screen.Fini()
			fmt.Printf("Unable to connect to server: %v\n", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection, screen)
	}

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
	remainingOffset := actionsOffset + actionsLength + 2

	for {
		width, height := screen.Size()

		drawRegion(screen, 0, 0, width, 3, defaultStyle, defaultStyle, ' ', false, false)
		drawRegion(screen, 0, 0, 5, 3, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), defaultStyle, block, true, false)
		drawRegion(screen, colorsOffset-1, 0, colorsLength+colorsOffset, 3, defaultStyle, defaultStyle, ' ', true, false)
		for index, color := range colors {
			drawRegion(screen,
				index+(colorsOffset-1),
				0,
				index+(colorsOffset+1),
				3,
				tcell.StyleDefault.Foreground(tcell.GetColor(color)),
				defaultStyle,
				block,
				false,
				false,
			)
		}
		drawRegion(screen, toolsOffset-1, 0, toolsLength+toolsOffset-2, 3, defaultStyle, defaultStyle, ' ', true, false)
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
			setContent(
				screen,
				selectedToolOffset+toolsOffset+i,
				2,
				'^',
				tcell.StyleDefault.Foreground(tcell.ColorWhite),
				false,
			)
		}
		drawRegion(screen, actionsOffset-3, 0, actionsLength+actionsOffset-4, 3, defaultStyle, defaultStyle, ' ', true, false)
		for action, offset := range actions {
			for letterOffset, letter := range action {
				setContent(
					screen,
					actionsOffset-2+letterOffset+offset,
					1,
					letter,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
					false,
				)
			}
		}
		if len(connections) > 0 {
			for letterOffset, letter := range "Connected to:" {
				setContent(
					screen,
					remainingOffset-2+letterOffset-1,
					1,
					letter,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
					false,
				)
			}
			addresses := ""
			for _, connection := range connections {
				addresses += connection.RemoteAddr().String() + ", "
			}
			for letterOffset, letter := range strings.Trim(addresses, ", ") {
				setContent(
					screen,
					remainingOffset-2+letterOffset-1,
					2,
					letter,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
					false,
				)
			}
		}

		screen.Show()
		event := screen.PollEvent()

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
					setContent(screen, textX, textY, ' ', textColor, true)
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
					setContent(screen, textX, textY, event.Rune(), textColor, true)
					textX++
				}
			}
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventMouse:
			x, y := event.Position()
			button := event.Buttons()
			if button == 1 {
				if y <= 3 {
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
									for _, connection := range connections {
										_, err := fmt.Fprintf(connection, "clear\n")
										if err != nil {
											connection = nil
										}
									}
								} else if action == "Save" {
									data, _ := dumpData(screen)
									screen.Suspend()

									reader := bufio.NewScanner(os.Stdin)
									fmt.Print("(Save) File Path: ")
									reader.Scan()
									filePath := reader.Text()
									if strings.TrimSpace(filePath) == "" {
										screen.Resume()
										drawData(string(data), screen)
										screen.PostEvent(tcell.NewEventResize(width, height))
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
									drawData(string(data), screen)
									screen.PostEvent(tcell.NewEventResize(width, height))
								} else if action == "Load" {
									data, _ := dumpData(screen)
									screen.Suspend()

									reader := bufio.NewScanner(os.Stdin)
									fmt.Print("(Load) File Path: ")
									reader.Scan()
									filePath := reader.Text()
									if strings.TrimSpace(filePath) == "" {
										screen.Resume()
										drawData(string(data), screen)
										screen.PostEvent(tcell.NewEventResize(width, height))
										break
									}
									fileData, err := ioutil.ReadFile(filePath)
									if err != nil {
										fmt.Printf("Unable to load %v: %v\n", filePath, err.Error())
										fmt.Print("Press Enter to continue...")
										reader.Scan()
										screen.Resume()
										drawData(string(data), screen)
										screen.PostEvent(tcell.NewEventResize(width, height))
									} else {
										screen.Resume()
										drawData(string(fileData), screen)
										screen.PostEvent(tcell.NewEventResize(width, height))
									}
								}
							}
						}
					}
				} else {
					if selectedTool == "Pencil" {
						setContent(screen, x, y, block, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), true)
					} else if selectedTool == "Region" {
						if !pressed {
							pressed = true
							startX = x
							startY = y
						}
						if lastX+lastY != 0 {
							drawRegion(screen, startX, startY, lastX, lastY, defaultStyle, defaultStyle, ' ', false, true)
						}
						lastX = x
						lastY = y
						drawRegion(screen, startX, startY, x, y, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), defaultStyle, block, false, true)
					} else if selectedTool == "Border" {
						if !pressed {
							pressed = true
							startX = x
							startY = y
						}
						if lastX+lastY != 0 {
							clearRegion(screen, startX, startY, lastX, lastY, true)
						}
						lastX = x
						lastY = y
						drawRegion(screen, startX, startY, x, y, defaultStyle, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), ' ', true, true)
					} else if selectedTool == "Text" {
						textX, textY = x, y
					}
				}
			} else if button == 2 {
				if selectedTool == "Pencil" {
					setContent(screen, x, y, ' ', defaultStyle, true)
				} else if selectedTool == "Region" {
					if !pressed {
						pressed = true
						erase = true
						startX = x
						startY = y
					}
					drawRegion(screen, startX, startY, x, y, defaultStyle, defaultStyle, ' ', false, true)
				} else if selectedTool == "Border" {
					if !pressed {
						pressed = true
						erase = true
						startX = x
						startY = y
					}
					drawRegion(screen, startX, startY, x, y, defaultStyle, defaultStyle, ' ', false, true)
				}
			} else if button == 0 {
				if pressed {
					pressed = false
					lastX, lastY = 0, 0
					if !erase {
						if selectedTool == "Region" {
							drawRegion(screen, startX, startY, x, y, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), defaultStyle, block, false, true)
						} else if selectedTool == "Border" {
							drawRegion(screen, startX, startY, x, y, defaultStyle, tcell.StyleDefault.Foreground(tcell.GetColor(selectedColor)), ' ', true, true)
						}
					}
				}
			}
		}
	}
}

func exit(screen tcell.Screen) {
	for _, connection := range connections {
		fmt.Fprintf(connection, "exit\n")
		connection.Close()
	}

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
