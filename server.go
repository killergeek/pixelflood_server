package pixelflood_server

import (
	"net"
	"log"
	"strconv"
	"strings"
	"fmt"
	"bufio"
)

type Pixel struct {
	R uint8
	G uint8
	B uint8
}

type PixelServer struct {
	Pixels       [][]Pixel
	screenWidth  uint16
	screenHeight uint16
	socket       *net.Listener
	connections  []net.Conn
	shouldClose  bool
}

func NewServer(width uint16, height uint16) (*PixelServer) {
	pixels := make([][]Pixel, width)
	for i := uint16(0); i < width; i++ {
		pixels[i] = make([]Pixel, height)
	}

	socket, err := net.Listen("tcp", ":1234")

	if err != nil {
		panic(err)
	}

	return &PixelServer{pixels, width, height, &socket, make([]net.Conn, 0), false}
}

func (server *PixelServer) Run() {
	for !server.shouldClose {
		conn, err := (*server.socket).Accept()

		if err != nil {
			log.Println("Error accepting new connection: ", err)
			continue
		}

		server.connections = append(server.connections, conn)
		go server.handleRequest(&conn)
	}
}

func (server *PixelServer) Stop() {
	server.shouldClose = true
	for _, conn := range server.connections {
		conn.Close()
	}
	(*server.socket).Close()
}

func (server *PixelServer) handleRequest(conn *net.Conn) {
	scanner := bufio.NewScanner(*conn)

	for !server.shouldClose && scanner.Scan() {
		data := scanner.Text()

		// Malformed packet, does not contain recognised command
		if len(data) < 1 {
			continue
		}

		// Strip newline, and split by spaces to get command components
		commandComponents := strings.Split(data, " ")

		// For every commandComponents data, pass on its components
		if len(commandComponents) > 0 {
			x, y, pixel, err := parsePixelCommand(commandComponents)
			if err == nil {
				server.setPixel(x, y, pixel)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading standard input:", err)
		return
	}
	(*conn).Close()
}

func (server *PixelServer) setPixel(x uint16, y uint16, pixel *Pixel) {
	if x >= server.screenWidth || y >= server.screenHeight {
		return
	}

	server.Pixels[x][y] = *pixel
}

func parsePixelCommand(commandPieces []string) (uint16, uint16, *Pixel, error) {
	if len(commandPieces) != 4 {
		return 0, 0, nil, fmt.Errorf("Command length mismatch")
	}

	x, err := strconv.ParseUint(commandPieces[1], 10, 16)
	if err != nil {
		return 0, 0, nil, err
	}

	y, err := strconv.ParseUint(commandPieces[2], 10, 16)
	if err != nil {
		return 0, 0, nil, err
	}

	pixelValue := commandPieces[3]
	if len(pixelValue) != 6 {
		return 0, 0, nil, fmt.Errorf("Pixel length mismatch")
	}

	r, err := strconv.ParseUint(pixelValue[0:2], 16, 8)
	if err != nil {
		return 0, 0, nil, err
	}

	g, err := strconv.ParseUint(pixelValue[2:4], 16, 8)
	if err != nil {
		return 0, 0, nil, err
	}

	b, err := strconv.ParseUint(pixelValue[4:6], 16, 8)
	if err != nil {
		return 0, 0, nil, err
	}

	return uint16(x), uint16(y), &Pixel{uint8(r), uint8(g), uint8(b)}, nil
}
