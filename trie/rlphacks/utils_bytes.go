package rlphacks

func generateByteArrayLen(buffer []byte, pos int, l int) int {
	if l < 56 {
		buffer[pos] = byte(128 + l)
		pos++
	} else if l < 256 {
		// len(vn) can be encoded as 1 byte
		buffer[pos] = byte(183 + 1)
		pos++
		buffer[pos] = byte(l)
		pos++
	} else if l < 65536 {
		// len(vn) is encoded as two bytes
		buffer[pos] = byte(183 + 2)
		pos++
		buffer[pos] = byte(l >> 8)
		pos++
		buffer[pos] = byte(l & 255)
		pos++
	} else {
		// len(vn) is encoded as three bytes
		buffer[pos] = byte(183 + 3)
		pos++
		buffer[pos] = byte(l >> 16)
		pos++
		buffer[pos] = byte((l >> 8) & 255)
		pos++
		buffer[pos] = byte(l & 255)
		pos++
	}
	return pos
}

func generateRlpPrefixLen(l int) int {
	if l < 2 {
		return 0
	}
	if l < 56 {
		return 1
	}
	if l < 256 {
		return 2
	}
	if l < 65536 {
		return 3
	}
	return 4
}

func generateRlpPrefixLenDouble(l int, firstByte byte) int {
	if l < 2 {
		if firstByte >= 0x80 {
			return 2
		}
		return 0
	}
	if l < 55 {
		return 2
	}
	if l < 56 { // 2 + 1
		return 3
	}
	if l < 254 {
		return 4
	}
	if l < 256 {
		return 5
	}
	if l < 65533 {
		return 6
	}
	if l < 65536 {
		return 7
	}
	return 8
}

func generateByteArrayLenDouble(buffer []byte, pos int, l int) int {
	if l < 55 {
		// After first wrapping, the length will be l + 1 < 56
		buffer[pos] = byte(128 + l + 1)
		pos++
		buffer[pos] = byte(128 + l)
		pos++
	} else if l < 56 {
		buffer[pos] = byte(183 + 1)
		pos++
		buffer[pos] = byte(l + 1)
		pos++
		buffer[pos] = byte(128 + l)
		pos++
	} else if l < 254 {
		// After first wrapping, the length will be l + 2 < 256
		buffer[pos] = byte(183 + 1)
		pos++
		buffer[pos] = byte(l + 2)
		pos++
		buffer[pos] = byte(183 + 1)
		pos++
		buffer[pos] = byte(l)
		pos++
	} else if l < 256 {
		// First wrapping is 2 bytes, second wrapping 3 bytes
		buffer[pos] = byte(183 + 2)
		pos++
		buffer[pos] = byte((l + 2) >> 8)
		pos++
		buffer[pos] = byte((l + 2) & 255)
		pos++
		buffer[pos] = byte(183 + 1)
		pos++
		buffer[pos] = byte(l)
		pos++
	} else if l < 65533 {
		// Both wrappings are 3 bytes
		buffer[pos] = byte(183 + 2)
		pos++
		buffer[pos] = byte((l + 3) >> 8)
		pos++
		buffer[pos] = byte((l + 3) & 255)
		pos++
		buffer[pos] = byte(183 + 2)
		pos++
		buffer[pos] = byte(l >> 8)
		pos++
		buffer[pos] = byte(l & 255)
		pos++
	} else if l < 65536 {
		// First wrapping is 3 bytes, second wrapping is 4 bytes
		buffer[pos] = byte(183 + 3)
		pos++
		buffer[pos] = byte((l + 3) >> 16)
		pos++
		buffer[pos] = byte(((l + 3) >> 8) & 255)
		pos++
		buffer[pos] = byte((l + 3) & 255)
		pos++
		buffer[pos] = byte(183 + 2)
		pos++
		buffer[pos] = byte((l >> 8) & 255)
		pos++
		buffer[pos] = byte(l & 255)
		pos++
	} else {
		// Both wrappings are 4 bytes
		buffer[pos] = byte(183 + 3)
		pos++
		buffer[pos] = byte((l + 4) >> 16)
		pos++
		buffer[pos] = byte(((l + 4) >> 8) & 255)
		pos++
		buffer[pos] = byte((l + 4) & 255)
		pos++
		buffer[pos] = byte(183 + 3)
		pos++
		buffer[pos] = byte(l >> 16)
		pos++
		buffer[pos] = byte((l >> 8) & 255)
		pos++
		buffer[pos] = byte(l & 255)
		pos++
	}
	return pos
}
