package model

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type EOLChar struct {
	Char uint32
}

func (ec *EOLChar) UnmarshalJSON(b []byte) error {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		value = strings.TrimPrefix(value, "0x")
		val, err := strconv.ParseUint(value, 16, 32)
		if err != nil {
			return errors.New("eol value format must be a hex value. err: " + err.Error())
		}
		ec.Char = uint32(val)
	default:
		return errors.New("error unmarshal 'eol-char' value")
	}
	return nil
}

func (ec *EOLChar) MarshalJSON() ([]byte, error) {
	s := "0x" + strconv.FormatUint(uint64(ec.Char), 16)
	return []byte(`"` + s + `"`), nil
}
