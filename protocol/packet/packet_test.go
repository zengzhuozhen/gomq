package packet

import (
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"testing"
)

func TestIsLegalIdentifier(t *testing.T) {

}

var ConnectFlagTests = []struct {
	CleanSession bool
	WillFlag     bool
	WillQos      bool
	WillRetain   bool
	UserNameFlag bool
	PasswordFlag bool
	out          byte
}{
	{false, false, false, false, false, false, 0},
	{true, false, false, false, false, false, 2},
	{false, true, false, false, false, false, 4},
	{false, false, true, false, false, false, 24},
	{false, false, false, true, false, false, 32},
	{false, false, false, false, false, true, 64},
	{false, false, false, false, true, false, 128},
}

func TestEncodeConnectFlag(t *testing.T) {
	for i, tt := range ConnectFlagTests {
		connectFlag := ConnectFlags{
			CleanSession: tt.CleanSession,
			WillFlag:     tt.WillFlag,
			WillQos:      tt.WillQos,
			WillRetain:   tt.WillRetain,
			UserNameFlag: tt.UserNameFlag,
			PasswordFlag: tt.PasswordFlag,
		}
		if connectFlag.encode() != tt.out {
			t.Errorf("%d => %q, wanted: %q", i, connectFlag.encode(), tt.out)
		}
	}
}

var PacketType = []struct {
	in  byte
	out byte
}{
	{byte(protocol.Reserved), 0},
	{byte(protocol.CONNECT), 1 << 4},                // 2^4
	{byte(protocol.CONNACK), 1 << 5},                // 2 ^5
	{byte(protocol.PUBLISH), 1<<4 + 1<<5},           // 2^4 + 2^5
	{byte(protocol.PUBACK), 1 << 6},                 // 2^6
	{byte(protocol.PUBREC), 1<<6 + 1<<4},            // 2^6 + 2^4
	{byte(protocol.PUBREL), 1<<6 + 1<<5},            // 2^6 + 2^5
	{byte(protocol.PUBCOMP), 1<<6 + 1<<5 + 1<<4},    // 2^6 + 2^5 + 2^4
	{byte(protocol.SUBSCRIBE), 1 << 7},              // 2^7
	{byte(protocol.SUBACK), 1<<7 + 1<<4},            // 2^7 + 2^4
	{byte(protocol.UNSUBSCRIBE), 1<<7 + 1<<5},       // 2^7 + 2^5
	{byte(protocol.UNSUBACK), 1<<7 + 1<<5 + 1<<4},   // 2^7 + 2^5 + 2^4
	{byte(protocol.PINGREQ), 1<<7 + 1<<6},           // 2^7 + 2^6
	{byte(protocol.PINGRESP), 1<<7 + 1<<6 + 1<<4},   // 2^7 + 2^6 + 2^4
	{byte(protocol.DISCONNECT), 1<<7 + 1<<6 + 1<<5}, // 2^7 + 2^6 + 2^5
}

func TestEncodePacketType(t *testing.T) {
	for i, tt := range PacketType {
		s := utils.EncodePacketType(tt.in)
		if s != tt.out {
			t.Errorf("%d => %d, wanted: %d", i, int(s), int(tt.out))
		}
	}
}

func TestDecodePacketType(t *testing.T) {
	for i, tt := range PacketType {
		for _, u := range []byte{1, 2, 4, 8} {
			s := utils.DecodePacketType(tt.out + u) // 加上 1 ~ 3 bit 位的标识也应该正确识别
			if s != tt.in {
				t.Errorf("%d => %d, wanted: %d", i, int(s), int(tt.in))
			}
		}
	}
}
