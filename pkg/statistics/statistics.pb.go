// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.18.1
// source: statistics.proto

package statistics

import (
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

type AttackStatisticsDTO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name                string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Requests            uint64            `protobuf:"varint,2,opt,name=requests,proto3" json:"requests,omitempty"`
	Failures            uint64            `protobuf:"varint,3,opt,name=failures,proto3" json:"failures,omitempty"`
	TotalResponseTime   int64             `protobuf:"varint,4,opt,name=total_response_time,json=totalResponseTime,proto3" json:"total_response_time,omitempty"`
	MinResponseTime     int64             `protobuf:"varint,5,opt,name=min_response_time,json=minResponseTime,proto3" json:"min_response_time,omitempty"`
	MaxResponseTime     int64             `protobuf:"varint,6,opt,name=max_response_time,json=maxResponseTime,proto3" json:"max_response_time,omitempty"`
	RecentSuccessBucket map[int64]int64   `protobuf:"bytes,7,rep,name=recent_success_bucket,json=recentSuccessBucket,proto3" json:"recent_success_bucket,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	RecentFailureBucket map[int64]int64   `protobuf:"bytes,8,rep,name=recent_failure_bucket,json=recentFailureBucket,proto3" json:"recent_failure_bucket,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ResponseBucket      map[int64]uint64  `protobuf:"bytes,9,rep,name=response_bucket,json=responseBucket,proto3" json:"response_bucket,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	FailureBucket       map[string]uint64 `protobuf:"bytes,10,rep,name=failure_bucket,json=failureBucket,proto3" json:"failure_bucket,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	FirstAttack         int64             `protobuf:"varint,11,opt,name=first_attack,json=firstAttack,proto3" json:"first_attack,omitempty"`
	LastAttack          int64             `protobuf:"varint,12,opt,name=last_attack,json=lastAttack,proto3" json:"last_attack,omitempty"`
	Interval            int64             `protobuf:"varint,13,opt,name=interval,proto3" json:"interval,omitempty"`
}

func (x *AttackStatisticsDTO) Reset() {
	*x = AttackStatisticsDTO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_statistics_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AttackStatisticsDTO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AttackStatisticsDTO) ProtoMessage() {}

func (x *AttackStatisticsDTO) ProtoReflect() protoreflect.Message {
	mi := &file_statistics_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AttackStatisticsDTO.ProtoReflect.Descriptor instead.
func (*AttackStatisticsDTO) Descriptor() ([]byte, []int) {
	return file_statistics_proto_rawDescGZIP(), []int{0}
}

func (x *AttackStatisticsDTO) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AttackStatisticsDTO) GetRequests() uint64 {
	if x != nil {
		return x.Requests
	}
	return 0
}

func (x *AttackStatisticsDTO) GetFailures() uint64 {
	if x != nil {
		return x.Failures
	}
	return 0
}

func (x *AttackStatisticsDTO) GetTotalResponseTime() int64 {
	if x != nil {
		return x.TotalResponseTime
	}
	return 0
}

func (x *AttackStatisticsDTO) GetMinResponseTime() int64 {
	if x != nil {
		return x.MinResponseTime
	}
	return 0
}

func (x *AttackStatisticsDTO) GetMaxResponseTime() int64 {
	if x != nil {
		return x.MaxResponseTime
	}
	return 0
}

func (x *AttackStatisticsDTO) GetRecentSuccessBucket() map[int64]int64 {
	if x != nil {
		return x.RecentSuccessBucket
	}
	return nil
}

func (x *AttackStatisticsDTO) GetRecentFailureBucket() map[int64]int64 {
	if x != nil {
		return x.RecentFailureBucket
	}
	return nil
}

func (x *AttackStatisticsDTO) GetResponseBucket() map[int64]uint64 {
	if x != nil {
		return x.ResponseBucket
	}
	return nil
}

func (x *AttackStatisticsDTO) GetFailureBucket() map[string]uint64 {
	if x != nil {
		return x.FailureBucket
	}
	return nil
}

func (x *AttackStatisticsDTO) GetFirstAttack() int64 {
	if x != nil {
		return x.FirstAttack
	}
	return 0
}

func (x *AttackStatisticsDTO) GetLastAttack() int64 {
	if x != nil {
		return x.LastAttack
	}
	return 0
}

func (x *AttackStatisticsDTO) GetInterval() int64 {
	if x != nil {
		return x.Interval
	}
	return 0
}

var File_statistics_proto protoreflect.FileDescriptor

var file_statistics_proto_rawDesc = []byte{
	0x0a, 0x10, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0c, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e,
	0x22, 0xfb, 0x07, 0x0a, 0x13, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x69,
	0x73, 0x74, 0x69, 0x63, 0x73, 0x44, 0x54, 0x4f, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x73, 0x12, 0x2e, 0x0a, 0x13, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x11, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x2a, 0x0a, 0x11, 0x6d, 0x69, 0x6e, 0x5f, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0f, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x2a, 0x0a, 0x11, 0x6d, 0x61, 0x78, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x6d, 0x61, 0x78,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x6e, 0x0a, 0x15,
	0x72, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x62,
	0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3a, 0x2e, 0x77, 0x6f,
	0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63,
	0x6b, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x44, 0x54, 0x4f, 0x2e, 0x52,
	0x65, 0x63, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x42, 0x75, 0x63, 0x6b,
	0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x13, 0x72, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x53,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x6e, 0x0a, 0x15,
	0x72, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x62,
	0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3a, 0x2e, 0x77, 0x6f,
	0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63,
	0x6b, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x44, 0x54, 0x4f, 0x2e, 0x52,
	0x65, 0x63, 0x65, 0x6e, 0x74, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x42, 0x75, 0x63, 0x6b,
	0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x13, 0x72, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x46,
	0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x5e, 0x0a, 0x0f,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18,
	0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c,
	0x74, 0x72, 0x6f, 0x6e, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x69,
	0x73, 0x74, 0x69, 0x63, 0x73, 0x44, 0x54, 0x4f, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0e, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x5b, 0x0a, 0x0e,
	0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x0a,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x77, 0x6f, 0x73, 0x61, 0x69, 0x2e, 0x75, 0x6c, 0x74,
	0x72, 0x6f, 0x6e, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73,
	0x74, 0x69, 0x63, 0x73, 0x44, 0x54, 0x4f, 0x2e, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x42,
	0x75, 0x63, 0x6b, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0d, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x69, 0x72,
	0x73, 0x74, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0b, 0x66, 0x69, 0x72, 0x73, 0x74, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x12, 0x1f, 0x0a, 0x0b,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0a, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x74, 0x74, 0x61, 0x63, 0x6b, 0x12, 0x1a, 0x0a,
	0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x1a, 0x46, 0x0a, 0x18, 0x52, 0x65, 0x63,
	0x65, 0x6e, 0x74, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x1a, 0x46, 0x0a, 0x18, 0x52, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x46, 0x61, 0x69, 0x6c, 0x75,
	0x72, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x41, 0x0a, 0x13, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x40, 0x0a, 0x12,
	0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x2b,
	0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x6f, 0x73,
	0x61, 0x69, 0x2f, 0x75, 0x6c, 0x74, 0x72, 0x6f, 0x6e, 0x2f, 0x76, 0x32, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_statistics_proto_rawDescOnce sync.Once
	file_statistics_proto_rawDescData = file_statistics_proto_rawDesc
)

func file_statistics_proto_rawDescGZIP() []byte {
	file_statistics_proto_rawDescOnce.Do(func() {
		file_statistics_proto_rawDescData = protoimpl.X.CompressGZIP(file_statistics_proto_rawDescData)
	})
	return file_statistics_proto_rawDescData
}

var file_statistics_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_statistics_proto_goTypes = []interface{}{
	(*AttackStatisticsDTO)(nil), // 0: wosai.ultron.AttackStatisticsDTO
	nil,                         // 1: wosai.ultron.AttackStatisticsDTO.RecentSuccessBucketEntry
	nil,                         // 2: wosai.ultron.AttackStatisticsDTO.RecentFailureBucketEntry
	nil,                         // 3: wosai.ultron.AttackStatisticsDTO.ResponseBucketEntry
	nil,                         // 4: wosai.ultron.AttackStatisticsDTO.FailureBucketEntry
}
var file_statistics_proto_depIdxs = []int32{
	1, // 0: wosai.ultron.AttackStatisticsDTO.recent_success_bucket:type_name -> wosai.ultron.AttackStatisticsDTO.RecentSuccessBucketEntry
	2, // 1: wosai.ultron.AttackStatisticsDTO.recent_failure_bucket:type_name -> wosai.ultron.AttackStatisticsDTO.RecentFailureBucketEntry
	3, // 2: wosai.ultron.AttackStatisticsDTO.response_bucket:type_name -> wosai.ultron.AttackStatisticsDTO.ResponseBucketEntry
	4, // 3: wosai.ultron.AttackStatisticsDTO.failure_bucket:type_name -> wosai.ultron.AttackStatisticsDTO.FailureBucketEntry
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_statistics_proto_init() }
func file_statistics_proto_init() {
	if File_statistics_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_statistics_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AttackStatisticsDTO); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_statistics_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_statistics_proto_goTypes,
		DependencyIndexes: file_statistics_proto_depIdxs,
		MessageInfos:      file_statistics_proto_msgTypes,
	}.Build()
	File_statistics_proto = out.File
	file_statistics_proto_rawDesc = nil
	file_statistics_proto_goTypes = nil
	file_statistics_proto_depIdxs = nil
}
