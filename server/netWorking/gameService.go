package netWorking

import (
	"Clans/server/flats"
	"Clans/server/log"
	"Clans/server/netPackages"
	"net"
	"time"
)

type Service struct {
	Connection      net.Conn
	ServiceName     int
	ServiceId       int
	ServerInstance  *Server
	heartBeatTicker *time.Ticker
	// ClientsMap      map[uint64]bool // clients -> exists
}

func InitService(conn net.Conn, s *Server, name int, id int) *Service {
	service := &Service{
		Connection:     conn,
		ServerInstance: s,
		ServiceName:    name,
		ServiceId:      id,
		// ClientsMap
	}

	service.heartBeatTicker = time.NewTicker(time.Second * 5)

	go func() {
		for _ = range service.heartBeatTicker.C {
			service.Ping()
		}
	}()

	go service.ReadLoop()
	return service
}

func (service *Service) Ping() {
	service.Forward(netPackages.HeartBeatPacket())
}

func (service *Service) ReadLoop() {
	// read loop
	readBytes := make([]byte, netPackages.PACKET_LIMIT)
	for {
		service.Connection.SetReadDeadline(time.Time{})
		// alloc a byte slice for reading data
		_, err := service.Connection.Read(readBytes)
		if err != nil {
			log.Logger().Errorf("read readBytes failed, reason:%v , service name:%s", err.Error(), flats.EnumNamesPacketId[service.ServiceName])
			continue
		}

		// 转化为package
		payload, err := netPackages.BytesToServicePackage(readBytes)
		if err != nil {
			log.Logger().Errorf("read payload faild, err :%v", err.Error())
			continue
		}

		log.Logger().Debugf("services %+v \n", *payload)

		// deliver the data to the input queue of agent()
		// service.ServerInstance.
	}
}

func (service *Service) Forward(data []byte) {
	_, err := service.Connection.Write(data)
	if err != nil {
		log.Logger().Errorf("error when forward data to name %d, id %d, err:%s", service.ServiceName, service.ServiceId, err.Error())
	}
}
