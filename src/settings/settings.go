package settings

import "surreal/src/model"

var (
	EditableSettingKeys = []string{
		"eol",
		"eol-char",
		"mode",
		"stop",
	}
)

type Settings struct {
	EOL       *model.EOLChar `json:"eol-char"`
	EOLEnable bool           `json:"eol"`
	Mode      string         `json:"mode"`
	StopPrint bool           `json:"stop"`
}
