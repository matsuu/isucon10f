// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.4
// source: xsuportal/services/admin/teams.proto

package admin

import (
	proto "github.com/golang/protobuf/proto"
	resources "github.com/isucon/isucon10-final/webapp/golang/proto/xsuportal/resources"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type ListTeamsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListTeamsRequest) Reset() {
	*x = ListTeamsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListTeamsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTeamsRequest) ProtoMessage() {}

func (x *ListTeamsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTeamsRequest.ProtoReflect.Descriptor instead.
func (*ListTeamsRequest) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{0}
}

type ListTeamsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Teams []*ListTeamsResponse_TeamListItem `protobuf:"bytes,1,rep,name=teams,proto3" json:"teams,omitempty"`
}

func (x *ListTeamsResponse) Reset() {
	*x = ListTeamsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListTeamsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTeamsResponse) ProtoMessage() {}

func (x *ListTeamsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTeamsResponse.ProtoReflect.Descriptor instead.
func (*ListTeamsResponse) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{1}
}

func (x *ListTeamsResponse) GetTeams() []*ListTeamsResponse_TeamListItem {
	if x != nil {
		return x.Teams
	}
	return nil
}

type GetTeamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetTeamRequest) Reset() {
	*x = GetTeamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTeamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTeamRequest) ProtoMessage() {}

func (x *GetTeamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTeamRequest.ProtoReflect.Descriptor instead.
func (*GetTeamRequest) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{2}
}

func (x *GetTeamRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetTeamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Team *resources.Team `protobuf:"bytes,1,opt,name=team,proto3" json:"team,omitempty"`
}

func (x *GetTeamResponse) Reset() {
	*x = GetTeamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTeamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTeamResponse) ProtoMessage() {}

func (x *GetTeamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTeamResponse.ProtoReflect.Descriptor instead.
func (*GetTeamResponse) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{3}
}

func (x *GetTeamResponse) GetTeam() *resources.Team {
	if x != nil {
		return x.Team
	}
	return nil
}

type UpdateTeamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Team        *resources.Team         `protobuf:"bytes,1,opt,name=team,proto3" json:"team,omitempty"`
	Contestants []*resources.Contestant `protobuf:"bytes,2,rep,name=contestants,proto3" json:"contestants,omitempty"`
}

func (x *UpdateTeamRequest) Reset() {
	*x = UpdateTeamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateTeamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTeamRequest) ProtoMessage() {}

func (x *UpdateTeamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTeamRequest.ProtoReflect.Descriptor instead.
func (*UpdateTeamRequest) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateTeamRequest) GetTeam() *resources.Team {
	if x != nil {
		return x.Team
	}
	return nil
}

func (x *UpdateTeamRequest) GetContestants() []*resources.Contestant {
	if x != nil {
		return x.Contestants
	}
	return nil
}

type UpdateTeamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateTeamResponse) Reset() {
	*x = UpdateTeamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateTeamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTeamResponse) ProtoMessage() {}

func (x *UpdateTeamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTeamResponse.ProtoReflect.Descriptor instead.
func (*UpdateTeamResponse) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{5}
}

type ListTeamsResponse_TeamListItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TeamId      int64    `protobuf:"varint,1,opt,name=team_id,json=teamId,proto3" json:"team_id,omitempty"`
	Name        string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	MemberNames []string `protobuf:"bytes,3,rep,name=member_names,json=memberNames,proto3" json:"member_names,omitempty"`
	IsStudent   bool     `protobuf:"varint,5,opt,name=is_student,json=isStudent,proto3" json:"is_student,omitempty"`
	Withdrawn   bool     `protobuf:"varint,6,opt,name=withdrawn,proto3" json:"withdrawn,omitempty"`
}

func (x *ListTeamsResponse_TeamListItem) Reset() {
	*x = ListTeamsResponse_TeamListItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xsuportal_services_admin_teams_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListTeamsResponse_TeamListItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTeamsResponse_TeamListItem) ProtoMessage() {}

func (x *ListTeamsResponse_TeamListItem) ProtoReflect() protoreflect.Message {
	mi := &file_xsuportal_services_admin_teams_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTeamsResponse_TeamListItem.ProtoReflect.Descriptor instead.
func (*ListTeamsResponse_TeamListItem) Descriptor() ([]byte, []int) {
	return file_xsuportal_services_admin_teams_proto_rawDescGZIP(), []int{1, 0}
}

func (x *ListTeamsResponse_TeamListItem) GetTeamId() int64 {
	if x != nil {
		return x.TeamId
	}
	return 0
}

func (x *ListTeamsResponse_TeamListItem) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ListTeamsResponse_TeamListItem) GetMemberNames() []string {
	if x != nil {
		return x.MemberNames
	}
	return nil
}

func (x *ListTeamsResponse_TeamListItem) GetIsStudent() bool {
	if x != nil {
		return x.IsStudent
	}
	return false
}

func (x *ListTeamsResponse_TeamListItem) GetWithdrawn() bool {
	if x != nil {
		return x.Withdrawn
	}
	return false
}

var File_xsuportal_services_admin_teams_proto protoreflect.FileDescriptor

