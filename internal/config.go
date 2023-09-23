package sti

import "sti/model"

var (
	EditableConfigKeys = []string{
		"path",
		"baud",
		"size",
		"parity",
		"stopbits",
	}
	EditableSettingKeys = []string{
		"eol",
		"eol-char",
		"mode",
		"stop",
	}
)

const (
	OutputModeString string = "string"
	OutputModeByte   string = "byte"

	DefaultEOL       uint32 = 0x0a
	DefaultEOLEnable bool   = false
	DefaultMode      string = OutputModeString
)

type Settings struct {
	EOL       *model.EOLChar `json:"eol-char"`
	EOLEnable bool           `json:"eol"`
	Mode      string         `json:"mode"`
	StopPrint bool           `json:"stop"`
}

type Config struct {
	Path     string          `json:"path"`
	Baud     int             `json:"baud"`
	Size     int             `json:"size"`
	Parity   *model.Parity   `json:"parity"`
	StopBits *model.StopBits `json:"stopbits"`
}
