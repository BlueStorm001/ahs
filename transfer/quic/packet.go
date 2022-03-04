// Copyright 2021 The Go wdw. All rights reserved.
// Use of this source code is governed by a BSD-style

// Package quic sequences and encoding and decoding of Packer.
//
package quic

import (
	"bytes"
	"encoding/binary"
)

type Packer struct {
	headLen     int
	datagram    []byte
	temp        bytes.Buffer
	tempLen     int
	tempHead    bytes.Buffer
	tempHeadLen int
	addHead     bool
}

func NewPacker() *Packer {
	return &Packer{headLen: 8}
}

func (pk Packer) Write(buffer []byte) []byte {
	body := bytes.Buffer{}
	body.WriteByte(1)
	body.WriteByte(2)
	body.WriteByte(3)
	body.WriteByte(0) //保留
	body.Write(intToBytes(len(buffer)))
	body.Write(buffer)
	return body.Bytes()
}

func (pk *Packer) Packet(receive func(buffer []byte)) {
	if receive == nil {
		return
	}
	if pk.tempHeadLen > 0 {
		if len(pk.datagram) >= pk.tempHeadLen {
			pack := pk.datagram[:pk.tempHeadLen]
			//包头收完
			pk.tempHead.Write(pack)
			//剩余包
			spk := pk.datagram[len(pack):]
			pk.addHead = true
			if len(spk) == 0 {
				pk.tempHeadLen = 0
				return
			} else { //多了
				pk.datagram = spk
				pk.tempHeadLen = 0
			}
		} else { //继续收
			pk.tempHeadLen = pk.tempHeadLen - len(pk.datagram)
			pk.tempHead.Write(pk.datagram)
			pk.datagram = pk.datagram[:0]
			return
		}
	}
	//数据加包头
	if pk.addHead {
		pk.addHead = false
		pk.tempHead.Write(pk.datagram)
		pk.datagram = pk.tempHead.Bytes()
		pk.tempHead.Reset()
	}
	//剩余包体
	if pk.tempLen > 0 {
		if len(pk.datagram) >= pk.tempLen {
			pk.temp.Write(pk.datagram[:pk.tempLen])
			go receive(pk.temp.Bytes())
			pk.temp.Reset()
			pk.datagram = pk.datagram[pk.tempLen:]
			pk.tempLen = 0
		} else { //还不够
			pk.tempLen = pk.tempLen - len(pk.datagram)
			pk.temp.Write(pk.datagram)
			pk.datagram = pk.datagram[:0]
			return
		}
	}
	for {
		length := len(pk.datagram)
		if length == 0 {
			return
		}
		//处理包头
		if length < pk.headLen {
			pk.tempHeadLen = pk.headLen - length
			pk.tempHead.Write(pk.datagram)
			pk.datagram = pk.datagram[:0]
			return
		}
		//开始
		if pk.datagram[0] == 1 && pk.datagram[1] == 2 && pk.datagram[2] == 3 {
			dataCount := bytesToInt(pk.datagram[pk.headLen-4 : pk.headLen])
			remaining := length - pk.headLen
			if dataCount <= remaining {
				go receive(pk.datagram[pk.headLen : pk.headLen+dataCount])
				pk.datagram = pk.datagram[pk.headLen+dataCount:]
			} else { //断包加包
				pk.temp.Write(pk.datagram[pk.headLen:])
				pk.tempLen = dataCount - remaining
				break
			}
		} else {
			//断包
			break
		}
	}

	pk.datagram = pk.datagram[:0]
}

func bytesToInt(b []byte) int {
	var x int32
	binary.Read(bytes.NewBuffer(b), binary.BigEndian, &x)
	return int(x)
}

func intToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
