// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.18.1
// source: ultron.proto

package genproto

import (
	statistics "github.com/wosai/ultron/v2/pkg/statistics"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EventType int32

const (
	EventType_UNKONWN            EventType = 0
	EventType_PING               EventType = 1 // 心跳包
	EventType_CONNECTED          EventType = 2 //已连接
	EventType_DISCONNECT         EventType = 3 // 要求slave断开连接
	EventType_PLAN_STARTED       EventType = 4 // 测试计划开始
	EventType_PLAN_FINISHED      EventType = 5 // 测试计划结束
	EventType_PLAN_INTERRUPTED   EventType = 6 // 测试计划中断执行
	EventType_NEXT_STAGE_STARTED EventType = 7 // 开始执行计划中的下一阶段
	EventType_STATS_AGGREGATE    EventType = 8 // 上报统计对象
)

// Enum value maps for EventType.
var (
	EventType_name = map[int32]string{
		0: "UNKONWN",
		1: "PING",
		2: "CONNECTED",
		3: "DISCONNECT",
		4: "PLAN_STARTED",
		5: "PLAN_FINISHED",
		6: "PLAN_INTERRUPTED",
		7: "NEXT_STAGE_STARTED",
		8: "STATS_AGGREGATE",
	}
	EventType_value = map[string]int32{
		"UNKONWN":            0,
		"PING":               1,
		"CONNECTED":          2,
		"DISCONNECT":         3,
		"PLAN_STARTED":       4,
		"PLAN_FINISHED":      5,
		"PLAN_INTERRUPTED":   6,
		"NEXT_STAGE_STARTED": 7,
		"STATS_AGGREGATE":    8,
	}
)

func (x EventType) Enum() *EventType {
	p := new(EventType)
	*p = x
	return p
}

func (x EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_ultron_proto_enumTypes[0].Descriptor()
}

func (EventType) Type() protoreflect.EnumType {
	return &file_ultron_proto_enumTypes[0]
}

func (x EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EventType.Descriptor instead.
func (EventType) EnumDescriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{0}
}

type ResponseSubmit_Result int32

const (
	ResponseSubmit_UNKNOWN            ResponseSubmit_Result = 0
	ResponseSubmit_ACCEPTED           ResponseSubmit_Result = 1 // 上报对象已经接收
	ResponseSubmit_UNREGISTERED_SLAVE ResponseSubmit_Result = 2 // 未登记的slave id
	ResponseSubmit_BATCH_REJECTED     ResponseSubmit_Result = 3 //  批次被拒绝
	ResponseSubmit_BAD_SUBMISSION     ResponseSubmit_Result = 4 // 错误的提交，一般为内容错误
)

// Enum value maps for ResponseSubmit_Result.
var (
	ResponseSubmit_Result_name = map[int32]string{
		0: "UNKNOWN",
		1: "ACCEPTED",
		2: "UNREGISTERED_SLAVE",
		3: "BATCH_REJECTED",
		4: "BAD_SUBMISSION",
	}
	ResponseSubmit_Result_value = map[string]int32{
		"UNKNOWN":            0,
		"ACCEPTED":           1,
		"UNREGISTERED_SLAVE": 2,
		"BATCH_REJECTED":     3,
		"BAD_SUBMISSION":     4,
	}
)

func (x ResponseSubmit_Result) Enum() *ResponseSubmit_Result {
	p := new(ResponseSubmit_Result)
	*p = x
	return p
}

func (x ResponseSubmit_Result) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ResponseSubmit_Result) Descriptor() protoreflect.EnumDescriptor {
	return file_ultron_proto_enumTypes[1].Descriptor()
}

func (ResponseSubmit_Result) Type() protoreflect.EnumType {
	return &file_ultron_proto_enumTypes[1]
}

func (x ResponseSubmit_Result) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ResponseSubmit_Result.Descriptor instead.
func (ResponseSubmit_Result) EnumDescriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{5, 0}
}

