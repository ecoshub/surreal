package surreal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/ecoshub/jin"
	"github.com/ecoshub/termium/component/style"
)

const (
	CMDExit  string = ":exit"
	CMDClear string = ":clear"
	CMDInfo  string = ":info"
	CMDSet   string = ":set"
	CMDHelp  string = ":help"

	VerboseDataFormat        string = "[%s] %c 0x%02x 0b%08b %d"
	VerboseNoPrintDataFormat string = "[%s] . 0x%02x 0b%08b %d"
)

func (sur *Surreal) commandSwitch(input string) {

	args := strings.Split(input, " ")
	cmd := args[0]

	switch cmd {
	case CMDClear:
		sur.mainPanel.Clear()
		sur.termScreen.CommandPalette.PromptLine.Clear()
		sur.termScreen.CommandPalette.AddToHistory(input)
		return
	case CMDInfo:
		sur.mainPanel.Push(input, &style.Style{ForegroundColor: 59})
		err := sur.cmdInfo(input, args)
		if err != nil {
			sur.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
			return
		}
		sur.termScreen.CommandPalette.AddToHistory(input)
		return
	case CMDSet:
		sur.mainPanel.Push(input, &style.Style{ForegroundColor: 59})
		err := sur.cmdSet(input, args)
		if err != nil {
			sur.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
			return
		}
		sur.termScreen.CommandPalette.AddToHistory(input)
		return
	case CMDHelp:
		sur.mainPanel.Push("help (use commands with ':' prefix)", style.DefaultStyleWarning)
		sur.mainPanel.Push("=>  :clear       clear the screen", style.DefaultStyleWarning)
		sur.mainPanel.Push("=>  :exit        exit the program. you can also use 'Esc' key", style.DefaultStyleWarning)
		sur.mainPanel.Push("=>  :info        get serial config info", style.DefaultStyleWarning)
		sur.mainPanel.Push("=>  :set         set a value to a config field  example: ':set baud 19200'", style.DefaultStyleWarning)
		sur.mainPanel.Push("all other inputs will directly sent to serial connection", style.DefaultStyleWarning)
		return
	case CMDExit:
		if sur.connected {
			sur.stream.Flush()
			sur.stream.Close()
		}
		os.Exit(0)
		return
	}

	if !sur.connected {
		sur.mainPanel.Push(ErrNotConnected.Error(), style.DefaultStyleWarning)
		return
	}

	err := sur.cmdWrite(input, args)
	if err != nil {
		sur.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
		return
	}
	sur.termScreen.CommandPalette.AddToHistory(input)
}

func (sur *Surreal) cmdInfo(raw string, args []string) error {
	conf, err := json.Marshal(sur.config)
	if err != nil {
		return err
	}
	jin.IterateKeyValue(conf, func(b1, b2 []byte) (bool, error) {
		key := string(b1)
		value := string(b2)
		line := fmt.Sprintf("=>  %s: %s", key, value)
		sur.mainPanel.Push(line, style.DefaultStyleInfo)
		return true, nil
	})
	settings, err := json.Marshal(sur.settings)
	if err != nil {
		return err
	}
	jin.IterateKeyValue(settings, func(b1, b2 []byte) (bool, error) {
		key := string(b1)
		value := string(b2)
		line := fmt.Sprintf("=>  %s: %s", key, value)
		sur.mainPanel.Push(line, style.DefaultStyleInfo)
		return true, nil
	})
	return nil
}

func (sur *Surreal) cmdSet(raw string, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("argument count is not valid for command '%s'", args[0])
	}

	key := args[1]
	value := args[2]

	if Contains(EditableConfigKeys, key) {
		return sur.editConfig(key, value)
	} else if Contains(EditableSettingKeys, key) {
		return sur.editSetting(key, value)
	}

	return fmt.Errorf("'%s' is not a settable field. (read only or not exists)", key)
}

func (sur *Surreal) editConfig(key, value string) error {
	conf, err := json.Marshal(sur.config)
	if err != nil {
		return err
	}

	newConfig, err := jin.SetString(conf, value, key)
	if err != nil {
		return err
	}

	temp := &Config{}
	err = json.Unmarshal(newConfig, &temp)
	if err != nil {
		return err
	}

	sur.StopSerial()

	currentConnectionState := sur.connected

	err = sur.Connect(temp)
	if err != nil {
		// if its connected before this config mod than it can connect it without any config change
		if currentConnectionState {
			sur.Connect(sur.config)
			sur.StartSerial()
		}
		return err
	}

	sur.config = temp

	sur.StartSerial()

	sur.mainPanel.Push(fmt.Sprintf("=> %s: %s", key, value), style.DefaultStyleInfo)

	if !currentConnectionState {
		sur.mainPanel.Push("connection success", &style.Style{ForegroundColor: 46})
	}

	return nil
}
func (sur *Surreal) editSetting(key, value string) error {
	if key == "mode" {
		if value != SystemModeByte && value != SystemModeText {
			return fmt.Errorf("mode '%s' is not exists. try using 'byte' and 'text' modes", value)
		}
	}

	conf, err := json.Marshal(sur.settings)
	if err != nil {
		return err
	}

	newConfig, err := jin.SetString(conf, value, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(newConfig, &sur.settings)
	if err != nil {
		return err
	}

	sur.mainPanel.Push(fmt.Sprintf("=> %s: %s", key, value), style.DefaultStyleInfo)

	return nil
}

func (sur *Surreal) cmdWrite(raw string, args []string) error {
	switch sur.settings.Mode {
	case SystemModeText:
		return sur.cmdWriteText(raw, args)
	case SystemModeByte:
		return sur.cmdWriteByte(raw, args)
	}
	return nil
}

func (sur *Surreal) cmdWriteText(raw string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("[error] argument count is not valid for command '%s'", args[0])
	}

	_, err := sur.stream.Write([]byte(raw))
	if err != nil {
		return err
	}

	raw = sur.correctEOL(raw)

	pushDataF(sur, "sent", 81, raw)
	return nil
}

func (sur *Surreal) cmdWriteByte(raw string, args []string) error { return nil }

func (sur *Surreal) correctEOL(raw string) string {
	if sur.settings.EOLEnable {
		if sur.settings.EOL == 0 {
			return raw + string(byte(0))
		}
		eol := ""
		arr := IntToByteArray(int64(sur.settings.EOL))
		for _, c := range arr {
			if c == 0 {
				continue
			}
			eol += string(c)
		}
		return raw + eol
	}
	return raw
}

func IntToByteArray(num int64) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}
	return arr
}
