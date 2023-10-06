package model

import (
	"encoding/json"
	"errors"

	"github.com/tarm/serial"
)

type Parity struct {
	serial.Parity
	DefaultParity serial.Parity
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
			s.Parity = serial.Parity(s.DefaultParity)
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
