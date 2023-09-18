package surreal

var (
	SettableConfigKeys = []string{
		"path",
		"baud",
		"size",
		"parity",
		"timeout",
		"stopbits",
		"verbose",
		"eol",
		"no-eol",
	}
)

type Config struct {
	Path     string   `json:"path"`
	Baud     int      `json:"baud"`
	Size     int      `json:"size"`
	Parity   Parity   `json:"parity"`
	Timeout  Duration `json:"timeout"`
	StopBits StopBits `json:"stopbits"`
	Verbose  bool     `json:"verbose"`
	EOL      uint64   `json:"eol"`
	NoEOL    bool     `json:"no-eol"`
}
