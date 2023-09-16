package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ecoshub/termium/component/palette"
	"github.com/ecoshub/termium/component/panel"
	"github.com/ecoshub/termium/component/screen"
	"github.com/ecoshub/termium/component/style"
	"github.com/ecoshub/termium/utils"
	"github.com/tarm/serial"
)

var (
	flagPath       = flag.String("path", "/dev/tty.usbserial-10", "device path")
	flagBaud       = flag.Int("baud", 9600, "baud rate. default 9600. 115200|57600|38400|19200|9600|4800|2400|1200|600|300|200|150|134|110|75|50")
	flagSize       = flag.Int("size", 8, "data bit size. default 8")
	flagParity     = flag.String("parity", "N", "parity. N|O|E|M|S")
	flagTimeout    = flag.Duration("timeout", time.Second, "read timeout. default 1 second")
	flagStopBit    = flag.Int("stop-bit", 1, "stop bit. 1|15|2. default 1")
	flagWrite      = flag.String("write", "", "string to write")
	flagOnce       = flag.Bool("once", false, "write only once")
	flagWriteDelay = flag.Duration("delay", time.Second, "delay between write operations")
	flagTestRead   = flag.String("test-char", "", "read test char")
)

func main() {

	flag.Parse()

	if *flagParity == "" {
		log.Fatal("define one char for 'parity' flag")
	}

	parity := byte((*flagParity)[0])

	config := &serial.Config{
		Name:        *flagPath,
		Baud:        *flagBaud,
		Size:        byte(*flagSize),
		Parity:      serial.Parity(parity),
		StopBits:    serial.StopBits(*flagStopBit),
		ReadTimeout: *flagTimeout,
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}

	sc, err := screen.New(&palette.CommandPaletteConfig{
		Prompt: "eco$ ",
		Style:  &style.Style{ForegroundColor: 192},
	})
	if err != nil {
		log.Fatal(err)
	}

	p := panel.NewStackPanel(&panel.Config{
		Width:  utils.TerminalWith,
		Height: utils.TerminalHeight - 2,
	})

	sc.Add(p, 0, 0)

	sc.CommandPalette.ListenKeyEventEnter(func(input string) {
		for _, r := range input {
			stream.Write([]byte(input))
			s := fmt.Sprintf("<< %2x %08b %c %d\n", r, r, r, r)
			p.Push(s)
		}
		sc.CommandPalette.AddToHistory(input)
	})

	go reader(stream, p)

	sc.Start()
}

func reader(stream *serial.Port, p *panel.Stack) {
	readBuffer := make([]byte, 64)
	for {
		n, err := stream.Read(readBuffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			log.Fatal(err)
		}
		s := string(readBuffer[:n])
		p.Push(s)
	}
}
