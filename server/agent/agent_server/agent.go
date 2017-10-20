package agent_server

import (
	"Clans/server/log"
	"Clans/server/netPackages"
	"sync"
	"time"
)

var (
	rpmLimit = 0
)

// PIPELINE #2: agent
// all the packets from handleClient() will be handled
func Agent(sess *Session, shuttingDownChan chan struct{}, wg *sync.WaitGroup, in chan *netPackages.NetPackage, out *Buffer) {
	defer wg.Done() // will decrease waitgroup by one, useful for manual server shutdown

	// init session
	// sess.MQ = make(chan pb.Game_Frame, 512)
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()

	// minute timer
	min_timer := time.After(time.Minute)

	// cleanup work
	defer func() {
		close(sess.Die)
		// if sess.Stream != nil {
		// 	sess.Stream.CloseSend()
		// }
	}()

	// >> the main message loop <<
	// handles 4 types of message:
	//  1. from client
	//  2. from game service
	//  3. timer
	//  4. server shutdown signal
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}

			sess.PacketCount++
			sess.PacketCount1Min++
			sess.PacketTime = time.Now()

			if result := route(sess, msg); result != nil && len(result) > 0 {
				out.send(sess, result)
			}
			sess.LastPacketTime = sess.PacketTime
		// case frame := <-sess.MQ: // packets from game
		// 	switch frame.Type {
		// 	case pb.Game_Message:
		// 		out.send(sess, frame.Message)
		// 	case pb.Game_Kick:
		// 		sess.Flag |= SESS_KICKED_OUT
		// 	}
		case <-min_timer: // minutes timer
			timerWork(sess, out)
			min_timer = time.After(time.Minute)
		case <-shuttingDownChan: // server is shuting down...
			sess.Flag |= SESS_KICKED_OUT
		}

		// see if the player should be kicked out.
		if sess.Flag&SESS_KICKED_OUT != 0 {
			return
		}
	}
}

func SetRpmLimit(limit int) {
	rpmLimit = limit
}

// 玩家1分钟定时器
func timerWork(sess *Session, out *Buffer) {
	defer func() {
		sess.PacketCount1Min = 0
	}()

	// 发包频率控制，太高的RPS直接踢掉
	if sess.PacketCount1Min > rpmLimit {
		sess.Flag |= SESS_KICKED_OUT
		log.Logger().Errorf("userid %d, packet in 1m %d, total %d", sess.UserId, sess.PacketCount1Min, sess.PacketCount)
		return
	}
}
