package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/structs/users"

	"github.com/google/flatbuffers/go"
)

func RqUserLogin(sess *netWorking.Session, pack *netPackages.NetPackage) {
	rq := flats.GetRootAsRqLogin(pack.Data, 0)

	name := string(rq.Name())
	pw := string(rq.Password())

	u := users.FindUserByName(name)
	if u == nil {
		u = users.CreateUser(name, pw)
	}

	sess.UserId = u.ID

	builder := flatbuffers.NewBuilder(0)
	rpName := builder.CreateByteString(rq.Name())

	flats.RpLoginStart(builder)
	flats.RpLoginAddName(builder, rpName)
	rp := flats.RpLoginEnd(builder)
	builder.Finish(rp)

	sess.Write(flats.ReponseIdLogin, pack, builder.FinishedBytes())
}
