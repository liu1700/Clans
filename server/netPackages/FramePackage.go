package netPackages

import (
	"encoding/binary"
)

type OpType int8

const (
	SPLITFLAG OpType = -1
	MOVE      OpType = 1
	SHOOT     OpType = 2
)

type FramePackage struct {
	// FrameId  uint32
	PacketId uint8
	PlayerId uint8
	SeqId    uint32
	// SrcDatas []byte // [optype1, d11, d12, splitflag, optype2, d21]
	SrcDatas []byte
}

// func GetFramePackageData(byteSlice []byte) []byte {
// 	sizeStart := uint32(1)
// 	sizeEnd := sizeStart + 2
// 	sz := binary.BigEndian.Uint16(byteSlice[sizeStart:sizeEnd])
// 	// 删掉空白数据
// 	return byteSlice[:sz+3]
// }

// 上报的数据帧不包含帧id，逻辑帧id需要服务器控制
func BytesToFramePackage(byteSlice []byte) (*FramePackage, error) {

	// 包类型
	packetIdStart := uint32(0)
	packetIdEnd := packetIdStart + 1

	packetId := uint8(byteSlice[packetIdStart])

	// seqid
	seqStart := packetIdEnd
	seqEnd := seqStart + 4
	seqData := binary.BigEndian.Uint32(byteSlice[seqStart:seqEnd])
	// fmt.Println("seqId ", seqData)

	// 包大小
	sizeStart := seqEnd
	sizeEnd := sizeStart + 4
	dataSize := binary.BigEndian.Uint32(byteSlice[sizeStart:sizeEnd])
	// fmt.Println("size ", dataSize)

	// 数据
	dataStart := sizeEnd
	dataEnd := dataStart + dataSize
	// cloneData := byteSlice[dataStart:dataEnd]
	cloneData := make([]byte, dataSize)
	copy(cloneData, byteSlice[dataStart:dataEnd])
	// fmt.Println("cloneData ", cloneData)
	// pid := uint8(cloneData[0])

	pack := &FramePackage{
		PacketId: packetId,
		// PlayerId: pid,
		SeqId:    seqData,
		SrcDatas: cloneData,
	}
	return pack, nil
}
