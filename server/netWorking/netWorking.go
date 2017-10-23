package netWorking

import (
	"Clans/server/log"
	"Clans/server/services"
	"Clans/server/utils"

	"net"

	kcp "github.com/xtaci/kcp-go"
)

type ClientHandler func(conn net.Conn, conf *services.Config)

func TcpServer(config *services.Config, handleClient ClientHandler) {
	// resolve address & start listening
	tcpAddr, err := net.ResolveTCPAddr("tcp4", config.Listen)
	utils.CheckError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	utils.CheckError(err)

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

func UdpServer(config *services.Config, handleClient ClientHandler) {
	l, err := kcp.Listen(config.Listen)
	utils.CheckError(err)

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
