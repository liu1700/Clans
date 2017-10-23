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
	IP net.IP
	// MQ      chan *netPackages.NetPackage // 返回给客户端的异步消息
	Encoder       *rc4.Cipher // 加密器
	Decoder       *rc4.Cipher // 解密器
	UserId        uint32      // 玩家ID
	GameServiceId int         // 游戏服ID;游戏服的id
	// Stream  net.Conn                     // 后端游戏服数据流
	GameService *Service      // 游戏服务,也就是房间服务
	Die         chan struct{} // 会话关闭信号
	OutBuffer   *Buffer       // 写回数据用的buffer
	ServerInst  *Server       // 服务实例

	// 会话标记
	Flag int32

	// 时间相关
	ConnectTime    time.Time // TCP链接建立时间
	PacketTime     time.Time // 当前包的到达时间
	LastPacketTime time.Time // 前一个包到达时间

	PacketCount     uint32 // 对收到的包进行计数，避免恶意发包
	PacketCount1Min int    // 每分钟的包统计，用于RPM判断
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
