package surreal

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ecoshub/termium/component/palette"
	"github.com/ecoshub/termium/component/panel"
	"github.com/ecoshub/termium/component/screen"
	"github.com/ecoshub/termium/component/style"
	"github.com/ecoshub/termium/utils"
	"github.com/tarm/serial"
)

type Surreal struct {
	config     *Config
	settings   *Settings
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

	c.Parity = Parity{Parity: serial.Parity(parity)}
	c.Timeout = Duration{Duration: timeout}
	c.StopBits = StopBits{StopBits: serial.StopBits(stopBits)}
	return c
}

func New(c *Config) (*Surreal, error) {

	sc, err := screen.New(&screen.Config{
		CommandPaletteConfig: &palette.Config{
			Prompt: "surreal-terminal$ ",
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

	sur := &Surreal{
		config: c,
		settings: &Settings{
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
		sur.commandSwitch(input)
	})

	if c.Path != "" {
		err = sur.Connect(c)
		if err != nil {
			return nil, err
		}
		go sur.StartSerial()
		sur.connected = true
	}

	return sur, nil
}

func (sur *Surreal) Connect(conf *Config) error {
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
	sur.stream, err = serial.OpenPort(config)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s. please pass a valid serial device path.", err.Error())
		}
		return err
	}

	sur.connected = true
	return nil
}

func (sur *Surreal) StartSerial() {
	go sur.reader()
}

func (sur *Surreal) StartTerminal() {
	go func() {
		// wait till terminal screen initialize
		time.Sleep(time.Millisecond * 250)
		if !sur.connected {
			sur.mainPanel.Push(ErrNotConnected.Error(), style.DefaultStyleWarning)
		} else {
			sur.mainPanel.Push("connection success", &style.Style{ForegroundColor: 46})
			sur.cmdInfo("", nil)
		}
	}()
	sur.termScreen.Start()
}

func (sur *Surreal) StopSerial() {
	if sur.stream == nil {
		return
	}
	sur.stream.Close()
	sur.stop <- struct{}{}
}
