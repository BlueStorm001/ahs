package toolkit

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

var guid int64
var gmu sync.Mutex

func GUID() int64 {
	gmu.Lock()
	if guid == 0 {
		guid = time.Now().Unix()
	}
	guid++
	gmu.Unlock()
	return guid
}

func GUIDString() string {
	return strconv.FormatInt(GUID(), 10)
}

func BytePointerToStr(src *[]byte) (dst string) {
	s := (*reflect.SliceHeader)(unsafe.Pointer(src))
	d := (*reflect.StringHeader)(unsafe.Pointer(&dst))
	d.Data = s.Data
	d.Len = s.Len
	s.Data = 0
	s.Len = 0
	s.Cap = 0
	return
}

// ByteToStr nocopy 转换string
func ByteToStr(src []byte) (dst string) {
	return ByteSliceToStr(src, 0)
}

func BytesToBytes(src []byte, last int) (dst []byte) {
	s := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d := (*reflect.StringHeader)(unsafe.Pointer(&dst))
	if last <= 0 {
		d.Len = s.Len
	} else if last < s.Len {
		d.Len = last
	} else {
		s.Data = 0
		s.Len = 0
		s.Cap = 0
	}
	d.Data = s.Data
	return
}

// ByteSliceToStr nocopy 转换string
func ByteSliceToStr(src []byte, last int) (dst string) {
	s := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d := (*reflect.StringHeader)(unsafe.Pointer(&dst))
	if last <= 0 {
		d.Len = s.Len
	} else if last < s.Len {
		d.Len = last
	} else {
		s.Data = 0
		s.Len = 0
		s.Cap = 0
		return ""
	}
	d.Data = s.Data
	//Clear
	s.Data = 0
	s.Len = 0
	s.Cap = 0
	return
}

func ToBase64DecodeBytes(src []byte) ([]byte, error) {
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dbuf, src)
	return dbuf[:n], err
}

// StrToByte nocopy 转换[]byte
//注意:不能修改转换后的byte会panic，因字符串复制是只读的，转换后的byte无法修改也许无法panic，会导致系统崩溃
func StrToByte(src string) (dst []byte) {
	d := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	s := (*reflect.StringHeader)(unsafe.Pointer(&src))
	d.Data = s.Data
	d.Len = s.Len
	d.Cap = s.Len
	return
}

// ToBase64Encode ToBase64 string
func ToBase64Encode(src []byte) string {
	return ByteToStr(ToBase64EncodeBytes(src))
}

// ToBase64EncodeBytes ToBase64Bytes
func ToBase64EncodeBytes(src []byte) []byte {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(buf, src)
	return buf
}
