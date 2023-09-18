package surreal

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
	DefaultEOL       uint32 = 0x00
	DefaultVerbosity bool   = true
	DefaultEOLEnable bool   = true
	SystemModeText   string = "text"
	SystemModeByte   string = "byte"
	DefaultMode      string = SystemModeText
)

type Settings struct {
	Verbose   bool   `json:"verbose"`
	EOL       uint32 `json:"eol"`
	EOLEnable bool   `json:"eol-enable"`
	Mode      string `json:"mode"`
}

type Config struct {
	Path     string   `json:"path"`
	Baud     int      `json:"baud"`
	Size     int      `json:"size"`
	Parity   Parity   `json:"parity"`
	Timeout  Duration `json:"timeout"`
	StopBits StopBits `json:"stopbits"`
}
