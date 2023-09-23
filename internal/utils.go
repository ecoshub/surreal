package sti

import (
	"sti/utils"
	"strings"
)

func byteFormat(sti *STI, input string) []byte {
	if !isStringInput(input) {
		inputBytes := utils.FormatUsingEOL(sti.setting.EOLEnable, sti.setting.EOL.Char, []byte(input))
		return []byte(inputBytes)
	}

	words := strings.Split(input, " ")
	arr := make([]byte, 0, len(input))
	for _, r := range words {
		b, _ := utils.StringToByte(string(r))
		arr = append(arr, b)
	}
	return arr
}

func isStringInput(input string) bool {
	words := strings.Split(input, " ")
	for _, n := range words {
		_, ok := utils.StringToByte(n)
		if !ok {
			return false
		}
	}
	return true
}
