package netPackages

import (
	"fmt"
)

type OpType int8

const (
	SPLITFLAG OpType = -1
	MOVE      OpType = 1
	SHOOT     OpType = 2
)

type FramePackage struct {
	FrameId  uint32
	PlayerId uint8
	SrcDatas []byte // [optype1, d11, d12, splitflag, optype2, d21]
}

// 上报的数据帧不包含帧id，逻辑帧id需要服务器控制
func BytesToFramePackage(byteSlice []byte) (*FramePackage, error) {

	// frameIdStart := uint32(1)
	// frameIdEnd := frameIdStart + 4

	// frameId := binary.BigEndian.Uint32(byteSlice[frameIdStart:frameIdEnd])

	// fmt.Printf("frame ID %d \n", frameId)

	// pidIndex := frameIdEnd
	pidIndex := uint32(0)

	pid := uint8(byteSlice[pidIndex])
	fmt.Printf("Pid %d \n", pid)

	// 数组为引用类型
	cloneData := make([]byte, len(byteSlice))
	copy(cloneData, byteSlice)
	pack := &FramePackage{
		PlayerId: pid,
		SrcDatas: cloneData,
	}
	fmt.Println(cloneData)
	return pack, nil
}
