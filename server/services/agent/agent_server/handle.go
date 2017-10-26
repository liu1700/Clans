package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/structs/users"

	"github.com/google/flatbuffers/go"
)

var pid int

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

	sess.Write(flats.ResponseIdLogin, pack, builder.FinishedBytes())
}

func RqJoinRoom(sess *netWorking.Session, pack *netPackages.NetPackage) {

	sess.JoinRoom("192.168.1.102", 9080)

	sess.Write(flats.ResponseIdJoinRoom, pack, []byte{})
}

func RqFetchPlayerSpawnData(sess *netWorking.Session, pack *netPackages.NetPackage) {
	pid++
	builder := flatbuffers.NewBuilder(0)

	flats.RpPlayerSpawnStart(builder)
	flats.RpPlayerSpawnAddPid(builder, byte(pid))
	flats.RpPlayerSpawnAddHealth(builder, byte(100))
	flats.RpPlayerSpawnAddShield(builder, byte(100))
	flats.RpPlayerSpawnAddSpawnAtX(builder, int16(31))
	flats.RpPlayerSpawnAddSpawnAtY(builder, int16(28))
	rp := flats.RpPlayerSpawnEnd(builder)
	builder.Finish(rp)

	sess.Write(flats.ResponseIdMySpawnData, pack, builder.FinishedBytes())
}
