# termcanvas
Draw stuff in your terminal!
![Screenshot](https://raw.githubusercontent.com/ErrorNoInternet/termcanvas/main/screenshots/0.png)

## Features
 - Placing pixels
 - 16 different colors
 - Drawing filled squares
 - Drawing empty boxes
 - Displaying custom text
 - Saving & loading (CSV)
 - Multiplayer support

#### Colors
It's possible to use more than 16 colors, by modifying the color names saved in the CSV files to hex codes.
See [examples/hex-colors.csv](https://github.com/ErrorNoInternet/termcanvas/blob/main/examples/hex-colors.csv) for an example.

#### Multiplayer support
To host a server, run `termcanvas -host`, which hosts a termcanvas server on port 55055 (you can change this with `termcanvas -host -port XXXXX`).
To connect to a termcanvas server, run `termcanvas -connect example.com` (or `termcanvas -connect example.com -port XXXXX` for a custom port).
The server host knows the IP address of whoever connects (clients can only see the server IP), and multiple people can connect to the same server.

## Keybindings
`esc`: exit termcanvas\
`left click`: place a pixel (works with the Region tool, which draws a region)\
`right click`: remove a pixel (works with the Region tool, which removes a region)

## Compiling
```
git clone https://github.com/ErrorNoInternet/termcanvas
cd termcanvas
go build
```

<sub>If you would like to modify or use this repository (including its code) in your own project, please be sure to credit!</sub>

