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
	PlayerId uint8
	SrcDatas []byte // [optype1, d11, d12, splitflag, optype2, d21]
}

func GetFramePackageData(byteSlice []byte) []byte {
	sizeStart := uint32(1)
	sizeEnd := sizeStart + 2
	sz := binary.BigEndian.Uint16(byteSlice[sizeStart:sizeEnd])
	// 删掉空白数据
	return byteSlice[:sz+3]
}

// 上报的数据帧不包含帧id，逻辑帧id需要服务器控制
func BytesToFramePackage(byteSlice []byte) (*FramePackage, error) {

	// frameIdStart := uint32(1)
	// frameIdEnd := frameIdStart + 4

	// frameId := binary.BigEndian.Uint32(byteSlice[frameIdStart:frameIdEnd])

	// fmt.Printf("frame ID %d \n", frameId)

	// pidIndex := frameIdEnd

	pid := uint8(byteSlice[0])
	// fmt.Printf("Pid %d \n", pid)

	// 数组为引用类型
	cloneData := make([]byte, len(byteSlice))
	copy(cloneData, byteSlice)
	pack := &FramePackage{
		PlayerId: pid,
		SrcDatas: cloneData,
	}
	return pack, nil
}
