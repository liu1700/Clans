package agent

import (
	"Clans/server/db"
	"Clans/server/log"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/services"
	"Clans/server/services/agent/agent_server"
	"Clans/server/structs/users"
	"Clans/server/utils"
	"net"
	"os"
	"sync"
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
func handleClient(conn net.Conn, config *services.Config) {
	defer conn.Close()
	// the input channel for agent()
	in := make(chan *netPackages.NetPackage)
	defer func() {
		close(in) // session will close
	}()

	// create a new session object for the connection
	// and record it's IP address
	var sess netWorking.Session
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
	out := agent_server.NewBuffer(conn, sess.Die, config.Txqueuelen)
	go out.Start()

	// start agent for PACKET processing
	Wg.Add(1)
	go agent_server.Agent(&sess, shuttingDownChan, &Wg, in, out)

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
}

func Start(config *services.Config) {
	Wg.Add(1)
	log.InitLogger(log.DEV)

	db.InitDB("139.162.96.106", 3306, "root", "root", "runaway")
	db.CheckConnecting()

	InitDBTables()

	go utils.SigHandler(&Wg, shuttingDownChan)

	agent_server.SetRpmLimit(config.RpmLimit)

	// need get from db
	agent_server.SetVersion(1)

	// listeners
	go netWorking.TcpServer(config, handleClient)
	go netWorking.UdpServer(config, handleClient)

	Wg.Wait()

	os.Exit(0)
	return
}