type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SlaveId string            `protobuf:"bytes,1,opt,name=slave_id,json=slaveId,proto3" json:"slave_id,omitempty"`
	Extras  map[string]string `protobuf:"bytes,2,rep,name=extras,proto3" json:"extras,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ultron_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_ultron_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{0}
}

func (x *Session) GetSlaveId() string {
	if x != nil {
		return x.SlaveId
	}
	return ""
}

func (x *Session) GetExtras() map[string]string {
	if x != nil {
		return x.Extras
	}
	return nil
}

type TimerDTO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type  string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Timer []byte `protobuf:"bytes,2,opt,name=timer,proto3" json:"timer,omitempty"`
}

func (x *TimerDTO) Reset() {
	*x = TimerDTO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ultron_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimerDTO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimerDTO) ProtoMessage() {}

func (x *TimerDTO) ProtoReflect() protoreflect.Message {
	mi := &file_ultron_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimerDTO.ProtoReflect.Descriptor instead.
func (*TimerDTO) Descriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{1}
}

func (x *TimerDTO) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *TimerDTO) GetTimer() []byte {
	if x != nil {
		return x.Timer
	}
	return nil
}

type AttackStrategyDTO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type           string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	AttackStrategy []byte `protobuf:"bytes,2,opt,name=attack_strategy,json=attackStrategy,proto3" json:"attack_strategy,omitempty"`
}

func (x *AttackStrategyDTO) Reset() {
	*x = AttackStrategyDTO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ultron_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AttackStrategyDTO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AttackStrategyDTO) ProtoMessage() {}

func (x *AttackStrategyDTO) ProtoReflect() protoreflect.Message {
	mi := &file_ultron_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AttackStrategyDTO.ProtoReflect.Descriptor instead.
func (*AttackStrategyDTO) Descriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{2}
}

func (x *AttackStrategyDTO) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *AttackStrategyDTO) GetAttackStrategy() []byte {
	if x != nil {
		return x.AttackStrategy
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type EventType `protobuf:"varint,1,opt,name=type,proto3,enum=wosai.ultron.EventType" json:"type,omitempty"`
	// Types that are assignable to Data:
	//	*Event_PlanName
	//	*Event_AttackStrategy
	//	*Event_Timer
	//	*Event_BatchId
	Data isEvent_Data `protobuf_oneof:"data"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ultron_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_ultron_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{3}
}

func (x *Event) GetType() EventType {
	if x != nil {
		return x.Type
	}
	return EventType_UNKONWN
}

func (m *Event) GetData() isEvent_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *Event) GetPlanName() string {
	if x, ok := x.GetData().(*Event_PlanName); ok {
		return x.PlanName
	}
	return ""
}

func (x *Event) GetAttackStrategy() *AttackStrategyDTO {
	if x, ok := x.GetData().(*Event_AttackStrategy); ok {
		return x.AttackStrategy
	}
	return nil
}

func (x *Event) GetTimer() *TimerDTO {
	if x, ok := x.GetData().(*Event_Timer); ok {
		return x.Timer
	}
	return nil
}

func (x *Event) GetBatchId() uint32 {
	if x, ok := x.GetData().(*Event_BatchId); ok {
		return x.BatchId
	}
	return 0
}

type isEvent_Data interface {
	isEvent_Data()
}

type Event_PlanName struct {
	PlanName string `protobuf:"bytes,2,opt,name=plan_name,json=planName,proto3,oneof"`
}

type Event_AttackStrategy struct {
	AttackStrategy *AttackStrategyDTO `protobuf:"bytes,3,opt,name=attack_strategy,json=attackStrategy,proto3,oneof"`
}

type Event_Timer struct {
	Timer *TimerDTO `protobuf:"bytes,4,opt,name=timer,proto3,oneof"`
}

type Event_BatchId struct {
	BatchId uint32 `protobuf:"varint,5,opt,name=batch_id,json=batchId,proto3,oneof"`
}

func (*Event_PlanName) isEvent_Data() {}

func (*Event_AttackStrategy) isEvent_Data() {}

func (*Event_Timer) isEvent_Data() {}

func (*Event_BatchId) isEvent_Data() {}

type RequestSubmit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SlaveId string                           `protobuf:"bytes,1,opt,name=slave_id,json=slaveId,proto3" json:"slave_id,omitempty"`
	BatchId uint32                           `protobuf:"varint,2,opt,name=batch_id,json=batchId,proto3" json:"batch_id,omitempty"`
	Stats   *statistics.StatisticianGroupDTO `protobuf:"bytes,3,opt,name=stats,proto3" json:"stats,omitempty"`
}

func (x *RequestSubmit) Reset() {
	*x = RequestSubmit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ultron_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestSubmit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestSubmit) ProtoMessage() {}

func (x *RequestSubmit) ProtoReflect() protoreflect.Message {
	mi := &file_ultron_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestSubmit.ProtoReflect.Descriptor instead.
func (*RequestSubmit) Descriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{4}
}

func (x *RequestSubmit) GetSlaveId() string {
	if x != nil {
		return x.SlaveId
	}
	return ""
}

func (x *RequestSubmit) GetBatchId() uint32 {
	if x != nil {
		return x.BatchId
	}
	return 0
}

func (x *RequestSubmit) GetStats() *statistics.StatisticianGroupDTO {
	if x != nil {
		return x.Stats
	}
	return nil
}

type ResponseSubmit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result ResponseSubmit_Result `protobuf:"varint,1,opt,name=result,proto3,enum=wosai.ultron.ResponseSubmit_Result" json:"result,omitempty"`
}

func (x *ResponseSubmit) Reset() {
	*x = ResponseSubmit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ultron_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResponseSubmit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResponseSubmit) ProtoMessage() {}

