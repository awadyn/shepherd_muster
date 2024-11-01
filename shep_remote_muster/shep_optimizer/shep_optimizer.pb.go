// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.12.4
// source: shep_optimizer/shep_optimizer.proto

package shep_remote_muster

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

type ControlEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Knob string `protobuf:"bytes,1,opt,name=knob,proto3" json:"knob,omitempty"`
	Val  uint64 `protobuf:"varint,2,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *ControlEntry) Reset() {
	*x = ControlEntry{}
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ControlEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ControlEntry) ProtoMessage() {}

func (x *ControlEntry) ProtoReflect() protoreflect.Message {
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ControlEntry.ProtoReflect.Descriptor instead.
func (*ControlEntry) Descriptor() ([]byte, []int) {
	return file_shep_optimizer_shep_optimizer_proto_rawDescGZIP(), []int{0}
}

func (x *ControlEntry) GetKnob() string {
	if x != nil {
		return x.Knob
	}
	return ""
}

func (x *ControlEntry) GetVal() uint64 {
	if x != nil {
		return x.Val
	}
	return 0
}

type RewardEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id  string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Val uint64 `protobuf:"varint,2,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *RewardEntry) Reset() {
	*x = RewardEntry{}
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RewardEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RewardEntry) ProtoMessage() {}

func (x *RewardEntry) ProtoReflect() protoreflect.Message {
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RewardEntry.ProtoReflect.Descriptor instead.
func (*RewardEntry) Descriptor() ([]byte, []int) {
	return file_shep_optimizer_shep_optimizer_proto_rawDescGZIP(), []int{1}
}

func (x *RewardEntry) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RewardEntry) GetVal() uint64 {
	if x != nil {
		return x.Val
	}
	return 0
}

type StartOptimizerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NTrials uint32 `protobuf:"varint,1,opt,name=n_trials,json=nTrials,proto3" json:"n_trials,omitempty"`
}

func (x *StartOptimizerRequest) Reset() {
	*x = StartOptimizerRequest{}
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartOptimizerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartOptimizerRequest) ProtoMessage() {}

func (x *StartOptimizerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartOptimizerRequest.ProtoReflect.Descriptor instead.
func (*StartOptimizerRequest) Descriptor() ([]byte, []int) {
	return file_shep_optimizer_shep_optimizer_proto_rawDescGZIP(), []int{2}
}

func (x *StartOptimizerRequest) GetNTrials() uint32 {
	if x != nil {
		return x.NTrials
	}
	return 0
}

type StartOptimizerReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Done  bool            `protobuf:"varint,1,opt,name=done,proto3" json:"done,omitempty"`
	Ctrls []*ControlEntry `protobuf:"bytes,2,rep,name=ctrls,proto3" json:"ctrls,omitempty"`
}

func (x *StartOptimizerReply) Reset() {
	*x = StartOptimizerReply{}
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartOptimizerReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartOptimizerReply) ProtoMessage() {}

func (x *StartOptimizerReply) ProtoReflect() protoreflect.Message {
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartOptimizerReply.ProtoReflect.Descriptor instead.
func (*StartOptimizerReply) Descriptor() ([]byte, []int) {
	return file_shep_optimizer_shep_optimizer_proto_rawDescGZIP(), []int{3}
}

func (x *StartOptimizerReply) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *StartOptimizerReply) GetCtrls() []*ControlEntry {
	if x != nil {
		return x.Ctrls
	}
	return nil
}

type OptimizeRewardRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rewards []*RewardEntry `protobuf:"bytes,1,rep,name=rewards,proto3" json:"rewards,omitempty"`
}

func (x *OptimizeRewardRequest) Reset() {
	*x = OptimizeRewardRequest{}
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OptimizeRewardRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OptimizeRewardRequest) ProtoMessage() {}

func (x *OptimizeRewardRequest) ProtoReflect() protoreflect.Message {
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OptimizeRewardRequest.ProtoReflect.Descriptor instead.
func (*OptimizeRewardRequest) Descriptor() ([]byte, []int) {
	return file_shep_optimizer_shep_optimizer_proto_rawDescGZIP(), []int{4}
}

func (x *OptimizeRewardRequest) GetRewards() []*RewardEntry {
	if x != nil {
		return x.Rewards
	}
	return nil
}

type OptimizeRewardReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Done  bool            `protobuf:"varint,1,opt,name=done,proto3" json:"done,omitempty"`
	Ctrls []*ControlEntry `protobuf:"bytes,2,rep,name=ctrls,proto3" json:"ctrls,omitempty"`
}

func (x *OptimizeRewardReply) Reset() {
	*x = OptimizeRewardReply{}
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OptimizeRewardReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OptimizeRewardReply) ProtoMessage() {}

