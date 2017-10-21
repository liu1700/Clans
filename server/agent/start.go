package agent

import (
	"Clans/server/agent/agent_server"
	"Clans/server/log"
	"Clans/server/netPackages"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kcp "github.com/xtaci/kcp-go"
)

const (
// SALT = "d601ea184522487c5182e0957a0cd23c"
)

var (
	Wg               sync.WaitGroup
	shuttingDownChan = make(chan struct{})

	Version int
)

type Config struct {
	Listen                        string
	ReadDeadline                  time.Duration
	Sockbuf                       int
	Udp_sockbuf                   int
	Txqueuelen                    int
	Dscp                          int
	Sndwnd                        int
	Rcvwnd                        int
	Mtu                           int
	Nodelay, Interval, Resend, Nc int
	RpmLimit                      int
}

func tcpServer(config *Config) {
	// resolve address & start listening
	tcpAddr, err := net.ResolveTCPAddr("tcp4", config.Listen)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	log.Logger().Info("listening on:", listener.Addr())

	// loop accepting
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Logger().Warn("accept failed:", err)
			continue
		}
		// set socket read buffer
		conn.SetReadBuffer(config.Sockbuf)
		// set socket write buffer
		conn.SetWriteBuffer(config.Sockbuf)
		// start a goroutine for every incoming connection for reading
		go handleClient(conn, config)
	}
}

func udpServer(config *Config) {
	l, err := kcp.Listen(config.Listen)
	checkError(err)
	log.Logger().Info("udp listening on:", l.Addr())
	lis := l.(*kcp.Listener)

	if err := lis.SetReadBuffer(config.Sockbuf); err != nil {
		log.Logger().Error("SetReadBuffer", err)
	}
	if err := lis.SetWriteBuffer(config.Sockbuf); err != nil {
		log.Logger().Error("SetWriteBuffer", err)
	}
	if err := lis.SetDSCP(config.Dscp); err != nil {
		log.Logger().Error("SetDSCP", err)
	}

	// loop accepting
	for {
		conn, err := lis.AcceptKCP()
		if err != nil {
			log.Logger().Warn("accept failed:", err)
			continue
		}
		// set kcp parameters
		conn.SetWindowSize(config.Sndwnd, config.Rcvwnd)
		conn.SetNoDelay(config.Nodelay, config.Interval, config.Resend, config.Nc)
		conn.SetStreamMode(true)
		conn.SetMtu(config.Mtu)

		log.Logger().Debug("accept kcp")

		// start a goroutine for every incoming connection for reading
		go handleClient(conn, config)
	}
}

// PIPELINE #1: handleClient
// the goroutine is used for reading incoming PACKETS
//
func handleClient(conn net.Conn, config *Config) {
	defer conn.Close()
	// the input channel for agent()
	in := make(chan *netPackages.NetPackage)
	defer func() {
		close(in) // session will close
	}()

	// create a new session object for the connection
	// and record it's IP address
	var sess agent_server.Session
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
		if sess.Flag&agent_server.SESS_ENCRYPT != 0 {
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

func checkError(err error) {
	if err != nil {
		panic(err)
		os.Exit(-1)
	}
}

// handle unix signals
func sig_handler(wg *sync.WaitGroup) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		msg := <-ch
		switch msg {
		case syscall.SIGTERM: // 关闭agent
			close(shuttingDownChan)
			log.Logger().Info("sigterm received")
			log.Logger().Info("waiting for agents close, please wait...")
			log.Logger().Info("agent shutdown.")
			wg.Done()
		}
	}
}

func Start(config *Config) {
	Wg.Add(1)

	go sig_handler(&Wg)

	log.InitLogger(log.DEV)

	agent_server.SetRpmLimit(config.RpmLimit)

	// need get from db
	agent_server.SetVersion(1)

	// listeners
	go tcpServer(config)
	go udpServer(config)

	Wg.Wait()

	os.Exit(0)
	return
}
