package game_server

import (
	"Clans/server/flats"
	"Clans/server/netPackages"
	"Clans/server/netWorking"
	"time"

	"github.com/google/flatbuffers/go"
)

var (
	AllFrameList map[uint32]*Frame // frameId -> FrameData 当前所有的帧数据

	gatherFrameChan chan *netPackages.FramePackage
	dispatchTicker  *time.Ticker // 帧数据分发计时器
	dispatchDur     = time.Duration(time.Millisecond * 100)
	dispatchChan    chan []byte

	LogicFrameId uint32 // 逻辑帧id

	builder *flatbuffers.Builder
)

type Frame struct {
	Id              uint32
	PlayerOprations map[uint8][]byte // playerId in room -> operationlist 客户端根据playerId来进行操作数据的分发
}

func init() {
	AllFrameList = make(map[uint32]*Frame)
	gatherFrameChan = make(chan *netPackages.FramePackage, 1024)
	dispatchChan = make(chan []byte, 1024)
	dispatchTicker = time.NewTicker(dispatchDur)

	builder = flatbuffers.NewBuilder(0)

	go func() {
		for _ = range dispatchTicker.C {
			gatherFrameChan <- nil
			// 插入nil作为分隔标志位
			pushDataToClient()
			LogicFrameId++
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
					sess.OutBuffer.SendFrame(data)
				}
			}
		}
	}(server)
}

func pushDataToClient() {

	for f := range gatherFrameChan {
		if f == nil {
			frameData := AllFrameList[LogicFrameId]
			if frameData == nil {
				return
			}

			var operationsOffset flatbuffers.UOffsetT
			// flats.LogicFrameStartOperationsVector(builder, len(frameData.PlayerOprations))
			operationOffsetList := make([]flatbuffers.UOffsetT, len(frameData.PlayerOprations))
			i := 0
			for pid, _ := range frameData.PlayerOprations {
				opts := frameData.PlayerOprations[pid]
				datas := builder.CreateByteVector(opts)

				flats.OperationStart(builder)
				flats.OperationAddPid(builder, byte(pid))
				flats.OperationAddData(builder, datas)

				rpData := flats.OperationEnd(builder)
				operationOffsetList[i] = rpData
				i++
				// operationOffsetList = append(operationOffsetList, rpData)
				// builder.PrependUOffsetT(rpData)
			}

			flats.LogicFrameStartOperationsVector(builder, len(frameData.PlayerOprations))
			for i := 0; i < len(operationOffsetList); i++ {
				builder.PrependUOffsetT(operationOffsetList[i])
			}
			operationsOffset = builder.EndVector(len(frameData.PlayerOprations))

			flats.LogicFrameStart(builder)
			flats.LogicFrameAddFrameId(builder, LogicFrameId)
			flats.LogicFrameAddOperations(builder, operationsOffset)

			builder.Finish(flats.LogicFrameEnd(builder))

			// 分发此帧操作给所有客户端
			dispatchChan <- builder.FinishedBytes()

			builder.Reset()
			return
		}

		// 保存帧数据到全局帧数据列表
		frame := AllFrameList[LogicFrameId]
		if frame == nil {
			frame = new(Frame)
			frame.Id = LogicFrameId
			AllFrameList[LogicFrameId] = frame
		}

		operations := frame.PlayerOprations
		if operations == nil {
			operations = make(map[uint8][]byte)
		}

		operations[f.PlayerId] = append(operations[f.PlayerId], f.SrcDatas...)
		frame.PlayerOprations = operations

		// dispatchChan <- f.SrcDatas
	}
}
