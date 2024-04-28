// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package thrylos

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type GetBalanceRequest struct {
	_tab flatbuffers.Table
}

func GetRootAsGetBalanceRequest(buf []byte, offset flatbuffers.UOffsetT) *GetBalanceRequest {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &GetBalanceRequest{}
	x.Init(buf, n+offset)
	return x
}

func FinishGetBalanceRequestBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsGetBalanceRequest(buf []byte, offset flatbuffers.UOffsetT) *GetBalanceRequest {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &GetBalanceRequest{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedGetBalanceRequestBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *GetBalanceRequest) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *GetBalanceRequest) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *GetBalanceRequest) Address() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func GetBalanceRequestStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func GetBalanceRequestAddAddress(builder *flatbuffers.Builder, address flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(address), 0)
}
func GetBalanceRequestEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
