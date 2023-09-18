package surreal

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/ecoshub/termium/component/style"
	"github.com/tarm/serial"
)

var (
	ErrNotConnected error = errors.New("system is not connected. add a device path to start. usage: ':set path <device_path>'")
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("error unmarshal 'duration' value")
	}
}

func (d *Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

type StopBits struct {
	serial.StopBits
}

func (s *StopBits) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		if value == 0 {
			s.StopBits = serial.StopBits(DefaultStopBits)
			return nil
		}
		s.StopBits = serial.StopBits(value)
		return nil

	default:
		return errors.New("error unmarshal 'stop-bits' value")
	}
}

func (s *StopBits) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprint(s.StopBits)), nil
}

type Parity struct {
	serial.Parity
}

func (s *Parity) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		if len(value) == 0 {
			s.Parity = serial.Parity('N')
			return nil
		}
		s.Parity = serial.Parity(value[0])
		return nil
	default:
		return errors.New("error unmarshal 'parity' value")
	}
}

func (p *Parity) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(p.Parity) + `"`), nil
}

func Contains(arr []string, element string) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}
	return false
}

func pushDataF(sur *Surreal, prefix string, color int, raw string) {
	if sur.settings.Verbose {
		for _, r := range raw {
			var s string
			if unicode.IsPrint(r) {
				s = fmt.Sprintf(VerboseDataFormat, prefix, r, r, r, r)
			} else {
				s = fmt.Sprintf(VerboseNoPrintDataFormat, prefix, r, r, r)
			}
			sur.mainPanel.Push(s, &style.Style{ForegroundColor: color})
		}
		return
	}
	sur.mainPanel.Push("> "+raw, &style.Style{ForegroundColor: color})
}
