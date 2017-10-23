package netPackages

import (
	"Clans/server/flats"
	"bytes"
	"encoding/binary"
)

type ServicePackage struct {
	PacketId uint8
	Data     []byte
}

func (p *ServicePackage) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, p.PacketId)          // packet id 包类型
	binary.Write(buf, binary.BigEndian, uint32(len(p.Data))) // 写入数据包长度
	binary.Write(buf, binary.BigEndian, p.Data)              // 数据内容

	return buf.Bytes()
}

func BytesToServicePackage(byteSlice []byte) (*ServicePackage, error) {
	packetIdStart := uint32(0)
	packetIdEnd := packetIdStart + 1

	packetId := uint8(byteSlice[packetIdStart])

	// 心跳包直接返回
	if packetId == flats.PacketIdHeartBeat {
		pack := &ServicePackage{
			PacketId: packetId,
		}
		return pack, nil
	}

	sizeStart := packetIdEnd
	sizeEnd := sizeStart + 4

	dataSize := binary.BigEndian.Uint32(byteSlice[sizeStart:sizeEnd])

	dataStart := sizeEnd
	dataEnd := dataStart + dataSize

	// 透传数据不需要更改端序
	data := byteSlice[dataStart:dataEnd]

	// // flatbuffer需要小端序数据
	// for i := len(data)/2 - 1; i >= 0; i-- {
	// 	opp := len(data) - 1 - i
	// 	data[i], data[opp] = data[opp], data[i]
	// }

	pack := &ServicePackage{
		PacketId: packetId,
		Data:     data,
	}
	return pack, nil
}
