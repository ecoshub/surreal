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
	CMDHelp  string = ":help"
	CMDDump  string = ":dump"
	CMDFlush string = ":flush"

	VerboseDataFormat        string = "[%s] %c 0x%02x 0b%08b %d"
	VerboseNoPrintDataFormat string = "[%s] . 0x%02x 0b%08b %d"
)

var (
	StyleNeutral = &style.Style{ForegroundColor: 59}
)

func (sti *STI) commandSwitch(input string) {

	args := strings.Split(input, " ")
	cmd := args[0]

	switch cmd {
	case CMDClear:
		sti.cmdClear(input, args)
		return
	case CMDInfo:
		sti.cmdInfo(input, args)
		return
	case CMDHelp:
		sti.cmdHelp(input, args)
		return
	case CMDDump:
		sti.cmdDump(input, args)
		return
	case CMDFlush:
		sti.cmdFlush(input, args)
		return
	case CMDExit:
		if sti.connected {
			sti.stream.Flush()
			sti.stream.Close()
		}
		os.Exit(0)
		return
	}

	if sti.isSetCommand(input, args) {
		sti.cmdSet(input, args)
		return
	}

	if !sti.connected {
		sti.mainPanel.Push(ErrNotConnected.Error(), style.DefaultStyleWarning)
		return
	}

	sti.cmdWrite(input, args)
}

func (sti *STI) cmdClear(input string, args []string) {
	sti.mainPanel.Clear()
	sti.termScreen.CommandPalette.PromptLine.Clear()
	sti.termScreen.CommandPalette.AddToHistory(input)
}

func (sti *STI) cmdInfo(input string, args []string) {
	sti.pushEcho(input)
	sti.termScreen.CommandPalette.AddToHistory(input)

	conf, err := json.Marshal(sti.config)
	if err != nil {
		sti.pushError(err)
		return
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
		sti.pushError(err)
		return
	}

	jin.IterateKeyValue(settings, func(b1, b2 []byte) (bool, error) {
		key := string(b1)
		value := string(b2)
		line := fmt.Sprintf("=>  %s: %s", key, value)
		sti.mainPanel.Push(line, style.DefaultStyleInfo)
		return true, nil
	})
}

func (sti *STI) cmdHelp(input string, args []string) {
	sti.mainPanel.Push("help (use commands with ':' prefix)", style.DefaultStyleWarning)
	sti.mainPanel.Push("=>  :clear       clear the screen", style.DefaultStyleWarning)
	sti.mainPanel.Push("=>  :exit        exit the program. you can also use 'Esc' key", style.DefaultStyleWarning)
	sti.mainPanel.Push("=>  :info        get serial config info", style.DefaultStyleWarning)
	sti.mainPanel.Push("=>  :<field>     set a value to a config field  example: ':baud 19200' or ':verbose true'", style.DefaultStyleWarning)
	sti.mainPanel.Push("all other inputs will directly sent to serial connection", style.DefaultStyleWarning)
}

func (sti *STI) cmdFlush(input string, args []string) {
	sti.pushEcho(input)
	sti.termScreen.CommandPalette.AddToHistory(input)

	sti.mainPanel.Flush()
}

func (sti *STI) cmdDump(input string, args []string) {
	sti.pushEcho(input)
	sti.termScreen.CommandPalette.AddToHistory(input)

	if len(args) < 2 {
		err := fmt.Errorf("argument count is not valid for command '%s'", args[0])
		sti.pushError(err)
		return
	}

	path := args[1]

	n, err := sti.mainPanel.Dump(path)
	if err != nil {
		sti.pushError(err)
		return
	}

	msg := fmt.Sprintf("file dump success. path: %s, size: %d bytes", path, n)
	sti.mainPanel.Push(msg, style.DefaultStyleSuccess)

}

func (sti *STI) cmdSet(input string, args []string) {
	sti.pushEcho(input)
	sti.termScreen.CommandPalette.AddToHistory(input)

	key := strings.TrimPrefix(args[0], ":")
	value := args[1]

	if utils.Contains(EditableConfigKeys, key) {
		err := sti.editConfig(key, value)
		if err != nil {
			sti.mainPanel.Push(err.Error(), style.DefaultStyleError)
			return
		}
		return
	} else if utils.Contains(EditableSettingKeys, key) {
		err := sti.editSetting(key, value)
		if err != nil {
			sti.mainPanel.Push(err.Error(), style.DefaultStyleError)
			return
		}
		return
	}

	err := fmt.Errorf("'%s' is not a settable field. (read only or not exists)", key)
	sti.mainPanel.Push(err.Error(), style.DefaultStyleError)
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
		sti.mainPanel.Push("connection success", style.DefaultStyleSuccess)
	}

	return nil
}

func (sti *STI) editSetting(key, value string) error {
	if key == "mode" {
		if value != OutputModeChar && value != OutputModeText {
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

func (sti *STI) cmdWrite(input string, args []string) error {
	sti.termScreen.CommandPalette.AddToHistory(input)

	switch sti.setting.Mode {
	case OutputModeText:
		return sti.cmdWriteText(input, args)
	case OutputModeChar:
		return sti.cmdWriteByte(input, args)
	}
	return nil
}

func (sti *STI) cmdWriteText(input string, args []string) error {

	n, err := sti.stream.Write([]byte(input))
	if err != nil {
		return err
	}

	inputBytes := []byte(input[:n])

	inputBytes = utils.FormatUsingEOL(sti.setting.EOLEnable, sti.setting.EOL, inputBytes)

	pushFormat(sti, "<<", 81, inputBytes)
	return nil
}

func (sti *STI) cmdWriteByte(input string, args []string) error {
	numbers := strings.Split(input, " ")
	arr := make([]byte, 0, len(input))
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

func pushFormat(sti *STI, direction string, color int, input []byte) {
	if sti.setting.Mode == OutputModeChar {
		pushVerboseFormat(sti, direction, color, input)
		return
	}
	if sti.setting.Verbose {
		pushVerboseFormat(sti, direction, color, input)
		return
	}
	msg := fmt.Sprintf("%s %s", direction, string(input))
	sti.mainPanel.Push(msg, &style.Style{ForegroundColor: color})
}

func pushVerboseFormat(sti *STI, direction string, color int, input []byte) {
	for i := 0; i < len(input); i++ {
		r := input[i]
		var s string
		if unicode.IsPrint(rune(r)) {
			s = fmt.Sprintf(VerboseDataFormat, direction, r, r, r, r)
		} else {
			s = fmt.Sprintf(VerboseNoPrintDataFormat, direction, r, r, r)
		}
		sti.mainPanel.Push(s, &style.Style{ForegroundColor: color})
	}
}

func (sti *STI) isSetCommand(input string, args []string) bool {
	if len(args) != 2 {
		return false
	}

	if !strings.HasPrefix(args[0], ":") {
		return false
	}

	return true
}

func (sti *STI) pushError(err error) {
	sti.mainPanel.Push("[error] "+err.Error(), style.DefaultStyleError)
}

func (sti *STI) pushEcho(text string) {
	sti.mainPanel.Push(text, StyleNeutral)
}
