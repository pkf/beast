package util

import (
	"bytes"
	"encoding/binary"
)

//php -r 'echo printf("%b", pack("n", 89));'
func Packn(i int) string {
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian

	binary.Write(buf, byteOrder, uint16(i))
	//fmt.Printf("uint32: %x\n", buf.Bytes())

	return string(buf.Bytes())
}

//php -r 'echo printf("%b", pack("xxxxN", 89));'
func PackxxxxN(i int) string {
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian
	binary.Write(buf, byteOrder, uint8(0))
	binary.Write(buf, byteOrder, uint8(0))
	binary.Write(buf, byteOrder, uint8(0))
	binary.Write(buf, byteOrder, uint8(0))
	binary.Write(buf, byteOrder, uint32(i))
	//fmt.Printf("uint32: %x\n", buf.Bytes())

	return string(buf.Bytes())
}
