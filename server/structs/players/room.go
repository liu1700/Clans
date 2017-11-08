package players

const (
	MaxPlayerCount = 2
)

type Room struct {
	RoomId      uint32
	Start       bool
	PlayerCount int
	PlayerList  map[uint32]*PlayerSpawn
}
