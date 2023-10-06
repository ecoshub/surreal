package config

import "surreal/src/model"

var (
	EditableConfigKeys = []string{
		"path",
		"baud",
		"size",
		"parity",
		"stopbits",
	}
)

const (
	OutputModeText string = "text"
	OutputModeByte string = "byte"

	DefaultEOL       uint32 = 0x0a
	DefaultEOLEnable bool   = false
	DefaultMode      string = OutputModeText
)

type Config struct {
	Path     string          `json:"path"`
	Baud     int             `json:"baud"`
	Size     int             `json:"size"`
	Parity   *model.Parity   `json:"parity"`
	StopBits *model.StopBits `json:"stopbits"`
}
