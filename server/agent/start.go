package agent

import (
	"Clans/server/agent/agent_server"
	"Clans/server/log"
	"encoding/binary"
	"io"
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

	log.Info("listening on:", listener.Addr())

	// loop accepting
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Warning("accept failed:", err)
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
	log.Info("udp listening on:", l.Addr())
	lis := l.(*kcp.Listener)

	if err := lis.SetReadBuffer(config.Sockbuf); err != nil {
		log.Println("SetReadBuffer", err)
	}
	if err := lis.SetWriteBuffer(config.Sockbuf); err != nil {
		log.Println("SetWriteBuffer", err)
	}
	if err := lis.SetDSCP(config.Dscp); err != nil {
		log.Println("SetDSCP", err)
	}

	// loop accepting
	for {
		conn, err := lis.AcceptKCP()
		if err != nil {
			log.Warning("accept failed:", err)
			continue
		}
		// set kcp parameters
		conn.SetWindowSize(config.Sndwnd, config.Rcvwnd)
		conn.SetNoDelay(config.Nodelay, config.Interval, config.Resend, config.Nc)
		conn.SetStreamMode(true)
		conn.SetMtu(config.Mtu)

		// start a goroutine for every incoming connection for reading
		go handleClient(conn, config)
	}
}

// PIPELINE #1: handleClient
// the goroutine is used for reading incoming PACKETS
// each packet is defined as :
// | 2B size |     DATA       |
//
func handleClient(conn net.Conn, config *Config) {
	defer conn.Close()
	// for reading the 2-Byte header
	header := make([]byte, 2)
	// the input channel for agent()
	in := make(chan []byte)
	defer func() {
		close(in) // session will close
	}()

	// create a new session object for the connection
	// and record it's IP address
	var sess agent_server.Session
	host, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Error("cannot get remote address:", err)
		return
	}
	sess.IP = net.ParseIP(host)
	log.Infof("new connection from:%v port:%v", host, port)

	// session die signal, will be triggered by agent()
	sess.Die = make(chan struct{})

	// create a write buffer
	out := agent_server.NewBuffer(conn, sess.Die, config.Txqueuelen)
	go out.Start()

	// start agent for PACKET processing
	Wg.Add(1)
	go agent_server.Agent(&sess, shuttingDownChan, &Wg, in, out)

	// read loop
	for {
		// solve dead link problem:
		// physical disconnection without any communcation between client and server
		// will cause the read to block FOREVER, so a timeout is a rescue.
		conn.SetReadDeadline(time.Now().Add(config.ReadDeadline))

		// read 2B header
		n, err := io.ReadFull(conn, header)
		if err != nil {
			log.Warningf("read header failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}
		size := binary.BigEndian.Uint16(header)

		// alloc a byte slice of the size defined in the header for reading data
		payload := make([]byte, size)
		n, err = io.ReadFull(conn, payload)
		if err != nil {
			log.Warningf("read payload failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}

		// deliver the data to the input queue of agent()
		select {
		case in <- payload: // payload queued
		case <-sess.Die:
			log.Warningf("connection closed by logic, flag:%v ip:%v", sess.Flag, sess.IP)
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
			log.Info("sigterm received")
			log.Info("waiting for agents close, please wait...")
			log.Info("agent shutdown.")
			wg.Done()
		}
	}
}

func Start(config *Config) {
	Wg.Add(1)

	go sig_handler(&Wg)

	log.InitLogger(log.DEV)

	agent_server.SetRpmLimit(config.RpmLimit)

	// listeners
	go tcpServer(config)
	go udpServer(config)

	Wg.Wait()

	os.Exit(0)
	return
}
