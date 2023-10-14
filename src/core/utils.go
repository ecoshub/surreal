package core

import (
	"strings"
	"surreal/utils"
)

func formatTheInput(sti *STI, input string) []byte {
	inputBytes, ok := isText(input)
	if !ok {
		return inputBytes
	}
	inputBytes = utils.FormatUsingEOL(sti.settings.EOLEnable, sti.settings.EOL.Char, inputBytes)
	return inputBytes
}

func isText(input string) ([]byte, bool) {
	words := strings.Split(input, " ")
	arr := make([]byte, 0, len(input))
	for _, n := range words {
		b, ok := utils.IsSingleByteNotation(n)
		if !ok {
			return []byte(input), true
		}
		arr = append(arr, b)
	}
	return arr, false
}
