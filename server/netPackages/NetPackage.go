package netPackages

import (
	"Clans/server/flats"
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	PACKET_LIMIT = 1024
)

type NetPackage struct {
	PacketId  uint8
	Version   uint8
	SeqId     uint32
	HandlerId uint8
	Data      []byte
}

func (p *NetPackage) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, p.PacketId)          // packet id 包类型
	binary.Write(buf, binary.BigEndian, p.Version)           // 版本号
	binary.Write(buf, binary.BigEndian, p.SeqId)             // seq id
	binary.Write(buf, binary.BigEndian, p.HandlerId)         // 如果是请求的话就是RequestId,返回的话就是ResponseId
	binary.Write(buf, binary.BigEndian, uint32(len(p.Data))) // 写入数据包长度
	binary.Write(buf, binary.BigEndian, p.Data)              // 数据内容

	return buf.Bytes()
}

func HeartBeatPacket() []byte {
	return []byte{flats.PacketIdHeartBeat}
}

func BytesToNetPackage(byteSlice []byte) (*NetPackage, error) {
	// minimalPackageSize := uint32(12)
	// length := uint32(len(byteSlice))
	// if length < minimalPackageSize {
	// 	return nil, errors.New(fmt.Sprintf("Data size  %d is less than mimial size ", length))
	// }

	// iSize := uint32(length - 1)
	// fmt.Println("iiiiiiiiiiiiiii")
	// fmt.Println(i)

	//seqStart := uint32(i + 1)
	//seqEnd := seqStart + 4

	//			seqNum := binary.BigEndian.Uint32(byteSlice[seqStart:seqEnd])
	//			fmt.Printf("Seq num%d \n", seqNum)

	packetIdStart := uint32(0)
	packetIdEnd := packetIdStart + 1

	// if handerIdStart > iSize || handerIdEnd > iSize {
	// 	// return nil, 0, errors.New(fmt.Sprintf(" handerIdStart > iSize || handerIdEnd > iSize"))
	// 	continue
	// }

	packetId := uint8(byteSlice[packetIdStart])

	//	fmt.Printf("Handler ID %d \n", handerId)

	verIndex := packetIdEnd

	// if verIndex > iSize {
	// 	// return nil, 0, errors.New(fmt.Sprintf(" verIndex > iSize"))
	// 	continue
	// }

	version := uint8(byteSlice[verIndex])
	//			fmt.Printf("Version %d \n", version)

	seqStart := verIndex + 1
	seqEnd := seqStart + 4
	seqData := binary.BigEndian.Uint32(byteSlice[seqStart:seqEnd])

	// 心跳包直接返回
	if packetId == flats.PacketIdHeartBeat {
		pack := &NetPackage{
			PacketId:  packetId,
			Version:   version,
			SeqId:     seqData,
			HandlerId: 0,
			Data:      []byte{},
		}
		return pack, nil
	}

	handlerIndex := seqEnd
	handlerId := uint8(byteSlice[handlerIndex])

	sizeStart := handlerIndex + 1
	sizeEnd := sizeStart + 4

	// if length < uint32(i)+minimalPackageSize {
	// 	// return nil, 0, errors.New(fmt.Sprintf("Data size %d is less than minimal size from prefix", length))
	// 	continue
	// }

	// if sizeStart > iSize || sizeEnd > iSize {
	// 	// return nil, 0, errors.New(fmt.Sprintf(" sizeStart > iSize || sizeEnd > iSize"))
	// 	continue
	// }

	dataSize := binary.BigEndian.Uint32(byteSlice[sizeStart:sizeEnd])

	//			fmt.Printf("Data size %d \n", dataSize)
	dataStart := sizeEnd
	dataEnd := dataStart + dataSize

	// if dataStart > iSize || dataEnd > iSize {
	// 	// return nil, 0, errors.New(fmt.Sprintf(" dataStart > iSize || dataEnd > iSize || crcStart > iSize || crcEnd > iSize "))
	// 	continue
	// }
	// fmt.Println("$$$$$$$$$$$$$$$$$$$")
	// fmt.Println(byteSlice)
	// fmt.Println(suffixIndex)
	// fmt.Println(byteSlice[suffixIndex])

	// 网络均为大端序传输
	fmt.Println("len ", len(byteSlice), " ds ", dataStart, " dn ", dataEnd)
	data := byteSlice[dataStart:dataEnd]

	// flatbuffer需要小端序数据
	for i := len(data)/2 - 1; i >= 0; i-- {
		opp := len(data) - 1 - i
		data[i], data[opp] = data[opp], data[i]
	}

	pack := &NetPackage{
		PacketId:  packetId,
		Version:   version,
		SeqId:     seqData,
		HandlerId: handlerId,
		Data:      data,
	}
	return pack, nil
}
