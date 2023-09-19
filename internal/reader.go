package sti

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/ecoshub/termium/component/style"
)

func (sti *STI) reader() {
	readBuffer := make([]byte, 64)
	raw := make([]byte, 0, 32)
	for {
		select {
		case <-sti.stop:
			return
		default:
			n, err := sti.stream.Read(readBuffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(raw) == 0 {
						continue
					}
					pushFormat(sti, ">>", 82, raw)
					raw = make([]byte, 0, 32)
					continue
				}
				sti.mainPanel.Push(err.Error(), style.DefaultStyleError)
				os.Exit(0)
				return
			}
			s := readBuffer[:n]
			raw = append(raw, s...)
			if bytes.HasSuffix(s, []byte{'\n'}) || bytes.HasSuffix(s, []byte{'\n', '\r'}) {
				pushFormat(sti, ">>", 82, raw)
				raw = make([]byte, 0, 32)
				continue
			}
		}
	}
}
