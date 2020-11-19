package utils

func EncodePacketType(i byte) byte {
	var res byte
	if i >= 8 {
		i -= 8
		res += 128
	}
	if i >= 4 {
		i -= 4
		res += 64
	}
	if i >= 2 {
		i -= 2
		res += 32
	}
	if i >= 1 {
		i -= 1
		res += 16
	}
	return res
}

func DecodePacketType(byte1 byte) byte {
	var res byte
	if byte1 >= 128 {
		byte1 -= 128
		res += 8
	}
	if byte1 >= 64 {
		byte1 -= 64
		res += 4
	}
	if byte1 >= 32 {
		byte1 -= 32
		res += 2
	}
	if byte1 >= 16 {
		byte1 -= 16
		res += 1
	}
	return res
}
