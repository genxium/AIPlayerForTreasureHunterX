package models

import (
	"github.com/golang/protobuf/proto"
)

type RoomDownsyncFrame struct {
	Id             int32                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	RefFrameId     int32                 `protobuf:"varint,2,opt,name=refFrameId,proto3" json:"refFrameId,omitempty"`
	Players        map[int32]*Player     `protobuf:"bytes,3,rep,name=players,proto3" json:"players,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SentAt         int64                 `protobuf:"varint,4,opt,name=sentAt,proto3" json:"sentAt,omitempty"`
	CountdownNanos int64                 `protobuf:"varint,5,opt,name=countdownNanos,proto3" json:"countdownNanos,omitempty"`
	Treasures      map[int32]*Treasure   `protobuf:"bytes,6,rep,name=treasures,proto3" json:"treasures,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Traps          map[int32]*Trap       `protobuf:"bytes,7,rep,name=traps,proto3" json:"traps,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Bullets        map[int32]*Bullet     `protobuf:"bytes,8,rep,name=bullets,proto3" json:"bullets,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SpeedShoes     map[int32]*SpeedShoe  `protobuf:"bytes,9,rep,name=speedShoes,proto3" json:"speedShoes,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Pumpkins       map[int32]*Pumpkin    `protobuf:"bytes,10,rep,name=pumpkin,proto3" json:"pumpkin,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	GuardTowers    map[int32]*GuardTower `protobuf:"bytes,11,rep,name=guardTowers,proto3" json:"guardTowers,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *RoomDownsyncFrame) Reset()         { *m = RoomDownsyncFrame{} }
func (m *RoomDownsyncFrame) String() string { return proto.CompactTextString(m) }
func (m *RoomDownsyncFrame) ProtoMessage()  {}
