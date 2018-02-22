package util

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	_ "fmt"
)

//php -r 'echo var_dump(pack("n", 97));'
func Packn(i int) string {
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian

	binary.Write(buf, byteOrder, uint16(i))

	//fmt.Printf("pack: %v\n", string(buf.Bytes()))
	//fmt.Printf("pack: %v\n", len(string(buf.Bytes())))
	return string(buf.Bytes())
}

//php -r 'echo var_dump(pack("xxxxN", 97));'
func PackxxxxN(i int) string {
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian
	binary.Write(buf, byteOrder, "\\0")
	binary.Write(buf, byteOrder, "\\0")
	binary.Write(buf, byteOrder, "\\0")
	binary.Write(buf, byteOrder, "\\0")
	/*
		var data = []interface{}{
			uint8(0),
			uint8(0),
		    uint8(0),
		    uint8(0),
		    uint32(i),
		}
	*/
	binary.Write(buf, byteOrder, uint32(i))

	//fmt.Printf("pack: %v\n", string(buf.Bytes()))
	//fmt.Printf("pack: %v\n", len(string(buf.Bytes())))
	return string(buf.Bytes())
}

//php -r 'echo var_dump(pack("H*", "6578616d706c65206865782064617461"));'
//php -r 'echo var_dump(hex2bin("6578616d706c65206865782064617461"));'
//see https://github.com/imroc/biu
func PackH(s string) string {
	//buf := new(bytes.Buffer)
	//byteOrder := binary.BigEndian
	//str:= hex.EncodeToString([]byte(s))
	//binary.Write(buf, byteOrder, []byte(str))
	//binary.Write(buf, byteOrder, string("6578616d706c65206865782064617461"))

	a, _ := hex.DecodeString(s)

	return string(a)
}
