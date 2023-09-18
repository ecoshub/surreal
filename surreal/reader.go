package surreal

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/ecoshub/termium/component/style"
)

func (sur *Surreal) reader() {
	readBuffer := make([]byte, 64)
	raw := make([]byte, 0, 32)
	for {
		select {
		case <-sur.stop:
			return
		default:
			n, err := sur.stream.Read(readBuffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(raw) == 0 {
						continue
					}
					pushDataF(sur, "recv", 82, string(raw))
					raw = make([]byte, 0, 32)
					continue
				}
				sur.mainPanel.Push(err.Error(), style.DefaultStyleError)
				os.Exit(0)
				return
			}
			s := readBuffer[:n]
			if bytes.HasSuffix(s, []byte{'\n'}) || bytes.HasSuffix(s, []byte{'\n', '\r'}) {
				pushDataF(sur, "recv", 82, string(raw))
				raw = make([]byte, 0, 32)
				continue
			}
			raw = append(raw, s...)
			continue
		}
	}
}
