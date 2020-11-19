package utils

import (
	"gomq/protocol"
	"testing"
)

var pakcet_tests = []struct {
	in  byte
	out byte
}{
	{byte(protocol.CONNECT), byte(protocol.CONNECT)},
	{byte(protocol.CONNACK), byte(protocol.CONNACK)},
	{byte(protocol.PUBLISH), byte(protocol.PUBLISH)},
	{byte(protocol.PUBACK), byte(protocol.PUBACK)},
	{byte(protocol.PUBREC), byte(protocol.PUBREC)},
	{byte(protocol.PUBREL), byte(protocol.PUBREL)},
	{byte(protocol.PUBCOMP), byte(protocol.PUBCOMP)},
	{byte(protocol.SUBSCRIBE), byte(protocol.SUBSCRIBE)},
	{byte(protocol.SUBACK), byte(protocol.SUBACK)},
	{byte(protocol.UNSUBSCRIBE), byte(protocol.UNSUBSCRIBE)},
	{byte(protocol.UNSUBACK), byte(protocol.UNSUBACK)},
	{byte(protocol.PINGREQ), byte(protocol.PINGREQ)},
	{byte(protocol.PINGRESP), byte(protocol.PINGRESP)},
	{byte(protocol.DISCONNECT), byte(protocol.DISCONNECT)},
}

func TestEncodePacketType(t *testing.T) {
	for _, tt := range pakcet_tests {
		s := EncodePacketType(tt.in)
		s = DecodePacketType(s)
		if s != tt.out {
			t.Errorf(" %d => %d, wanted: %d", tt.in, s, tt.out)
		}
	}
}
