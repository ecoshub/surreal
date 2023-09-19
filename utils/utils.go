package utils

import (
	"strconv"
	"strings"
	"unsafe"
)

func Contains(arr []string, element string) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}
	return false
}

func StringToByte(raw string) (byte, error) {
	if strings.HasPrefix(raw, "0x") {
		raw = strings.TrimPrefix(raw, "0x")
		val, err := strconv.ParseUint(raw, 16, 8)
		if err != nil {
			return 0, err
		}
		return byte(val), nil
	}
	if strings.HasPrefix(raw, "0b") {
		raw = strings.TrimPrefix(raw, "0b")
		val, err := strconv.ParseUint(raw, 2, 8)
		if err != nil {
			return 0, err
		}
		return byte(val), nil
	}
	val, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {

		if len(raw) == 1 {
			return raw[0], nil
		}
		return 0, err
	}
	return byte(val), nil
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

func FormatUsingEOL(eolEnable bool, eol uint32, raw []byte) []byte {
	if eolEnable {
		if eol == 0 {
			return append(raw, 0)
		}
		eolArr := make([]byte, 0, 8)
		arr := IntToByteArray(int64(eol))
		for _, c := range arr {
			if c == 0 {
				continue
			}
			eolArr = append(eolArr, c)
		}
		return append(raw, eolArr...)
	}
	return raw
}
