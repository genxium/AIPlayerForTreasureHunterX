package models

import (
	"github.com/golang/protobuf/proto"
)

type Player struct {
	Id                int64      `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	X                 float64    `protobuf:"fixed64,2,opt,name=x,proto3" json:"x,omitempty"`
	Y                 float64    `protobuf:"fixed64,3,opt,name=y,proto3" json:"y,omitempty"`
	Dir               *Direction `protobuf:"bytes,4,opt,name=dir,proto3" json:"dir,omitempty"`
	Speed             int32      `protobuf:"varint,5,opt,name=speed,proto3" json:"speed,omitempty"`
	BattleState       int32      `protobuf:"varint,6,opt,name=battleState,proto3" json:"battleState,omitempty"`
	LastMoveGmtMillis int32      `protobuf:"varint,7,opt,name=lastMoveGmtMillis,proto3" json:"lastMoveGmtMillis,omitempty"`
	Name              string     `protobuf:"bytes,8,opt,name=name,proto3" json:"name,omitempty"`
	DisplayName       string     `protobuf:"bytes,9,opt,name=displayName,proto3" json:"displayName,omitempty"`
	Score             int32      `protobuf:"varint,10,opt,name=score,proto3" json:"score,omitempty"`
	Removed           bool       `protobuf:"varint,11,opt,name=removed,proto3" json:"removed,omitempty"`
	JoinIndex         int32      `protobuf:"varint,12,opt,name=joinIndex,proto3" json:"joinIndex,omitempty"`
	         int32      `protobuf:"varint,12,opt,name=joinIndex,proto3" json:"joinIndex,omitempty"`
}

func (m *Player) Reset()         { *m = Player{} }
func (m *Player) String() string { return proto.CompactTextString(m) }
func (*Player) ProtoMessage()    {}
