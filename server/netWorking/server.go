package netWorking

import (
	"Clans/server/flats"
	"Clans/server/log"
	"Clans/server/services"
	"Clans/server/utils"
	"errors"
	"fmt"
	"time"

	"net"

	kcp "github.com/xtaci/kcp-go"
)

type ClientHandler func(conn net.Conn, s *Server)

type Server struct {
	ServiceConnGroups map[int]map[int]*Service // 服务类型 -> 服务实例id -> 实例
	Config            *services.Config
	ClientsId         uint64
	Clients           map[uint64]*Session
	// ClientsMap        map[int]map[int]map[uint64]bool // 服务类型 -> 服务实例id -> clientId -> 是否存在
}

func (s *Server) InitServer(conf *services.Config) {
	s.ServiceConnGroups = make(map[int]map[int]*Service)
	s.Clients = make(map[uint64]*Session)
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
		go handleClient(conn, s)
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
		go handleClient(conn, s)
	}
}

func (s *Server) DialRoom(ip string, port int) (net.Conn, error) {
	retry, retryMax, retryDuration := 0, 5, time.Duration(2)

	log.Logger().Debugf("DialRoom ip:%s, id:%d", ip, port)
	for retry < retryMax {
		conn, err := kcp.Dial(fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			log.Logger().Warnf("dial kcp service ip:%s, port:%d, for %d time(s), reason:%s", ip, port, retry, err.Error())
			retry++
			time.Sleep(retryDuration)
		} else {
			return conn, nil
		}
	}

	return nil, errors.New("dial remote server err")
}

func (s *Server) AddService(ip string, port int, serviceName int, serviceId int) (net.Conn, error) {
	retry, retryMax, retryDuration := 0, 5, time.Duration(2)

	log.Logger().Debugf("adding service name:%s, id:%d", flats.EnumNamesPacketId[serviceName], serviceId)
	for retry < retryMax {
		conn, err := kcp.Dial(fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			log.Logger().Warnf("dial kcp service ip:%s, port:%d, for %d time(s), reason:%s", ip, port, retry, err.Error())
			retry++
			time.Sleep(retryDuration)
		} else {
			// log.Logger().Debug("connection:", conn.LocalAddr(), "->", conn.RemoteAddr())

			serviceGroup := s.ServiceConnGroups[serviceName]
			if serviceGroup == nil {
				serviceGroup = make(map[int]*Service)
			}

			service := InitService(conn, s, serviceName, serviceId)

			serviceGroup[serviceId] = service

			s.ServiceConnGroups[serviceName] = serviceGroup
			return conn, nil
		}
	}

	return nil, errors.New("dial remote server err")
}

func (s *Server) GetService(serviceName int, serviceId int) *Service {
	if groups := s.ServiceConnGroups[serviceName]; groups != nil {
		return groups[serviceId]
	}
	return nil
}
