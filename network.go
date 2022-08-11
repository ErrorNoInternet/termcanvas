package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func handleConnections(listener net.Listener, screen tcell.Screen) {
	for {
		connection, _ = listener.Accept()
		data, empty := dumpData(screen)
		var newData string
		if !empty {
			lines := strings.Split(data, "\n")
			for index, line := range lines {
				if index == 0 || strings.TrimSpace(line) == "" {
					continue
				}
				newData += "set:" + line + "\n"
			}
		}
		fmt.Fprintf(connection, newData)
		go handleConnection(screen)
	}
}

func handleConnection(screen tcell.Screen) {
	reader := bufio.NewReader(connection)
	for {
		rawMessage, err := reader.ReadString('\n')
		if err != nil {
			connection = nil
			return
		}
		message := strings.TrimSpace(string(rawMessage))
		if message == "exit" {
			connection.Close()
			connection = nil
		}

		if strings.HasPrefix(message, "set:") {
			segments := strings.Split(strings.Split(message, "set:")[1], ",")
			x, err := strconv.Atoi(segments[0])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid X coordinate received")
				os.Exit(1)
			}
			y, err := strconv.Atoi(segments[1])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid Y coordinate received")
				os.Exit(1)
			}
			character := ' '
			if strings.HasSuffix(message, ",,") {
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
			screen.Show()
		} else if strings.HasPrefix(message, "region:") {
			segments := strings.Split(strings.Split(message, "region:")[1], ",")
			x1, err := strconv.Atoi(segments[0])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid X1 coordinate received")
				os.Exit(1)
			}
			y1, err := strconv.Atoi(segments[1])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid Y1 coordinate received")
				os.Exit(1)
			}
			x2, err := strconv.Atoi(segments[2])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid X2 coordinate received")
				os.Exit(1)
			}
			y2, err := strconv.Atoi(segments[3])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid Y2 coordinate received")
				os.Exit(1)
			}
			textColor := tcell.StyleDefault.
				Foreground(tcell.GetColor(segments[4])).
				Background(tcell.GetColor(segments[5]))
			borderStyle := tcell.StyleDefault.
				Foreground(tcell.GetColor(segments[6])).
				Background(tcell.GetColor(segments[7]))
			drawBorders := false
			if segments[9] == "true" {
				drawBorders = true
			}
			drawRegion(screen, x1, y1, x2, y2, textColor, borderStyle, []rune(segments[8])[0], drawBorders, false)
			width, height := screen.Size()
			screen.PostEvent(tcell.NewEventResize(width, height))
		} else if strings.HasPrefix(message, "clearRegion:") {
			segments := strings.Split(strings.Split(message, "clearRegion:")[1], ",")
			x1, err := strconv.Atoi(segments[0])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid X1 coordinate received")
				os.Exit(1)
			}
			y1, err := strconv.Atoi(segments[1])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid Y1 coordinate received")
				os.Exit(1)
			}
			x2, err := strconv.Atoi(segments[2])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid X2 coordinate received")
				os.Exit(1)
			}
			y2, err := strconv.Atoi(segments[3])
			if err != nil {
				screen.Fini()
				fmt.Println("Invalid Y2 coordinate received")
				os.Exit(1)
			}
			clearRegion(screen, x1, y1, x2, y2, false)
		} else if message == "clear" {
			screen.Clear()
			width, height := screen.Size()
			screen.PostEvent(tcell.NewEventResize(width, height))
		}
	}
}
