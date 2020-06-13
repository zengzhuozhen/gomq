package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func StructToBytes(x interface{}) []byte {
	buf := new(bytes.Buffer)   // 创建一个buffer区
	enc := gob.NewEncoder(buf) // 创建新的需要转化二进制区域对象
	err := enc.Encode(x)       // 将数据转化为二进制流
	if err != nil {
		fmt.Println(err)
		panic("数据压包出错")
	}
	b := buf.Bytes() // 将二进制流赋值给变量b
	return b
}

func BytesToStruct(b []byte, x interface{}) interface{} {
	dec := gob.NewDecoder(bytes.NewBuffer(b)) // 创建一个对象 把需要转化的对象放入
	err := dec.Decode(x)                     // 进行流转化
	if err != nil {
		fmt.Println("gob decode failed, err", err)
		panic("数据解包出错")
	}
	return x
}

func EncodeString(field string) []byte {
	return encodeBytes([]byte(field))
}

func DecodeString(b io.Reader) (string, error) {
	buf, err := decodeBytes(b)
	return string(buf), err
}

func encodeBytes(field []byte) []byte {
	fieldLength := make([]byte, 2)
	binary.BigEndian.PutUint16(fieldLength, uint16(len(field)))
	return append(fieldLength, field...)
}


func decodeBytes(b io.Reader) ([]byte, error) {
	fieldLength, err := DecodeUint16(b)
	if err != nil {
		return nil, err
	}
	field := make([]byte, fieldLength)
	_, err = b.Read(field)
	if err != nil {
		return nil, err
	}
	return field, nil
}

func EncodeByte(b byte) []byte{
	return []byte{b}
}

func DecodeByte(b io.Reader) (byte, error) {
	temp := make([]byte, 1)
	_, err := b.Read(temp)
	if err != nil {
		return 0, nil
	}
	return 	temp[0] ,err
}


func EncodeUint64(field uint64) []byte{
	fieldLength := make([]byte,8)
	binary.BigEndian.PutUint64(fieldLength, field)
	return fieldLength
}

func DecodeUint64(b io.Reader) (uint64, error) {
	num := make([]byte, 8)
	_, err := b.Read(num)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(num), nil
}

func EncodeUint16(num uint16) []byte {
	bytesResult := make([]byte, 2)
	binary.BigEndian.PutUint16(bytesResult, num)
	return bytesResult
}

func DecodeUint16(b io.Reader) (uint16, error) {
	num := make([]byte, 2)
	_, err := b.Read(num)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(num), nil
}


