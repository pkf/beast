package util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

//php -r 'echo  strlen(sha1("/KVRPelPJ/ByC6WNu1ncGg==258EAFA5-E914-47DA-95CA-C5AB0DC85B11", true));'
func Sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	ret := fmt.Sprintf("%s", t.Sum(nil))
	return ret
}

func Ord(b byte) rune {
	r := []rune(string(b))
	return r[0]
}

func GetCurrentPath() string {
	s, _ := exec.LookPath(os.Args[0])
	sep := string(os.PathSeparator)
	i := strings.LastIndex(s, sep)
	path := string(s[0 : i+1])
	return path
}

func FileGetContents(file string) []byte {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	} else {
		defer f.Close()
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return data
}
