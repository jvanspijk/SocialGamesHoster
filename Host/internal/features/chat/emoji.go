package chat

import "unicode"

func isEmojiOnly(value string) bool {
	hasEmoji := false
	runes := []rune(value)
	for index := 0; index < len(runes); index++ {
		value := runes[index]
		if unicode.IsSpace(value) {
			continue
		}
		if isEmojiBase(value) {
			hasEmoji = true
			continue
		}
		if isEmojiJoiner(value) && hasEmoji {
			continue
		}
		if (value == '#' || value == '*' || value >= '0' && value <= '9') && keycapFollows(runes, index) {
			hasEmoji = true
			continue
		}
		return false
	}
	return hasEmoji
}

func keycapFollows(values []rune, index int) bool {
	index++
	if index < len(values) && values[index] == '\ufe0f' {
		index++
	}
	return index < len(values) && values[index] == '\u20e3'
}

func isEmojiJoiner(value rune) bool {
	return value == '\ufe0f' || value == '\ufe0e' || value == '\u200d' || value == '\u20e3' ||
		value >= 0x1f3fb && value <= 0x1f3ff
}

func isEmojiBase(value rune) bool {
	switch {
	case value >= 0x1f000 && value <= 0x1faff:
		return true
	case value >= 0x2600 && value <= 0x27bf:
		return true
	case value >= 0x2300 && value <= 0x23ff:
		return true
	case value >= 0x2b00 && value <= 0x2bff:
		return true
	}
	switch value {
	case '\u00a9', '\u00ae', '\u203c', '\u2049', '\u2122', '\u2139',
		'\u3030', '\u303d', '\u3297', '\u3299':
		return true
	default:
		return false
	}
}
