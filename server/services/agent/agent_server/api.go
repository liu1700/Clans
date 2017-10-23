package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
)

var ReqHandler = map[uint8]func(sess *netWorking.Session, pack *netPackages.NetPackage, outBuffer *Buffer){
	flats.RequestIdLogin: RqUserLogin,
}