func (x *ResponseSubmit) ProtoReflect() protoreflect.Message {
	mi := &file_ultron_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResponseSubmit.ProtoReflect.Descriptor instead.
func (*ResponseSubmit) Descriptor() ([]byte, []int) {
	return file_ultron_proto_rawDescGZIP(), []int{5}
}

func (x *ResponseSubmit) GetResult() ResponseSubmit_Result {
	if x != nil {
		return x.Result
	}
	return ResponseSubmit_UNKNOWN
}

var File_ultron_proto protoreflect.FileDescriptor

var file_ultron_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c,
	0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x1a, 0x10, 0x73, 0x74,
	0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9a,
	0x01, 0x0a, 0x07, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x6c,
	0x61, 0x76, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x6c,
	0x61, 0x76, 0x65, 0x49, 0x64, 0x12, 0x39, 0x0a, 0x06, 0x65, 0x78, 0x74, 0x72, 0x61, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c,
	0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x45, 0x78, 0x74,
	0x72, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x65, 0x78, 0x74, 0x72, 0x61, 0x73,
	0x1a, 0x39, 0x0a, 0x0b, 0x45, 0x78, 0x74, 0x72, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x34, 0x0a, 0x08, 0x54,
	0x69, 0x6d, 0x65, 0x72, 0x44, 0x54, 0x4f, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74,
	0x69, 0x6d, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x74, 0x69, 0x6d, 0x65,
	0x72, 0x22, 0x50, 0x0a, 0x11, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x72, 0x61, 0x74,
	0x65, 0x67, 0x79, 0x44, 0x54, 0x4f, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x6b, 0x5f, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0e, 0x61, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x72, 0x61, 0x74,
	0x65, 0x67, 0x79, 0x22, 0xf4, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x2b, 0x0a,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x77, 0x6f,
	0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x0a, 0x09, 0x70, 0x6c,
	0x61, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x08, 0x70, 0x6c, 0x61, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x4a, 0x0a, 0x0f, 0x61, 0x74, 0x74,
	0x61, 0x63, 0x6b, 0x5f, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f,
	0x6e, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79,
	0x44, 0x54, 0x4f, 0x48, 0x00, 0x52, 0x0e, 0x61, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x72,
	0x61, 0x74, 0x65, 0x67, 0x79, 0x12, 0x2e, 0x0a, 0x05, 0x74, 0x69, 0x6d, 0x65, 0x72, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74,
	0x72, 0x6f, 0x6e, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x72, 0x44, 0x54, 0x4f, 0x48, 0x00, 0x52, 0x05,
	0x74, 0x69, 0x6d, 0x65, 0x72, 0x12, 0x1b, 0x0a, 0x08, 0x62, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x69,
	0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x00, 0x52, 0x07, 0x62, 0x61, 0x74, 0x63, 0x68,
	0x49, 0x64, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x7f, 0x0a, 0x0d, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x73,
	0x6c, 0x61, 0x76, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73,
	0x6c, 0x61, 0x76, 0x65, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x62, 0x61, 0x74, 0x63, 0x68, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x62, 0x61, 0x74, 0x63, 0x68, 0x49,
	0x64, 0x12, 0x38, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x22, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x69, 0x61, 0x6e, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x44, 0x54, 0x4f, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x22, 0xb2, 0x01, 0x0a, 0x0e,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x12, 0x3b,
	0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x23,
	0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x2e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x63, 0x0a, 0x06, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e,
	0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54, 0x45, 0x44, 0x10, 0x01,
	0x12, 0x16, 0x0a, 0x12, 0x55, 0x4e, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54, 0x45, 0x52, 0x45, 0x44,
	0x5f, 0x53, 0x4c, 0x41, 0x56, 0x45, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e, 0x42, 0x41, 0x54, 0x43,
	0x48, 0x5f, 0x52, 0x45, 0x4a, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x03, 0x12, 0x12, 0x0a, 0x0e,
	0x42, 0x41, 0x44, 0x5f, 0x53, 0x55, 0x42, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x10, 0x04,
	0x2a, 0xa9, 0x01, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b,
	0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4f, 0x4e, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x50,
	0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54,
	0x45, 0x44, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e, 0x4e, 0x45,
	0x43, 0x54, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x50, 0x4c, 0x41, 0x4e, 0x5f, 0x53, 0x54, 0x41,
	0x52, 0x54, 0x45, 0x44, 0x10, 0x04, 0x12, 0x11, 0x0a, 0x0d, 0x50, 0x4c, 0x41, 0x4e, 0x5f, 0x46,
	0x49, 0x4e, 0x49, 0x53, 0x48, 0x45, 0x44, 0x10, 0x05, 0x12, 0x14, 0x0a, 0x10, 0x50, 0x4c, 0x41,
	0x4e, 0x5f, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x52, 0x55, 0x50, 0x54, 0x45, 0x44, 0x10, 0x06, 0x12,
	0x16, 0x0a, 0x12, 0x4e, 0x45, 0x58, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x47, 0x45, 0x5f, 0x53, 0x54,
	0x41, 0x52, 0x54, 0x45, 0x44, 0x10, 0x07, 0x12, 0x13, 0x0a, 0x0f, 0x53, 0x54, 0x41, 0x54, 0x53,
	0x5f, 0x41, 0x47, 0x47, 0x52, 0x45, 0x47, 0x41, 0x54, 0x45, 0x10, 0x08, 0x32, 0x93, 0x01, 0x0a,
	0x0d, 0x55, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3b,
	0x0a, 0x09, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x12, 0x15, 0x2e, 0x77, 0x6f,
	0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x1a, 0x13, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f,
	0x6e, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x00, 0x30, 0x01, 0x12, 0x45, 0x0a, 0x06, 0x53,
	0x75, 0x62, 0x6d, 0x69, 0x74, 0x12, 0x1b, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c,
	0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x74, 0x1a, 0x1c, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f,
	0x6e, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74,
	0x22, 0x00, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2f, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2f, 0x76, 0x32,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ultron_proto_rawDescOnce sync.Once
	file_ultron_proto_rawDescData = file_ultron_proto_rawDesc
)

