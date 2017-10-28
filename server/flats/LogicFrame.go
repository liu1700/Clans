// automatically generated by the FlatBuffers compiler, do not modify

package flats

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type LogicFrame struct {
	_tab flatbuffers.Table
}

func GetRootAsLogicFrame(buf []byte, offset flatbuffers.UOffsetT) *LogicFrame {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &LogicFrame{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *LogicFrame) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *LogicFrame) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *LogicFrame) FrameId() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *LogicFrame) MutateFrameId(n uint32) bool {
	return rcv._tab.MutateUint32Slot(4, n)
}

func (rcv *LogicFrame) Operations(j int) byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetByte(a + flatbuffers.UOffsetT(j*1))
	}
	return 0
}

func (rcv *LogicFrame) OperationsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *LogicFrame) OperationsBytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func LogicFrameStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func LogicFrameAddFrameId(builder *flatbuffers.Builder, frameId uint32) {
	builder.PrependUint32Slot(0, frameId, 0)
}
func LogicFrameAddOperations(builder *flatbuffers.Builder, operations flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(operations), 0)
}
func LogicFrameStartOperationsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(1, numElems, 1)
}
func LogicFrameEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
