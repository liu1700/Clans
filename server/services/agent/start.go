package agent

import (
	"Clans/server/db"
	"Clans/server/log"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/services"
	"Clans/server/services/agent/agent_server"
	"Clans/server/structs/thirdParty"
	"Clans/server/structs/users"
	"Clans/server/utils"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
// SALT = "d601ea184522487c5182e0957a0cd23c"
)

var (
	Wg               sync.WaitGroup
	shuttingDownChan = make(chan struct{})

	Version int
)

// PIPELINE #1: handleClient
// the goroutine is used for reading incoming PACKETS
//
func handleClient(conn net.Conn, s *netWorking.Server) {
	defer conn.Close()
	// the input channel for agent()
	in := make(chan *netPackages.NetPackage)

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
	go agent_server.Agent(sess, shuttingDownChan, &Wg, in, out)

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
	readBytes := make([]byte, netPackages.PACKET_LIMIT)
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
		payload, err := netPackages.BytesToNetPackage(readBytes)
		if err != nil {
			log.Logger().Errorf("read payload faild, err :%v", err.Error())
			return
		}

		log.Logger().Debugf("%+v \n", *payload)

		// deliver the data to the input queue of agent()
		select {
		case in <- payload: // payload queued
		case <-sess.Die:
			log.Logger().Warnf("connection closed by logic, flag:%v ip:%v", sess.Flag, sess.IP)
			return
		}
	}
}

// 不同的服务器需要关心的结构体不同，所以将初始化表的操作拿到db之外
func InitDBTables() {
	// AutoMigrate 只做新增操作，不会修改原有数据，不修改旧记录的数据类型，不删除旧记录的无用字段
	db.DB().AutoMigrate(&users.User{})
	db.DB().AutoMigrate(&thirdParty.ServiceRecord{})
}

func Start(config *services.Config) {
	Wg.Add(1)
	log.InitLogger(log.DEV)

	db.InitDB("127.0.0.1", 3306, "root", "test", "runaway")
	db.CheckConnecting()

	InitDBTables()

	go utils.SigHandler(&Wg, shuttingDownChan)

	agent_server.SetRpmLimit(config.RpmLimit)

	// need get from db
	agent_server.SetVersion(1)

	server := new(netWorking.Server)
	server.InitServer(config)

	// listeners
	go server.TcpServer(handleClient)
	go server.UdpServer(handleClient)

	// 其他游戏服务监听
	// // 游戏主逻辑,也就是房间服务
	// server.AddService("192.168.1.102", 9080, flats.PacketIdGame, 1)

	Wg.Wait()

	os.Exit(0)
	return
}
