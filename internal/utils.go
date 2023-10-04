package sti

import (
	"sti/utils"
	"strings"
)

func byteFormat(sti *STI, input string) []byte {
	if isStringInput(input) {
		arr := make([]byte, 0, len(input))
		for _, r := range input {
			b, _ := utils.StringToByte(string(r))
			arr = append(arr, b)
		}
		return arr
	}

	inputBytes := utils.FormatUsingEOL(sti.settings.EOLEnable, sti.settings.EOL.Char, []byte(input))
	return []byte(inputBytes)
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
