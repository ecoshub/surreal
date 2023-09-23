package sti

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unicode"

	"github.com/ecoshub/termium/component/style"
)

var (
	ReceiveStyle = &style.Style{
		ForegroundColor: 41,
	}
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
					if len(raw) > 0 {
						receivePushFormat(sti, raw)
						raw = make([]byte, 0, 32)
						continue
					}
					continue
				}
				sti.pushError(err)
				os.Exit(0)
				return
			}
			s := readBuffer[:n]
			raw = append(raw, s...)
		}
	}
}

func receivePushFormat(sti *STI, buffer []byte) {
	if sti.setting.Mode == OutputModeByte {
		for _, r := range buffer {
			var s string
			if unicode.IsPrint(rune(r)) {
				s = fmt.Sprintf(DataFormat, ">>", r, r, r, r)
			} else {
				s = fmt.Sprintf(DataFormatNoPrint, ">>", r, r, r)
			}
			sti.Print(style.SetStyle(s, ReceiveStyle))
		}
		return
	}
	sti.Print(style.SetStyle(">> "+string(buffer), ReceiveStyle))
}