var file_xsuportal_services_admin_teams_proto_rawDesc = []byte{
	0x0a, 0x24, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x1a, 0x1e, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x12, 0x0a, 0x10,
	0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x87, 0x02, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x54, 0x0a, 0x05, 0x74, 0x65, 0x61, 0x6d, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3e, 0x2e, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x4c, 0x69, 0x73,
	0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x74, 0x65, 0x61, 0x6d, 0x73, 0x1a, 0x9b, 0x01, 0x0a,
	0x0c, 0x54, 0x65, 0x61, 0x6d, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x17, 0x0a,
	0x07, 0x74, 0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06,
	0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x6d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x0b, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x1d, 0x0a,
	0x0a, 0x69, 0x73, 0x5f, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x09, 0x69, 0x73, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x77, 0x69, 0x74, 0x68, 0x64, 0x72, 0x61, 0x77, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x09, 0x77, 0x69, 0x74, 0x68, 0x64, 0x72, 0x61, 0x77, 0x6e, 0x22, 0x20, 0x0a, 0x0e, 0x47, 0x65,
	0x74, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22, 0x46, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x33, 0x0a, 0x04, 0x74, 0x65, 0x61, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e,
	0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x04,
	0x74, 0x65, 0x61, 0x6d, 0x22, 0x91, 0x01, 0x0a, 0x11, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54,
	0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x04, 0x74, 0x65,
	0x61, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x78, 0x73, 0x75, 0x70, 0x6f,
	0x72, 0x74, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x04, 0x74, 0x65, 0x61, 0x6d, 0x12,
	0x47, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x52, 0x0b, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x73, 0x22, 0x14, 0x0a, 0x12, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x4f,
	0x5a, 0x4d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x73, 0x75,
	0x63, 0x6f, 0x6e, 0x2f, 0x69, 0x73, 0x75, 0x63, 0x6f, 0x6e, 0x31, 0x30, 0x2d, 0x66, 0x69, 0x6e,
	0x61, 0x6c, 0x2f, 0x77, 0x65, 0x62, 0x61, 0x70, 0x70, 0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x78, 0x73, 0x75, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xsuportal_services_admin_teams_proto_rawDescOnce sync.Once
	file_xsuportal_services_admin_teams_proto_rawDescData = file_xsuportal_services_admin_teams_proto_rawDesc
)

func file_xsuportal_services_admin_teams_proto_rawDescGZIP() []byte {
	file_xsuportal_services_admin_teams_proto_rawDescOnce.Do(func() {
		file_xsuportal_services_admin_teams_proto_rawDescData = protoimpl.X.CompressGZIP(file_xsuportal_services_admin_teams_proto_rawDescData)
	})
	return file_xsuportal_services_admin_teams_proto_rawDescData
}

var file_xsuportal_services_admin_teams_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_xsuportal_services_admin_teams_proto_goTypes = []interface{}{
	(*ListTeamsRequest)(nil),               // 0: xsuportal.proto.services.admin.ListTeamsRequest
	(*ListTeamsResponse)(nil),              // 1: xsuportal.proto.services.admin.ListTeamsResponse
	(*GetTeamRequest)(nil),                 // 2: xsuportal.proto.services.admin.GetTeamRequest
	(*GetTeamResponse)(nil),                // 3: xsuportal.proto.services.admin.GetTeamResponse
	(*UpdateTeamRequest)(nil),              // 4: xsuportal.proto.services.admin.UpdateTeamRequest
	(*UpdateTeamResponse)(nil),             // 5: xsuportal.proto.services.admin.UpdateTeamResponse
	(*ListTeamsResponse_TeamListItem)(nil), // 6: xsuportal.proto.services.admin.ListTeamsResponse.TeamListItem
	(*resources.Team)(nil),                 // 7: xsuportal.proto.resources.Team
	(*resources.Contestant)(nil),           // 8: xsuportal.proto.resources.Contestant
}
var file_xsuportal_services_admin_teams_proto_depIdxs = []int32{
	6, // 0: xsuportal.proto.services.admin.ListTeamsResponse.teams:type_name -> xsuportal.proto.services.admin.ListTeamsResponse.TeamListItem
	7, // 1: xsuportal.proto.services.admin.GetTeamResponse.team:type_name -> xsuportal.proto.resources.Team
	7, // 2: xsuportal.proto.services.admin.UpdateTeamRequest.team:type_name -> xsuportal.proto.resources.Team
	8, // 3: xsuportal.proto.services.admin.UpdateTeamRequest.contestants:type_name -> xsuportal.proto.resources.Contestant
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_xsuportal_services_admin_teams_proto_init() }
func file_xsuportal_services_admin_teams_proto_init() {
	if File_xsuportal_services_admin_teams_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_xsuportal_services_admin_teams_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListTeamsRequest); i {
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
		file_xsuportal_services_admin_teams_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListTeamsResponse); i {
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
		file_xsuportal_services_admin_teams_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTeamRequest); i {
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
		file_xsuportal_services_admin_teams_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTeamResponse); i {
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
		file_xsuportal_services_admin_teams_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateTeamRequest); i {
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
		file_xsuportal_services_admin_teams_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateTeamResponse); i {
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
		file_xsuportal_services_admin_teams_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListTeamsResponse_TeamListItem); i {
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
			RawDescriptor: file_xsuportal_services_admin_teams_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_xsuportal_services_admin_teams_proto_goTypes,
		DependencyIndexes: file_xsuportal_services_admin_teams_proto_depIdxs,
		MessageInfos:      file_xsuportal_services_admin_teams_proto_msgTypes,
	}.Build()
	File_xsuportal_services_admin_teams_proto = out.File
	file_xsuportal_services_admin_teams_proto_rawDesc = nil
	file_xsuportal_services_admin_teams_proto_goTypes = nil
	file_xsuportal_services_admin_teams_proto_depIdxs = nil
}
