package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tarm/serial"
)

type StopBits struct {
	serial.StopBits
	DefaultStopBits serial.StopBits
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
			s.StopBits = serial.StopBits(s.DefaultStopBits)
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
