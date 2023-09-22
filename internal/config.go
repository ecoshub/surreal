package sti

import "sti/model"

var (
	EditableConfigKeys = []string{
		"path",
		"baud",
		"size",
		"parity",
		"timeout",
		"stopbits",
	}
	EditableSettingKeys = []string{
		"verbose",
		"eol",
		"eol-enable",
		"mode",
	}
)

const (
	OutputModeText string = "text"
	OutputModeChar string = "char"

	DefaultEOL       uint32 = 0x12
	DefaultVerbosity bool   = false
	DefaultEOLEnable bool   = false
	DefaultMode      string = OutputModeText
)

type Settings struct {
	Verbose   bool   `json:"verbose"`
	EOL       uint32 `json:"eol"`
	EOLEnable bool   `json:"eol-enable"`
	Mode      string `json:"mode"`
}

type Config struct {
	Path     string          `json:"path"`
	Baud     int             `json:"baud"`
	Size     int             `json:"size"`
	Parity   *model.Parity   `json:"parity"`
	Timeout  *model.Duration `json:"timeout"`
	StopBits *model.StopBits `json:"stopbits"`
}
