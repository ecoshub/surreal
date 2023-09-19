package sti

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sti/utils"
	"strings"
	"unicode"

	"github.com/ecoshub/jin"
	"github.com/ecoshub/termium/component/style"
)

var (
	ErrNotConnected error = errors.New("system is not connected. add a device path to start. usage: ':set path <device_path>'")
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

func (sti *STI) commandSwitch(input string) {

	args := strings.Split(input, " ")
	cmd := args[0]

	switch cmd {
	case CMDClear:
		sti.mainPanel.Clear()
		sti.termScreen.CommandPalette.PromptLine.Clear()
		sti.termScreen.CommandPalette.AddToHistory(input)
		return
	case CMDInfo:
		sti.mainPanel.Push(input, &style.Style{ForegroundColor: 59})
		err := sti.cmdInfo(input, args)
		if err != nil {
			sti.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
			return
		}
		sti.termScreen.CommandPalette.AddToHistory(input)
		return
	case CMDSet:
		sti.mainPanel.Push(input, &style.Style{ForegroundColor: 59})
		err := sti.cmdSet(input, args)
		if err != nil {
			sti.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
			return
		}
		sti.termScreen.CommandPalette.AddToHistory(input)
		return
	case CMDHelp:
		sti.mainPanel.Push("help (use commands with ':' prefix)", style.DefaultStyleWarning)
		sti.mainPanel.Push("=>  :clear       clear the screen", style.DefaultStyleWarning)
		sti.mainPanel.Push("=>  :exit        exit the program. you can also use 'Esc' key", style.DefaultStyleWarning)
		sti.mainPanel.Push("=>  :info        get serial config info", style.DefaultStyleWarning)
		sti.mainPanel.Push("=>  :set         set a value to a config field  example: ':set baud 19200'", style.DefaultStyleWarning)
		sti.mainPanel.Push("all other inputs will directly sent to serial connection", style.DefaultStyleWarning)
		return
	case CMDExit:
		if sti.connected {
			sti.stream.Flush()
			sti.stream.Close()
		}
		os.Exit(0)
		return
	}

	if !sti.connected {
		sti.mainPanel.Push(ErrNotConnected.Error(), style.DefaultStyleWarning)
		return
	}

	err := sti.cmdWrite(input, args)
	if err != nil {
		sti.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
		return
	}
	sti.termScreen.CommandPalette.AddToHistory(input)
}

func (sti *STI) cmdInfo(raw string, args []string) error {
	conf, err := json.Marshal(sti.config)
	if err != nil {
		return err
	}
	jin.IterateKeyValue(conf, func(b1, b2 []byte) (bool, error) {
		key := string(b1)
		value := string(b2)
		line := fmt.Sprintf("=>  %s: %s", key, value)
		sti.mainPanel.Push(line, style.DefaultStyleInfo)
		return true, nil
	})
	settings, err := json.Marshal(sti.setting)
	if err != nil {
		return err
	}
	jin.IterateKeyValue(settings, func(b1, b2 []byte) (bool, error) {
		key := string(b1)
		value := string(b2)
		line := fmt.Sprintf("=>  %s: %s", key, value)
		sti.mainPanel.Push(line, style.DefaultStyleInfo)
		return true, nil
	})
	return nil
}

func (sti *STI) cmdSet(raw string, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("argument count is not valid for command '%s'", args[0])
	}

	key := args[1]
	value := args[2]

	if utils.Contains(EditableConfigKeys, key) {
		return sti.editConfig(key, value)
	} else if utils.Contains(EditableSettingKeys, key) {
		return sti.editSetting(key, value)
	}

	return fmt.Errorf("'%s' is not a stitable field. (read only or not exists)", key)
}

func (sti *STI) editConfig(key, value string) error {
	conf, err := json.Marshal(sti.config)
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

	sti.StopSerial()

	currentConnectionState := sti.connected

	err = sti.Connect(temp)
	if err != nil {
		// if its connected before this config mod than it can connect it without any config change
		if currentConnectionState {
			sti.Connect(sti.config)
			sti.StartSerial()
		}
		return err
	}

	sti.config = temp

	sti.StartSerial()

	sti.mainPanel.Push(fmt.Sprintf("=> %s: %s", key, value), style.DefaultStyleInfo)

	if !currentConnectionState {
		sti.mainPanel.Push("connection success", &style.Style{ForegroundColor: 46})
	}

	return nil
}
func (sti *STI) editSetting(key, value string) error {
	if key == "mode" {
		if value != SystemModeByte && value != SystemModeText {
			return fmt.Errorf("mode '%s' is not exists. try using 'byte' and 'text' modes", value)
		}
	}

	conf, err := json.Marshal(sti.setting)
	if err != nil {
		return err
	}

	newConfig, err := jin.SetString(conf, value, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(newConfig, &sti.setting)
	if err != nil {
		return err
	}

	sti.mainPanel.Push(fmt.Sprintf("=> %s: %s", key, value), style.DefaultStyleInfo)

	return nil
}

func (sti *STI) cmdWrite(raw string, args []string) error {
	switch sti.setting.Mode {
	case SystemModeText:
		return sti.cmdWriteText(raw, args)
	case SystemModeByte:
		return sti.cmdWriteByte(raw, args)
	}
	return nil
}

func (sti *STI) cmdWriteText(raw string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("[error] argument count is not valid for command '%s'", args[0])
	}

	n, err := sti.stream.Write([]byte(raw))
	if err != nil {
		return err
	}

	rawBytes := []byte(raw[:n])

	rawBytes = utils.FormatUsingEOL(sti.setting.EOLEnable, sti.setting.EOL, rawBytes)

	pushFormat(sti, "<<", 81, rawBytes)
	return nil
}

func (sti *STI) cmdWriteByte(raw string, args []string) error {
	numbers := strings.Split(raw, " ")
	arr := make([]byte, 0, len(raw))
	for _, n := range numbers {
		b, err := utils.StringToByte(n)
		if err != nil {
			sti.mainPanel.Push(fmt.Sprintf("'%s' is not a byte value", n), style.DefaultStyleError)
			continue
		}
		arr = append(arr, b)
	}

	n, err := sti.stream.Write(arr)
	if err != nil {
		sti.mainPanel.Push(fmt.Sprintf("write error. err: %s", err), style.DefaultStyleError)
		return err
	}

	pushFormat(sti, "<<", 81, arr[:n])
	return nil
}

func pushFormat(sti *STI, direction string, color int, raw []byte) {
	if sti.setting.Verbose {
		pushVerboseFormat(sti, direction, color, raw)
		return
	}
	sti.mainPanel.Push("> "+string(raw), &style.Style{ForegroundColor: color})
}

func pushVerboseFormat(sti *STI, direction string, color int, raw []byte) {
	for i := 0; i < len(raw); i++ {
		r := raw[i]
		var s string
		if unicode.IsPrint(rune(r)) {
			s = fmt.Sprintf(VerboseDataFormat, direction, r, r, r, r)
		} else {
			s = fmt.Sprintf(VerboseNoPrintDataFormat, direction, r, r, r)
		}
		sti.mainPanel.Push(s, &style.Style{ForegroundColor: color})
	}
}
