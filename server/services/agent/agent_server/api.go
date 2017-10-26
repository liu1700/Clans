package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
)

var ReqHandler = map[uint8]func(sess *netWorking.Session, pack *netPackages.NetPackage){
	flats.RequestIdLogin:       RqUserLogin,
	flats.RequestIdJoinRoom:    RqJoinRoom,
	flats.RequestIdMySpawnData: RqFetchPlayerSpawnData,
}