func (x *OptimizeRewardReply) ProtoReflect() protoreflect.Message {
	mi := &file_shep_optimizer_shep_optimizer_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OptimizeRewardReply.ProtoReflect.Descriptor instead.
func (*OptimizeRewardReply) Descriptor() ([]byte, []int) {
	return file_shep_optimizer_shep_optimizer_proto_rawDescGZIP(), []int{5}
}

func (x *OptimizeRewardReply) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

func (x *OptimizeRewardReply) GetCtrls() []*ControlEntry {
	if x != nil {
		return x.Ctrls
	}
	return nil
}

var File_shep_optimizer_shep_optimizer_proto protoreflect.FileDescriptor

var file_shep_optimizer_shep_optimizer_proto_rawDesc = []byte{
	0x0a, 0x23, 0x73, 0x68, 0x65, 0x70, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72,
	0x2f, 0x73, 0x68, 0x65, 0x70, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x68, 0x65, 0x70, 0x68, 0x65, 0x72, 0x64, 0x22,
	0x34, 0x0a, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x12, 0x0a, 0x04, 0x6b, 0x6e, 0x6f, 0x62, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b,
	0x6e, 0x6f, 0x62, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x2f, 0x0a, 0x0b, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x32, 0x0a, 0x15, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4f,
	0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x19, 0x0a, 0x08, 0x6e, 0x5f, 0x74, 0x72, 0x69, 0x61, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x07, 0x6e, 0x54, 0x72, 0x69, 0x61, 0x6c, 0x73, 0x22, 0x57, 0x0a, 0x13, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x04, 0x64, 0x6f, 0x6e, 0x65, 0x12, 0x2c, 0x0a, 0x05, 0x63, 0x74, 0x72, 0x6c, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x68, 0x65, 0x70, 0x68, 0x65, 0x72, 0x64, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x63, 0x74,
	0x72, 0x6c, 0x73, 0x22, 0x48, 0x0a, 0x15, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x52,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2f, 0x0a, 0x07,
	0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e,
	0x73, 0x68, 0x65, 0x70, 0x68, 0x65, 0x72, 0x64, 0x2e, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x73, 0x22, 0x57, 0x0a,
	0x13, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x12, 0x2c, 0x0a, 0x05, 0x63, 0x74, 0x72, 0x6c,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x68, 0x65, 0x70, 0x68, 0x65,
	0x72, 0x64, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x05, 0x63, 0x74, 0x72, 0x6c, 0x73, 0x32, 0xae, 0x01, 0x0a, 0x08, 0x4f, 0x70, 0x74, 0x69, 0x6d,
	0x69, 0x7a, 0x65, 0x12, 0x50, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4f, 0x70, 0x74, 0x69,
	0x6d, 0x69, 0x7a, 0x65, 0x72, 0x12, 0x1f, 0x2e, 0x73, 0x68, 0x65, 0x70, 0x68, 0x65, 0x72, 0x64,
	0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x73, 0x68, 0x65, 0x70, 0x68, 0x65, 0x72,
	0x64, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x72,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x50, 0x0a, 0x0e, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a,
	0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x12, 0x1f, 0x2e, 0x73, 0x68, 0x65, 0x70, 0x68, 0x65,
	0x72, 0x64, 0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x73, 0x68, 0x65, 0x70, 0x68,
	0x65, 0x72, 0x64, 0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x42, 0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x77, 0x61, 0x64, 0x79, 0x6e, 0x2f, 0x73, 0x68, 0x65,
	0x70, 0x5f, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x5f, 0x6d, 0x75, 0x73, 0x74, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_shep_optimizer_shep_optimizer_proto_rawDescOnce sync.Once
	file_shep_optimizer_shep_optimizer_proto_rawDescData = file_shep_optimizer_shep_optimizer_proto_rawDesc
)

func file_shep_optimizer_shep_optimizer_proto_rawDescGZIP() []byte {
	file_shep_optimizer_shep_optimizer_proto_rawDescOnce.Do(func() {
		file_shep_optimizer_shep_optimizer_proto_rawDescData = protoimpl.X.CompressGZIP(file_shep_optimizer_shep_optimizer_proto_rawDescData)
	})
	return file_shep_optimizer_shep_optimizer_proto_rawDescData
}

var file_shep_optimizer_shep_optimizer_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_shep_optimizer_shep_optimizer_proto_goTypes = []any{
	(*ControlEntry)(nil),          // 0: shepherd.ControlEntry
	(*RewardEntry)(nil),           // 1: shepherd.RewardEntry
	(*StartOptimizerRequest)(nil), // 2: shepherd.StartOptimizerRequest
	(*StartOptimizerReply)(nil),   // 3: shepherd.StartOptimizerReply
	(*OptimizeRewardRequest)(nil), // 4: shepherd.OptimizeRewardRequest
	(*OptimizeRewardReply)(nil),   // 5: shepherd.OptimizeRewardReply
}
var file_shep_optimizer_shep_optimizer_proto_depIdxs = []int32{
	0, // 0: shepherd.StartOptimizerReply.ctrls:type_name -> shepherd.ControlEntry
	1, // 1: shepherd.OptimizeRewardRequest.rewards:type_name -> shepherd.RewardEntry
	0, // 2: shepherd.OptimizeRewardReply.ctrls:type_name -> shepherd.ControlEntry
	2, // 3: shepherd.Optimize.StartOptimizer:input_type -> shepherd.StartOptimizerRequest
	4, // 4: shepherd.Optimize.OptimizeReward:input_type -> shepherd.OptimizeRewardRequest
	3, // 5: shepherd.Optimize.StartOptimizer:output_type -> shepherd.StartOptimizerReply
	5, // 6: shepherd.Optimize.OptimizeReward:output_type -> shepherd.OptimizeRewardReply
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_shep_optimizer_shep_optimizer_proto_init() }
func file_shep_optimizer_shep_optimizer_proto_init() {
	if File_shep_optimizer_shep_optimizer_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_shep_optimizer_shep_optimizer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_shep_optimizer_shep_optimizer_proto_goTypes,
		DependencyIndexes: file_shep_optimizer_shep_optimizer_proto_depIdxs,
		MessageInfos:      file_shep_optimizer_shep_optimizer_proto_msgTypes,
	}.Build()
	File_shep_optimizer_shep_optimizer_proto = out.File
	file_shep_optimizer_shep_optimizer_proto_rawDesc = nil
	file_shep_optimizer_shep_optimizer_proto_goTypes = nil
	file_shep_optimizer_shep_optimizer_proto_depIdxs = nil
}