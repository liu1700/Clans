package game_server

import (
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"time"
)

var (
	AllFrameList map[uint32]*Frame // frameId -> FrameData 当前所有的帧数据

	gatherFrameChan chan *netPackages.FramePackage
	dispatchTicker  *time.Ticker // 帧数据分发计时器
	dispatchDur     = time.Duration(time.Millisecond * 50)
	dispatchChan    chan []byte
)

type Frame struct {
	Id              uint32
	PlayerOprations map[byte][]byte // playerId in room -> operationlist 客户端根据playerId来进行操作数据的分发
}

func init() {
	AllFrameList = make(map[uint32]*Frame)
	gatherFrameChan = make(chan *netPackages.FramePackage, 1024)
	dispatchChan = make(chan []byte, 1024)
	dispatchTicker = time.NewTicker(dispatchDur)

	go func() {
		for _ = range dispatchTicker.C {
			gatherFrameChan <- nil
			// 插入nil作为分隔标志位
			pushDataToClient()
		}
	}()
}

func InitDispatcher(server *netWorking.Server) {
	go func(s *netWorking.Server) {
		for {
			select {
			case data, ok := <-dispatchChan:
				if !ok {
					continue
				}
				for id, _ := range server.Clients {
					sess := server.Clients[id]
					if sess == nil {
						continue
					}
					sess.OutBuffer.RawSend(data)
				}
			}
		}
	}(server)
}

func pushDataToClient() {
	for f := range gatherFrameChan {
		if f == nil {
			return
		}

		frame := AllFrameList[f.FrameId]
		if frame == nil {
			frame = new(Frame)
			frame.Id = f.FrameId
			AllFrameList[f.FrameId] = frame
		}

		operations := frame.PlayerOprations
		if operations == nil {
			operations = make(map[byte][]byte)
		}

		operations[f.PlayerId] = append(operations[f.PlayerId], f.SrcDatas...)
		frame.PlayerOprations = operations

		dispatchChan <- f.SrcDatas
	}
}
