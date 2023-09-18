package surreal

import (
	"errors"
	"io"
	"os"

	"github.com/ecoshub/termium/component/style"
)

func (sur *Surreal) reader() {
	readBuffer := make([]byte, 64)
	raw := ""
	for {
		select {
		case <-sur.stop:
			return
		default:
			n, err := sur.stream.Read(readBuffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if raw == "" {
						continue
					}
					pushDataF(sur, "recv", 82, raw)
					raw = ""
					continue
				}
				sur.mainPanel.Push(err.Error(), style.DefaultStyleError)
				os.Exit(0)
				return
			}
			s := string(readBuffer[:n])
			if s == "\n" {
				pushDataF(sur, "recv", 82, raw)
				raw = ""
				continue
			}
			raw += s
			continue
		}
	}
}
