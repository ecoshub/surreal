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
	ErrNotConnected error = errors.New("serial device not connected. To connect a device add its path. example: ':path /dev/tty.usbserial-110'")
)

const (
	CMDExit  string = ":exit"
	CMDClear string = ":clear"
	CMDInfo  string = ":info"
	CMDHelp  string = ":help"

	DataFormat        string = "[%s] %c 0x%02x 0b%08b %d"
	DataFormatNoPrint string = "[%s] . 0x%02x 0b%08b %d"
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
		sti.Print(style.SetStyle(ErrNotConnected.Error(), style.DefaultStyleWarning))
		return
	}

	// escape column char
	if strings.HasPrefix(input, "::") {
		input = strings.TrimPrefix(input, ":")
	}

	sti.cmdWrite(input, args)
}

func (sti *STI) cmdClear(input string, args []string) {
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
		sti.Print(style.SetStyle(line, style.DefaultStyleInfo))
		return true, nil
	})

	settings, err := json.Marshal(sti.settings)
	if err != nil {
		sti.pushError(err)
		return
	}

	jin.IterateKeyValue(settings, func(b1, b2 []byte) (bool, error) {
		key := string(b1)
		value := string(b2)
		line := fmt.Sprintf("=>  %s: %s", key, value)
		sti.Print(style.SetStyle(line, style.DefaultStyleInfo))
		return true, nil
	})
}

func (sti *STI) cmdHelp(input string, args []string) {
	sti.Print(style.SetStyle("help", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("  :help               show this help dialog", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("  :clear              clear the screen", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("  :exit               exit the program. you can also use 'Esc' key", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("  :info               get serial config info", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("  :<field> <value>    set a config <field> to a <value>. example: ':baud 9600'", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("  :<field>            toggle boolean settings (mode, eol, stop).  example: ':mode' (toggles between 'string' and 'byte' mode)", style.DefaultStyleInfo))
	sti.Print(style.SetStyle("all other inputs will directly sent to serial connection", style.DefaultStyleInfo))
}

func (sti *STI) cmdSet(input string, args []string) {
	sti.pushEcho(input)
	sti.termScreen.CommandPalette.AddToHistory(input)

	key := strings.TrimPrefix(args[0], ":")

	var value string
	if len(args) > 1 {
		value = args[1]
	}

	if utils.Contains(EditableConfigKeys, key) {
		err := sti.editConfig(key, value)
		if err != nil {
			sti.pushError(err)
			return
		}
		return
	} else if utils.Contains(EditableSettingKeys, key) {
		err := sti.editSetting(key, value)
		if err != nil {
			sti.pushError(err)
			return
		}
		return
	}

	err := fmt.Errorf("'%s' is not a settable field. (read only or not exists)", key)
	sti.pushError(err)
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

	s := style.SetStyle(fmt.Sprintf("=> %s: %s", key, value), style.DefaultStyleInfo)
	sti.Print(s)

	if !currentConnectionState {
		s := style.SetStyle("connection success", style.DefaultStyleSuccess)
		sti.Print(s)
	}

	return nil
}

func (sti *STI) editSetting(key, value string) error {
	key, value, err := sti.validateAndModifySetCommand(key, value)
	if err != nil {
		return err
	}

	conf, err := json.Marshal(sti.settings)
	if err != nil {
		return err
	}

	newConfig, err := jin.SetString(conf, value, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(newConfig, &sti.settings)
	if err != nil {
		return err
	}

	line := fmt.Sprintf("=> %s: %s", key, value)
	sti.Print(style.SetStyle(line, style.DefaultStyleInfo))

	return nil
}

func (sti *STI) cmdWrite(input string, args []string) error {
	sti.termScreen.CommandPalette.AddToHistory(input)

	switch sti.settings.Mode {
	case OutputModeString:
		return sti.cmdWriteText(input, args)
	case OutputModeByte:
		return sti.cmdWriteByte(input, args)
	}
	return nil
}

func (sti *STI) cmdWriteText(input string, args []string) error {

	sti.Print(style.SetStyle("<< "+input, style.DefaultStyleEvent))

	inputBytes := utils.FormatUsingEOL(sti.settings.EOLEnable, sti.settings.EOL.Char, []byte(input))

	_, err := sti.stream.Write(inputBytes)
	if err != nil {
		return err
	}

	return nil
}

func (sti *STI) cmdWriteByte(input string, args []string) error {
	arr := byteFormat(sti, input)

	for _, r := range arr {
		var s string
		if unicode.IsPrint(rune(r)) {
			s = fmt.Sprintf(DataFormat, "<<", r, r, r, r)
		} else {
			s = fmt.Sprintf(DataFormatNoPrint, "<<", r, r, r)
		}
		sti.Print(style.SetStyle(s, style.DefaultStyleEvent))
	}

	n, err := sti.stream.Write(arr)
	if err != nil {
		sti.pushError(err)
		return err
	}

	arr = arr[:n]

	return nil
}

func (sti *STI) isSetCommand(input string, args []string) bool {
	if !strings.HasPrefix(args[0], ":") {
		return false
	}

	return true
}

func (sti *STI) validateAndModifySetCommand(key, value string) (string, string, error) {
	switch key {
	case "mode":
		switch value {
		case "":
			if sti.settings.Mode == OutputModeByte {
				return key, OutputModeString, nil
			}
			if sti.settings.Mode == OutputModeString {
				return key, OutputModeByte, nil
			}
		case "s":
			return key, OutputModeString, nil
		case "b":
			return key, OutputModeByte, nil
		case OutputModeByte, OutputModeString:
			return key, value, nil
		}
	case "eol":
		if value == "" {
			value = fmt.Sprint(!sti.settings.EOLEnable)
		}
	case "stop":
		if value == "" {
			value = fmt.Sprint(!sti.settings.StopPrint)
		}
	}
	if value == "" {
		return "", "", errors.New("value can not be empty")
	}
	return key, value, nil
}

func (sti *STI) pushError(err error) {
	sti.Print(style.SetStyle("[error] "+err.Error(), style.DefaultStyleError))
}

func (sti *STI) pushEcho(text string) {
	sti.Print(text)
}
