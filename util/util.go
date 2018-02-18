package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func Echo(args ...interface{}) {
	fmt.Println(args...)
}

/*
func f(args ...interface{}) {
	fmt.Fprintln(datafile, args...)
}
*/

func JsonString(t interface{}) string {
	j, _ := json.Marshal(t)
	return string(j)
}

func JsonByte(t interface{}) []byte {
	str := JsonString(t)
	return []byte(str)
}

func Time() int64 {
	return time.Now().UnixNano() / 1000000
}

func Md5(data []byte) string {
	h := md5.New()
	h.Write(data)
	key := hex.EncodeToString(h.Sum(nil))
	return key
}
