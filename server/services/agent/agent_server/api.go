package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
)

var ReqHandler = map[uint8]func(sess *Session, pack *netPackages.NetPackage, outBuffer *Buffer){
	flats.RequestIdLogin: RqUserLogin,
}
