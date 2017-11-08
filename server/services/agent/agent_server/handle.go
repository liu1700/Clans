package agent_server

import (
	"Clans/server/flats"
	"Clans/server/log"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"Clans/server/structs/players"
	"Clans/server/structs/users"
	"Clans/server/utils"
	"fmt"

	"github.com/google/flatbuffers/go"
)

var roomReady bool

var availableRoomPool map[uint32]*players.Room
var gamingRoomPool map[uint32]*players.Room

var positions = [][]int{
	[]int{40, 28}, // x,y
	[]int{41, 27},
	[]int{40, 28},
	[]int{41, 29},
	[]int{41, 30},
	[]int{40, 28},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
	[]int{40, 29},
}

var availableBattleServer = []string{
	"192.168.1.106",
	"123.206.66.47",
}

func init() {
	availableRoomPool = make(map[uint32]*players.Room)
	gamingRoomPool = make(map[uint32]*players.Room)
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

func RqBroadCastRoomIsReady(sess *netWorking.Session, pack *netPackages.NetPackage) {
	if room := gamingRoomPool[sess.RoomId]; room != nil {
		for uId, _ := range room.PlayerList {
			if s := sess.ServerInst.UserClients[uId]; s != nil {
				s.Write(flats.ResponseIdRoomIsReady, pack, []byte{})
			}
		}
	}
}

func RqJoinRoom(sess *netWorking.Session, pack *netPackages.NetPackage) {

	// sess.JoinRoom("192.168.1.102", 9080)
	roomId := uint32(0)
	pid := -1
	room := new(players.Room)
	if len(availableRoomPool) > 0 {
		for id, _ := range availableRoomPool {
			room = availableRoomPool[id]
			if room != nil && !room.Start && room.PlayerCount < players.MaxPlayerCount {
				roomId = room.RoomId
				room.PlayerCount++
				if room.PlayerCount == players.MaxPlayerCount {
					room.Start = true
				}
				playerlist := room.PlayerList
				pid = room.PlayerCount
				if playerlist == nil {
					playerlist = make(map[uint32]*players.PlayerSpawn)
				}
				playerlist[sess.UserId] = &players.PlayerSpawn{
					X:               positions[pid][0],
					Y:               positions[pid][1],
					UserId:          sess.UserId,
					PlayerIdInRound: pid,
				}

				if room.Start {
					gamingRoomPool[id] = room
					delete(availableRoomPool, id)
				} else {
					availableRoomPool[id] = room
				}
				break
			}
		}
	} else {
		roomId = <-utils.LCG
		pid = 1
		player := &players.PlayerSpawn{
			X:               positions[pid][0],
			Y:               positions[pid][1],
			UserId:          sess.UserId,
			PlayerIdInRound: pid,
		}
		playerlist := map[uint32]*players.PlayerSpawn{
			sess.UserId: player,
		}
		room = &players.Room{
			RoomId:      roomId,
			PlayerList:  playerlist,
			PlayerCount: pid,
		}

		availableRoomPool[roomId] = room
	}

	if pid == -1 {
		log.Logger().Error("invalid pid")

		return
	}
	fmt.Println("roomId ", roomId)
	// dataMap := map[string]interface{}{
	// 	"RoomId":         randNum,
	// 	"MaxPlayerCount": 10,
	// }
	// netWorking.HttpPost(availableBattleServer[0], battleServerHttpPort, netWorking.CreateRoom, dataMap, CreateRoomCallBack)

	sess.RoomId = roomId

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

	rpIp := builder.CreateString(availableBattleServer[1])
	flats.RpStartMatchMakingStart(builder)
	flats.RpStartMatchMakingAddBattleServerIp(builder, rpIp)
	flats.RpStartMatchMakingAddServerPort(builder, uint16(5055))
	rpServer := flats.RpStartMatchMakingEnd(builder)
	builder.Finish(rpServer)
	sess.Write(flats.ResponseIdStartMatchMaking, pack, builder.FinishedBytes())
	builder.Reset()

	if room != nil {
		counter := 0
		for uId, _ := range room.PlayerList {
			if s := sess.ServerInst.UserClients[uId]; s != nil {
				s.Write(flats.ResponseIdMatchMaking, pack, cloneJoinMatchMaking)
				// 人满，开打
				if room.Start {
					flats.RpJoinRoomStart(builder)
					chosenOne := int8(0)
					if counter == 0 {
						chosenOne = int8(1)
					}
					flats.RpJoinRoomAddChosenOne(builder, chosenOne)
					flats.RpJoinRoomAddRoomId(builder, room.RoomId)
					flats.RpJoinRoomAddMaxPlayerCountInRoom(builder, byte(players.MaxPlayerCount))
					rpRoom := flats.RpJoinRoomEnd(builder)
					builder.Finish(rpRoom)
					// 不应该复用pack
					s.Write(flats.ResponseIdJoinRoom, pack, builder.FinishedBytes())
					builder.Reset()
					counter++
				}
			}
		}
	}
}

func RqFetchSpawnData(sess *netWorking.Session, pack *netPackages.NetPackage) {

	builder := flatbuffers.NewBuilder(0)

	var playersOffset flatbuffers.UOffsetT
	lenght := 1
	playersOffsetList := make([]flatbuffers.UOffsetT, lenght)
	i := 0

	myRoom := gamingRoomPool[sess.RoomId]
	if myRoom == nil {
		log.Logger().Error("room nil")
		return
	}

	mySqawnInfo := myRoom.PlayerList[sess.UserId]
	if mySqawnInfo == nil {
		log.Logger().Error("spawn info nil")
		return
	}

	flats.RpPlayerSpawnStart(builder)
	flats.RpPlayerSpawnAddPid(builder, byte(mySqawnInfo.PlayerIdInRound))
	flats.RpPlayerSpawnAddHealth(builder, byte(100))
	flats.RpPlayerSpawnAddShield(builder, byte(0))
	flats.RpPlayerSpawnAddSpawnAtX(builder, int16(mySqawnInfo.X))
	flats.RpPlayerSpawnAddSpawnAtY(builder, int16(mySqawnInfo.Y))

	rp := flats.RpPlayerSpawnEnd(builder)

	playersOffsetList[i] = rp

	playerId := mySqawnInfo.PlayerIdInRound

	flats.RpAllPlayerSpawnsStartOthersVector(builder, lenght)
	builder.PrependUOffsetT(playersOffsetList[i])
	playersOffset = builder.EndVector(lenght)

	// for uId, _ := range playerList {
	// 	l := playerList[uId]
	// 	//
	// 	flats.RpPlayerSpawnStart(builder)
	// 	flats.RpPlayerSpawnAddPid(builder, byte(l.PlayerIdInRound))
	// 	flats.RpPlayerSpawnAddHealth(builder, byte(100))
	// 	flats.RpPlayerSpawnAddShield(builder, byte(0))
	// 	flats.RpPlayerSpawnAddSpawnAtX(builder, int16(l.X))
	// 	flats.RpPlayerSpawnAddSpawnAtY(builder, int16(l.Y))

	// 	rp := flats.RpPlayerSpawnEnd(builder)

	// 	playersOffsetList[i] = rp
	// 	i++

	// 	if uId == sess.UserId {
	// 		playerId = l.PlayerIdInRound
	// 	}
	// }

	// flats.RpAllPlayerSpawnsStartOthersVector(builder, len(playerList))
	// for i := 0; i < len(playersOffsetList); i++ {
	// 	builder.PrependUOffsetT(playersOffsetList[i])
	// }
	// playersOffset = builder.EndVector(len(playerList))

	// 广播位置
	flats.RpAllPlayerSpawnsStart(builder)
	flats.RpAllPlayerSpawnsAddPid(builder, byte(playerId))
	flats.RpAllPlayerSpawnsAddOthers(builder, playersOffset)
	builder.Finish(flats.RpAllPlayerSpawnsEnd(builder))

	sess.Write(flats.ResponseIdMySpawnData, pack, builder.FinishedBytes())

	// // 广播自己位置
	// flats.RpPlayerSpawnStart(builder)
	// flats.RpPlayerSpawnAddPid(builder, byte(pid))
	// flats.RpPlayerSpawnAddHealth(builder, byte(100))
	// flats.RpPlayerSpawnAddShield(builder, byte(100))
	// flats.RpPlayerSpawnAddSpawnAtX(builder, int16(40))
	// flats.RpPlayerSpawnAddSpawnAtY(builder, int16(28))
	// rp := flats.RpPlayerSpawnEnd(builder)
	// builder.Finish(rp)
	// sess.Write(flats.ResponseIdMySpawnData, pack, builder.FinishedBytes())

	// // 广播其他玩家的初始位置

}
