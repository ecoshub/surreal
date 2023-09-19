package sti

import (
	"flag"
	"fmt"
	"os"
	"sti/model"
	"time"

	"github.com/ecoshub/termium/component/palette"
	"github.com/ecoshub/termium/component/panel"
	"github.com/ecoshub/termium/component/screen"
	"github.com/ecoshub/termium/component/style"
	"github.com/ecoshub/termium/utils"
	"github.com/tarm/serial"
)

type STI struct {
	config     *Config
	setting    *Settings
	stream     *serial.Port
	termScreen *screen.Screen
	mainPanel  *panel.Stack
	stop       chan struct{}
	connected  bool
}

const (
	DefaultParity   int = 'N'
	DefaultBitSize  int = 8
	DefaultBaud     int = 115200
	DefaultStopBits int = 1
)

func ParseConfigFlags() *Config {

	c := &Config{}

	var parity int
	var timeout time.Duration
	var stopBits int

	flag.StringVar(&c.Path, "path", "", "device path")
	flag.IntVar(&c.Baud, "baud", DefaultBaud, "baud rate. default 115200. 115200|57600|38400|19200|9600|4800|2400|1200|600|300|200|150|134|110|75|50")
	flag.IntVar(&c.Size, "size", DefaultBitSize, "data bit size. default 8")
	flag.IntVar(&parity, "parity", DefaultParity, "parity. N|O|E|M|S")
	flag.DurationVar(&timeout, "timeout", time.Second, "read timeout. default 1 second")
	flag.IntVar(&stopBits, "stop-bit", DefaultStopBits, "stop bit. 1|15|2. default 1")

	flag.Parse()

	c.Timeout = &model.Duration{Duration: timeout}
	c.Parity = &model.Parity{Parity: serial.Parity(parity), DefaultParity: serial.Parity(DefaultParity)}
	c.StopBits = &model.StopBits{StopBits: serial.StopBits(stopBits), DefaultStopBits: serial.StopBits(DefaultStopBits)}
	return c
}

func New(c *Config) (*STI, error) {

	sc, err := screen.New(&screen.Config{
		CommandPaletteConfig: &palette.Config{
			Prompt: "sti$ ",
			Style:  &style.Style{ForegroundColor: 154},
		},
	})
	if err != nil {
		return nil, err
	}

	mainPanel := panel.NewStackPanel(&panel.Config{
		Width:  utils.TerminalWith,
		Height: utils.TerminalHeight - 2,
	})

	sc.Add(mainPanel, 0, 0)

	s := &STI{
		config: c,
		setting: &Settings{
			Verbose:   DefaultVerbosity,
			EOL:       DefaultEOL,
			EOLEnable: DefaultEOLEnable,
			Mode:      DefaultMode,
		},
		termScreen: sc,
		mainPanel:  mainPanel,
		stop:       make(chan struct{}, 1),
	}

	sc.CommandPalette.ListenKeyEventEnter(func(input string) {
		s.commandSwitch(input)
	})

	if c.Path != "" {
		err = s.Connect(c)
		if err != nil {
			return nil, err
		}
		go s.StartSerial()
		s.connected = true
	}

	return s, nil
}

func (sti *STI) Connect(conf *Config) error {
	config := &serial.Config{
		Name:        conf.Path,
		Baud:        conf.Baud,
		Size:        byte(conf.Size),
		Parity:      conf.Parity.Parity,
		StopBits:    conf.StopBits.StopBits,
		ReadTimeout: conf.Timeout.Duration,
	}

	// calculate timeout from baud rate and add an extra microsecond
	t := (1.0 / float64(conf.Baud)) * 1e9 * 8
	to := time.Duration(t) * time.Nanosecond
	to += time.Microsecond

	config.ReadTimeout = to

	var err error
	sti.stream, err = serial.OpenPort(config)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s. please pass a valid serial device path.", err.Error())
		}
		return err
	}

	sti.connected = true
	return nil
}

func (sti *STI) StartSerial() {
	go sti.reader()
}

func (sti *STI) StartTerminal() {
	go func() {
		// wait till terminal screen initialize
		time.Sleep(time.Millisecond * 250)
		if !sti.connected {
			sti.mainPanel.Push(ErrNotConnected.Error(), style.DefaultStyleWarning)
		} else {
			sti.mainPanel.Push("connection success", &style.Style{ForegroundColor: 46})
			sti.cmdInfo("", nil)
		}
	}()
	sti.termScreen.Start()
}

func (sti *STI) StopSerial() {
	if sti.stream == nil {
		return
	}
	sti.stream.Close()
	sti.stop <- struct{}{}
}