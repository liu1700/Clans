package netWorking

import (
	"Clans/server/log"
	"Clans/server/netPackages"
	"crypto/rc4"
	"net"
	"time"
)

const (
	SESS_KEYEXCG    = 0x1 // 是否已经交换完毕KEY
	SESS_ENCRYPT    = 0x2 // 是否可以开始加密
	SESS_KICKED_OUT = 0x4 // 踢掉
	SESS_AUTHORIZED = 0x8 // 已授权访问
)

type Session struct {
	IP            net.IP
	MQ            chan []byte // 返回给客户端的异步消息
	Encoder       *rc4.Cipher // 加密器
	Decoder       *rc4.Cipher // 解密器
	UserId        uint32      // 玩家ID
	RoomId        uint32      // 玩家所处的房间id
	GameServiceId int         // 游戏服ID;游戏服的id
	Stream        net.Conn    // 后端房间服数据流
	// GameService *Service      // 游戏服务,也就是房间服务
	Die        chan struct{} // 会话关闭信号
	OutBuffer  *Buffer       // 写回数据用的buffer
	ServerInst *Server       // 服务实例

	// 会话标记
	Flag int32

	// 时间相关
	ConnectTime    time.Time // 链接建立时间
	PacketTime     time.Time // 当前包的到达时间
	LastPacketTime time.Time // 前一个包到达时间

	PacketCount     uint32 // 对收到的包进行计数，避免恶意发包
	PacketCount1Min int    // 每分钟的包统计，用于RPM判断
}

func (s *Session) JoinRoom(ip string, port int) {
	// stream, err := s.ServerInst.DialRoom(ip, port)
	// if err != nil {
	// 	log.Logger().Error("error when dial room stream err:", err.Error())
	// 	return
	// }

	// s.Stream = stream

	// go func() {
	// 	// read loop
	// 	readBytes := make([]byte, netPackages.DISPATCH_FRAME_PACKET_LIMIT)
	// 	for {
	// 		// solve dead link problem:
	// 		// physical disconnection without any communcation between client and server
	// 		// will cause the read to block FOREVER, so a timeout is a rescue.
	// 		s.Stream.SetReadDeadline(time.Now().Add(time.Minute * 5))

	// 		// alloc a byte slice of the size defined in the header for reading data
	// 		n, err := s.Stream.Read(readBytes)
	// 		if err != nil {
	// 			log.Logger().Errorf("read readBytes from room failed, ip:%v reason:%v size:%v", ip, err, n)
	// 			return
	// 		}

	// 		// deliver the data to the input queue of agent()
	// 		select {
	// 		case s.MQ <- netPackages.GetFramePackageData(readBytes): // payload queued
	// 		case <-s.Die:
	// 			log.Logger().Warnf("connection closed by logic, flag:%v session ip:%v", s.Flag, s.IP)
	// 			s.LeaveRoom()
	// 			return
	// 		}
	// 	}
	// }()
}

func (s *Session) LeaveRoom() {
	// if err := s.Stream.Close(); err != nil {
	// 	log.Logger().Error("error when leave room err ", err.Error())
	// }
}

func (s *Session) Push(data []byte) error {
	// if s.Stream != nil {
	// 	if _, err := s.Stream.Write(data); err != nil {
	// 		log.Logger().Error("error when push data to game room err: ", err.Error())
	// 		return err
	// 	}
	// 	return nil
	// } else {
	// 	log.Logger().Error("nil stream")
	// 	return errors.New("nil stream")
	// }
	return nil
}

// packet sending procedure
func (s *Session) Write(respId int, srcPack *netPackages.NetPackage, sendData []byte) {
	// encryption
	// (NOT_ENCRYPTED) -> KEYEXCG -> ENCRYPT
	if s.Flag&SESS_ENCRYPT != 0 { // encryption is enabled
		s.Encoder.XORKeyStream(sendData, sendData)
	} else if s.Flag&SESS_KEYEXCG != 0 { // key is exchanged, encryption is not yet enabled
		s.Flag &^= SESS_KEYEXCG
		s.Flag |= SESS_ENCRYPT
	}

	// 更新数据包的数据内容
	srcPack.HandlerId = byte(respId)
	// srcPack.Version = version
	srcPack.Data = sendData

	// queue the data for sending
	select {
	case s.OutBuffer.pending <- srcPack.Bytes():
	default: // packet will be dropped if txqueuelen exceeds
		log.Logger().Warnf("userid %d ip %s", s.UserId, s.IP)
	}
	return
}

func (s *Session) GetService(serviceName int, serviceId int) *Service {
	return s.ServerInst.GetService(serviceName, serviceId)
}
