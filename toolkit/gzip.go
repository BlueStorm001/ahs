package toolkit

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// GzipCompress Gzip压缩  Base64 string
func GzipCompress(input string) string {
	return ByteToStr(GzipCompressBase64Bytes(StrToByte(input)))
}

// GzipCompressBase64Bytes Gzip压缩  Base64 bytes
func GzipCompressBase64Bytes(input []byte) []byte {
	return ToBase64EncodeBytes(GzipCompressBytes(input))
}

// GzipCompressBytes Gzip压缩 bytes
func GzipCompressBytes(input []byte) []byte {
	var buf = &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	leng, err := w.Write(input)
	if err != nil || leng == 0 {
		return nil
	}
	err = w.Flush()
	if err != nil {
		return nil
	}
	err = w.Close()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// GzipUnCompress Gzip解压缩 Base64 string
func GzipUnCompress(input string) string {
	return ByteToStr(GzipUnCompressBase64Bytes(StrToByte(input)))
}

// GzipUnCompressBase64Bytes Gzip解压缩 Base64 bytes
func GzipUnCompressBase64Bytes(src []byte) []byte {
	b, err := ToBase64DecodeBytes(src)
	if err != nil {
		return nil
	}
	return GzipUnCompressBytes(b)
}

// GzipUnCompressBytes Gzip解压缩
func GzipUnCompressBytes(input []byte) []byte {
	if input == nil {
		return nil
	}
	out := bytes.NewBuffer(input)
	r, _ := gzip.NewReader(out)
	defer r.Close()
	un, _ := ioutil.ReadAll(r)
	return un
}
