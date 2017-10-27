package agent_server

import (
	"Clans/server/flats"
	"Clans/server/log"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
)

// route client protocol
func route(sess *netWorking.Session, pack *netPackages.NetPackage) {
	if pack != nil {
		// 读客户端数据包序列号(1,2,3...)
		// 客户端发送的数据包必须包含一个自增的序号，必须严格递增
		// 加密后，可避免重放攻击-REPLAY-ATTACK
		// 数据包序列号验证
		if pack.SeqId != sess.PacketCount {
			log.Logger().Errorf("illegal packet sequence id:%v should be:%v size:%v", pack.SeqId, sess.PacketCount, len(pack.Data))
			sess.Flag |= netWorking.SESS_KICKED_OUT
			return
		}

		// 根据协议号断做服务划分
		// 协议号的划分采用分割协议区间, 用户可以自定义多个区间，用于转发到不同的后端服务
		if pack.PacketId == uint8(flats.PacketIdGame) {
			// if sess.GameService != nil {
			// 	sess.GameService.Forward(pack.Data)
			// } else {
			// 	sess.Flag |= netWorking.SESS_KICKED_OUT
			// 	return
			// }
			// if err := sess.Push(pack.Data); err != nil {
			// 	sess.Flag |= netWorking.SESS_KICKED_OUT
			// 	return
			// }
		} else if pack.PacketId > uint8(flats.PacketIdGame) {

		} else {
			if h := ReqHandler[pack.HandlerId]; h != nil {
				log.Logger().Debugf("processing request id %d ,name %s", pack.HandlerId, flats.EnumNamesRequestId[int(pack.HandlerId)])
				h(sess, pack)
				log.Logger().Debugf("finishing last request id %d ,name %s", pack.HandlerId, flats.EnumNamesResponseId[int(pack.HandlerId)])
			}
			// if h := client_handler.Handlers[b]; h != nil {
			// 	ret = h(sess, reader)
			// } else {
			// 	log.Errorf("service id:%v not bind", b)
			// 	sess.Flag |= SESS_KICKED_OUT
			// 	return nil
			// }
		}

		// elasped := time.Now().Sub(start)
		if pack.PacketId != 0 { // 排除心跳包日志
			// log.WithFields(log.Fields{"cost": elasped,
			// 	"api":  client_handler.RCode[b],
			// 	"code": b}).Debug("REQ")
		}

		return
	} else {
		log.Logger().Error("nil pack")
	}
	return
}
