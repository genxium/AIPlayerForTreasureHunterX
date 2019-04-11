package models

type Player struct {
	Id                  int32         `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty" db:"id"`
	X                   float64       `protobuf:"fixed64,2,opt,name=x,proto3" json:"x,omitempty"`
	Y                   float64       `protobuf:"fixed64,3,opt,name=y,proto3" json:"y,omitempty"`
	Dir                 *Direction    `protobuf:"bytes,4,opt,name=dir,proto3" json:"dir,omitempty"`
	Speed               int32         `protobuf:"varint,5,opt,name=speed,proto3" json:"speed,omitempty"`
	BattleState         int32         `protobuf:"varint,6,opt,name=battleState,proto3" json:"battleState,omitempty"`
	LastMoveGmtMillis   int32         `protobuf:"varint,7,opt,name=lastMoveGmtMillis,proto3" json:"lastMoveGmtMillis,omitempty"`
	Name                string        `protobuf:"bytes,8,opt,name=name,proto3" json:"name,omitempty" db:"name"`
	DisplayName         string        `protobuf:"bytes,9,opt,name=displayName,proto3" json:"displayName,omitempty" db:"display_name"`
	Score               int32         `protobuf:"varint,10,opt,name=score,proto3" json:"score,omitempty"`
	JoinIndex           int32         `protobuf:"varint,12,opt,name=joinIndex,proto3 " json:"joinIndex"`
	Avatar              string        `protobuf:"varint,13,opt,name=avatar,proto3" json:"avatar,omitempty"`
}
