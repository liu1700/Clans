package netWorking

import (
	"Clans/server/log"
	"Clans/server/services"
	"Clans/server/utils"
	"errors"
	"fmt"
	"time"

	"net"

	kcp "github.com/xtaci/kcp-go"
)

type ClientHandler func(conn net.Conn, conf *services.Config)

type Server struct {
	ServiceConnGroups map[string]map[string]net.Conn // 服务类型 -> 服务实例id -> 具体链接
	Config            *services.Config
	Clients           map[uint32]*Session
}

func (s *Server) InitServer(conf *services.Config) {
	s.ServiceConnGroups = make(map[string]map[string]net.Conn)
	s.Config = conf
}

func (s *Server) TcpServer(handleClient ClientHandler) {
	config := s.Config

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

func (s *Server) UdpServer(handleClient ClientHandler) {
	config := s.Config

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

func (s *Server) AddService(ip string, port int, serviceName string, serviceId string) (net.Conn, error) {
	retry, retryMax, retryDuration := 0, 5, time.Duration(2)

	for retry < retryMax {
		conn, err := kcp.Dial(fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			log.Logger().Warnf("dial kcp service ip:%s, port:%d, for %d time(s), reason:%s", ip, port, retry, err.Error())
			retry++
			time.Sleep(retryDuration)
		} else {
			serviceGroup := s.ServiceConnGroups[serviceName]
			if serviceGroup == nil {
				serviceGroup = make(map[string]net.Conn)
			}
			serviceGroup[serviceId] = conn

			s.ServiceConnGroups[serviceName] = serviceGroup
			return conn, nil
		}
	}

	return nil, errors.New("dial remote server err")
}
