package game

import (
	"Clans/server/db"
	"Clans/server/log"
	"Clans/server/netWorking"
	"Clans/server/services"
	"Clans/server/utils"
	"net"
	"os"
	"sync"
)

var (
	Wg               sync.WaitGroup
	shuttingDownChan = make(chan struct{})
	Version          int
)

func InitDBTables() {

}

func handleClient(conn net.Conn, s *netWorking.Server) {

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

	Wg.Wait()

	os.Exit(0)
	return
}
