package agent_server

import (
	"Clans/server/flats"
)

var ReqHandler = map[uint8]func(sess *Session, outBuffer *Buffer){
	flats.RequestIdLogin: RqUserLogin,
}
