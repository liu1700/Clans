package agent_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/structs/users"

	"github.com/google/flatbuffers/go"
)

var pid int

var playerList map[uint32]bool
var roomReady bool

func init() {
	playerList = make(map[uint32]bool)
}

func RqUserLogin(sess *netWorking.Session, pack *netPackages.NetPackage) {
	rq := flats.GetRootAsRqLogin(pack.Data, 0)

	name := string(rq.Name())
	pw := string(rq.Password())

	u := users.FindUserByName(name)
	if u == nil {
		u = users.CreateUser(name, pw)
	}

	sess.UserId = u.ID
	sess.ServerInst.UserClients[sess.UserId] = sess

	builder := flatbuffers.NewBuilder(0)
	rpName := builder.CreateByteString(rq.Name())

	flats.RpLoginStart(builder)
	flats.RpLoginAddName(builder, rpName)
	rp := flats.RpLoginEnd(builder)
	builder.Finish(rp)

	sess.Write(flats.ResponseIdLogin, pack, builder.FinishedBytes())
}

func RqJoinRoom(sess *netWorking.Session, pack *netPackages.NetPackage) {

	// sess.JoinRoom("192.168.1.102", 9080)

	playerList[sess.UserId] = true

	builder := flatbuffers.NewBuilder(0)

	flats.RpMatchMakingStart(builder)
	flats.RpMatchMakingAddUserId(builder, sess.UserId)
	flats.RpMatchMakingAddIsJoin(builder, int8(1))
	rp := flats.RpMatchMakingEnd(builder)
	builder.Finish(rp)
	joinMatchMakingData := builder.FinishedBytes()
	cloneJoinMatchMaking := make([]byte, len(joinMatchMakingData))
	copy(cloneJoinMatchMaking, joinMatchMakingData)
	builder.Reset()

	rpIp := builder.CreateString("192.168.1.102")
	flats.RpStartMatchMakingStart(builder)
	flats.RpStartMatchMakingAddBattleServerIp(builder, rpIp)
	flats.RpStartMatchMakingAddServerPort(builder, uint16(9080))
	rpServer := flats.RpStartMatchMakingEnd(builder)
	builder.Finish(rpServer)
	sess.Write(flats.ResponseIdStartMatchMaking, pack, builder.FinishedBytes())
	builder.Reset()

	for uId, _ := range playerList {
		if s := sess.ServerInst.UserClients[uId]; s != nil {
			sess.Write(flats.ResponseIdMatchMaking, pack, cloneJoinMatchMaking)

			// 人满，开打
			if len(playerList) == 2 {
				roomReady = true
				// 不应该复用pack
				s.Write(flats.ResponseIdJoinRoom, pack, []byte{})
			}
		}
	}
}

func RqFetchSpawnData(sess *netWorking.Session, pack *netPackages.NetPackage) {
	// 本局pid
	pid++
	builder := flatbuffers.NewBuilder(0)

	flats.RpPlayerSpawnStart(builder)
	flats.RpPlayerSpawnAddPid(builder, byte(pid))
	flats.RpPlayerSpawnAddHealth(builder, byte(100))
	flats.RpPlayerSpawnAddShield(builder, byte(100))
	flats.RpPlayerSpawnAddSpawnAtX(builder, int16(40))
	flats.RpPlayerSpawnAddSpawnAtY(builder, int16(28))
	rp := flats.RpPlayerSpawnEnd(builder)
	builder.Finish(rp)
	sess.Write(flats.ResponseIdMySpawnData, pack, builder.FinishedBytes())
}
