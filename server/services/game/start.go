package game

import (
	"Clans/server/db"
	"Clans/server/log"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/services"
	"Clans/server/services/game/game_server"
	"Clans/server/utils"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	Wg               sync.WaitGroup
	shuttingDownChan = make(chan struct{})
	Version          int
)

func InitDBTables() {

}

func handleClient(conn net.Conn, s *netWorking.Server) {
	defer conn.Close()
	// the input channel for agent()
	in := make(chan *netPackages.FramePackage)

	config := s.Config

	// create a new session object for the connection
	sess := new(netWorking.Session)

	sess.MQ = make(chan []byte, 512)

	// and record it's IP address
	// var sess netWorking.Session
	host, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Logger().Error("cannot get remote address:", err)
		return
	}
	sess.IP = net.ParseIP(host)
	log.Logger().Infof("new connection from:%v port:%v", host, port)

	// session die signal, will be triggered by agent()
	sess.Die = make(chan struct{})

	// create a write buffer
	out := netWorking.NewBuffer(conn, sess.Die, config.Txqueuelen)
	go out.Start()

	// start agent for PACKET processing
	Wg.Add(1)
	go game_server.Agent(sess, shuttingDownChan, &Wg, in, out)

	sess.OutBuffer = out

	id := atomic.AddUint64(&s.ClientsId, 1)

	s.Clients[id] = sess

	sess.ServerInst = s

	// // 获取游戏逻辑服务
	// service := s.GetService(flats.PacketIdGame, 1)
	// sess.GameService = service

	defer func() {
		close(in) // session will close
		delete(s.Clients, id)
	}()

	// read loop
	readBytes := make([]byte, netPackages.UPLOAD_FRAME_PACKET_LIMIT)
	for {
		// solve dead link problem:
		// physical disconnection without any communcation between client and server
		// will cause the read to block FOREVER, so a timeout is a rescue.
		conn.SetReadDeadline(time.Now().Add(config.ReadDeadline))

		// alloc a byte slice of the size defined in the header for reading data
		n, err := conn.Read(readBytes)
		if err != nil {
			log.Logger().Errorf("read readBytes failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}

		// 解密
		if sess.Flag&netWorking.SESS_ENCRYPT != 0 {
			sess.Decoder.XORKeyStream(readBytes, readBytes)
		}

		// 转化为package
		payload, err := netPackages.BytesToFramePackage(readBytes)
		if err != nil {
			log.Logger().Errorf("read payload faild, err :%v", err.Error())
			return
		}

		log.Logger().Debugf("frame payload %+v \n", *payload)

		// deliver the data to the input queue of agent()
		select {
		case in <- payload: // payload queued
		case <-sess.Die:
			log.Logger().Warnf("connection closed by logic, flag:%v ip:%v", sess.Flag, sess.IP)
			return
		}
	}
}

func Start(config *services.Config) {
	Wg.Add(1)
	log.InitLogger(log.DEV)

	db.InitDB("139.162.96.106", 3306, "root", "root", "runaway")
	db.CheckConnecting()

	InitDBTables()

	go utils.SigHandler(&Wg, shuttingDownChan)

	server := new(netWorking.Server)
	server.InitServer(config)

	// listeners
	go server.UdpServer(handleClient)

	game_server.InitDispatcher(server)

	Wg.Wait()

	os.Exit(0)
	return
}
