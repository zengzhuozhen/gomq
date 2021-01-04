package common

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMessage_PackAndUnPack(t *testing.T) {
	m := Message{
		MsgKey: "hello",
		Body:   "world",
	}
	data := m.Pack()
	var m1 Message
	m1 = *m1.UnPack(data)
	fmt.Println(m1)

	if !reflect.DeepEqual(m,m1){
		t.Errorf("m is not equal before pack and after unpack")
	}
}
