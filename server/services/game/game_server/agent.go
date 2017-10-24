package game_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"sync"
	"time"
)

var (
	rpmLimit = 0
)

// PIPELINE #2: agent
// all the packets from handleClient() will be handled
func Agent(sess *netWorking.Session, shuttingDownChan chan struct{}, wg *sync.WaitGroup, in chan *netPackages.NetPackage, out *netWorking.Buffer) {
	defer wg.Done() // will decrease waitgroup by one, useful for manual server shutdown

	// init session
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()

	// cleanup work
	defer func() {
		close(sess.Die)
		// if sess.Stream != nil {
		// 	sess.Stream.CloseSend()
		// }
	}()

	// >> the main message loop <<
	//  1. from client
	//  2. server shutdown signal
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}

			sess.PacketCount++
			sess.PacketCount1Min++
			sess.PacketTime = time.Now()
			sess.LastPacketTime = sess.PacketTime

			// 如果是心跳包则不处理,直接写回
			if msg.PacketId != flats.PacketIdHeartBeat {
				route(sess, msg)
			} else {
				out.RawSend(netPackages.HeartBeatPacket())
			}
		// case <-min_timer: // minutes timer
		// 	timerWork(sess)
		// 	min_timer = time.After(time.Minute)
		case <-shuttingDownChan: // server is shuting down...
			// broad cast shutting down
		}

		// see if the player should be kicked out.
		if sess.Flag&netWorking.SESS_KICKED_OUT != 0 {
			return
		}
	}
}

func SetRpmLimit(limit int) {
	rpmLimit = limit
}
