package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"fmt"
)

func RqUserLogin(sess *Session, pack *netPackages.NetPackage, outBuffer *Buffer) {
	rq := flats.GetRootAsRqLogin(pack.Data, 0)

	name := string(rq.Name())
	pw := string(rq.Password())

	fmt.Println("name: ", name, " password ", pw)
}
