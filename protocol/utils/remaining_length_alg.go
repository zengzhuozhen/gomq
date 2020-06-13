package utils

import "io"

const (
	FirstBytePosition  = 0x80
	SecondBytePosition = FirstBytePosition << 7
	ThirdBytePosition  = SecondBytePosition << 7
	FourthBytePosition = ThirdBytePosition << 7
)

func DecodeRemainingLengthAlg(r io.Reader) (int, error) {
	var rLength uint32
	var multiplier uint32
	b := make([]byte, 1)
	for multiplier < 27 { //fix: Infinite '(digit & 128) == 1' will cause the dead loop
		_, err := io.ReadFull(r, b)
		if err != nil {
			return 0, err
		}

		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
	}
	return int(rLength), nil
}


func EncodeRemainingLengthAlg(length int) []byte {
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength
}