func file_ultron_proto_rawDescGZIP() []byte {
	file_ultron_proto_rawDescOnce.Do(func() {
		file_ultron_proto_rawDescData = protoimpl.X.CompressGZIP(file_ultron_proto_rawDescData)
	})
	return file_ultron_proto_rawDescData
}

var file_ultron_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_ultron_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_ultron_proto_goTypes = []interface{}{
	(EventType)(0),                          // 0: wosai.ultron.EventType
	(ResponseSubmit_Result)(0),              // 1: wosai.ultron.ResponseSubmit.Result
	(*Session)(nil),                         // 2: wosai.ultron.Session
	(*TimerDTO)(nil),                        // 3: wosai.ultron.TimerDTO
	(*AttackStrategyDTO)(nil),               // 4: wosai.ultron.AttackStrategyDTO
	(*Event)(nil),                           // 5: wosai.ultron.Event
	(*RequestSubmit)(nil),                   // 6: wosai.ultron.RequestSubmit
	(*ResponseSubmit)(nil),                  // 7: wosai.ultron.ResponseSubmit
	nil,                                     // 8: wosai.ultron.Session.ExtrasEntry
	(*statistics.StatisticianGroupDTO)(nil), // 9: wosai.ultron.StatisticianGroupDTO
}
var file_ultron_proto_depIdxs = []int32{
	8, // 0: wosai.ultron.Session.extras:type_name -> wosai.ultron.Session.ExtrasEntry
	0, // 1: wosai.ultron.Event.type:type_name -> wosai.ultron.EventType
	4, // 2: wosai.ultron.Event.attack_strategy:type_name -> wosai.ultron.AttackStrategyDTO
	3, // 3: wosai.ultron.Event.timer:type_name -> wosai.ultron.TimerDTO
	9, // 4: wosai.ultron.RequestSubmit.stats:type_name -> wosai.ultron.StatisticianGroupDTO
	1, // 5: wosai.ultron.ResponseSubmit.result:type_name -> wosai.ultron.ResponseSubmit.Result
	2, // 6: wosai.ultron.UltronService.Subscribe:input_type -> wosai.ultron.Session
	6, // 7: wosai.ultron.UltronService.Submit:input_type -> wosai.ultron.RequestSubmit
	5, // 8: wosai.ultron.UltronService.Subscribe:output_type -> wosai.ultron.Event
	7, // 9: wosai.ultron.UltronService.Submit:output_type -> wosai.ultron.ResponseSubmit
	8, // [8:10] is the sub-list for method output_type
	6, // [6:8] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_ultron_proto_init() }
func file_ultron_proto_init() {
	if File_ultron_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ultron_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Session); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ultron_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimerDTO); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ultron_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AttackStrategyDTO); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ultron_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ultron_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestSubmit); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ultron_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResponseSubmit); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_ultron_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*Event_PlanName)(nil),
		(*Event_AttackStrategy)(nil),
		(*Event_Timer)(nil),
		(*Event_BatchId)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ultron_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ultron_proto_goTypes,
		DependencyIndexes: file_ultron_proto_depIdxs,
		EnumInfos:         file_ultron_proto_enumTypes,
		MessageInfos:      file_ultron_proto_msgTypes,
	}.Build()
	File_ultron_proto = out.File
	file_ultron_proto_rawDesc = nil
	file_ultron_proto_goTypes = nil
	file_ultron_proto_depIdxs = nil
}
