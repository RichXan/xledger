// Code generated by protoc-gen-go. DO NOT EDIT.
// source: user.proto

package user

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type User struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Username             string   `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Email                string   `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	CreatedAt            string   `protobuf:"bytes,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt            string   `protobuf:"bytes,5,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{0}
}

func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (m *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(m, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *User) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *User) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *User) GetCreatedAt() string {
	if m != nil {
		return m.CreatedAt
	}
	return ""
}

func (m *User) GetUpdatedAt() string {
	if m != nil {
		return m.UpdatedAt
	}
	return ""
}

type RegisterRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Email                string   `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
	Password             string   `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterRequest) Reset()         { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()    {}
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{1}
}

func (m *RegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterRequest.Unmarshal(m, b)
}
func (m *RegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterRequest.Marshal(b, m, deterministic)
}
func (m *RegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterRequest.Merge(m, src)
}
func (m *RegisterRequest) XXX_Size() int {
	return xxx_messageInfo_RegisterRequest.Size(m)
}
func (m *RegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterRequest proto.InternalMessageInfo

func (m *RegisterRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *RegisterRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *RegisterRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type RegisterResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	User                 *User    `protobuf:"bytes,3,opt,name=user,proto3" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterResponse) Reset()         { *m = RegisterResponse{} }
func (m *RegisterResponse) String() string { return proto.CompactTextString(m) }
func (*RegisterResponse) ProtoMessage()    {}
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{2}
}

func (m *RegisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterResponse.Unmarshal(m, b)
}
func (m *RegisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterResponse.Marshal(b, m, deterministic)
}
func (m *RegisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterResponse.Merge(m, src)
}
func (m *RegisterResponse) XXX_Size() int {
	return xxx_messageInfo_RegisterResponse.Size(m)
}
func (m *RegisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterResponse proto.InternalMessageInfo

func (m *RegisterResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *RegisterResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *RegisterResponse) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

type LoginRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password             string   `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginRequest) Reset()         { *m = LoginRequest{} }
func (m *LoginRequest) String() string { return proto.CompactTextString(m) }
func (*LoginRequest) ProtoMessage()    {}
func (*LoginRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{3}
}

func (m *LoginRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginRequest.Unmarshal(m, b)
}
func (m *LoginRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginRequest.Marshal(b, m, deterministic)
}
func (m *LoginRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginRequest.Merge(m, src)
}
func (m *LoginRequest) XXX_Size() int {
	return xxx_messageInfo_LoginRequest.Size(m)
}
func (m *LoginRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginRequest.DiscardUnknown(m)
}

var xxx_messageInfo_LoginRequest proto.InternalMessageInfo

func (m *LoginRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *LoginRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type LoginResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Token                string   `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
	User                 *User    `protobuf:"bytes,4,opt,name=user,proto3" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginResponse) Reset()         { *m = LoginResponse{} }
func (m *LoginResponse) String() string { return proto.CompactTextString(m) }
func (*LoginResponse) ProtoMessage()    {}
func (*LoginResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{4}
}

func (m *LoginResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginResponse.Unmarshal(m, b)
}
func (m *LoginResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginResponse.Marshal(b, m, deterministic)
}
func (m *LoginResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginResponse.Merge(m, src)
}
func (m *LoginResponse) XXX_Size() int {
	return xxx_messageInfo_LoginResponse.Size(m)
}
func (m *LoginResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginResponse.DiscardUnknown(m)
}

var xxx_messageInfo_LoginResponse proto.InternalMessageInfo

func (m *LoginResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *LoginResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *LoginResponse) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *LoginResponse) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

type GetUserInfoRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetUserInfoRequest) Reset()         { *m = GetUserInfoRequest{} }
func (m *GetUserInfoRequest) String() string { return proto.CompactTextString(m) }
func (*GetUserInfoRequest) ProtoMessage()    {}
func (*GetUserInfoRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{5}
}

func (m *GetUserInfoRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUserInfoRequest.Unmarshal(m, b)
}
func (m *GetUserInfoRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUserInfoRequest.Marshal(b, m, deterministic)
}
func (m *GetUserInfoRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUserInfoRequest.Merge(m, src)
}
func (m *GetUserInfoRequest) XXX_Size() int {
	return xxx_messageInfo_GetUserInfoRequest.Size(m)
}
func (m *GetUserInfoRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUserInfoRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetUserInfoRequest proto.InternalMessageInfo

func (m *GetUserInfoRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

type GetUserInfoResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	User                 *User    `protobuf:"bytes,3,opt,name=user,proto3" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetUserInfoResponse) Reset()         { *m = GetUserInfoResponse{} }
func (m *GetUserInfoResponse) String() string { return proto.CompactTextString(m) }
func (*GetUserInfoResponse) ProtoMessage()    {}
func (*GetUserInfoResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{6}
}

func (m *GetUserInfoResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUserInfoResponse.Unmarshal(m, b)
}
func (m *GetUserInfoResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUserInfoResponse.Marshal(b, m, deterministic)
}
func (m *GetUserInfoResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUserInfoResponse.Merge(m, src)
}
func (m *GetUserInfoResponse) XXX_Size() int {
	return xxx_messageInfo_GetUserInfoResponse.Size(m)
}
func (m *GetUserInfoResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUserInfoResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetUserInfoResponse proto.InternalMessageInfo

func (m *GetUserInfoResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *GetUserInfoResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *GetUserInfoResponse) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

type UpdateUserInfoRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Username             string   `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Email                string   `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateUserInfoRequest) Reset()         { *m = UpdateUserInfoRequest{} }
func (m *UpdateUserInfoRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateUserInfoRequest) ProtoMessage()    {}
func (*UpdateUserInfoRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{7}
}

func (m *UpdateUserInfoRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateUserInfoRequest.Unmarshal(m, b)
}
func (m *UpdateUserInfoRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateUserInfoRequest.Marshal(b, m, deterministic)
}
func (m *UpdateUserInfoRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateUserInfoRequest.Merge(m, src)
}
func (m *UpdateUserInfoRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateUserInfoRequest.Size(m)
}
func (m *UpdateUserInfoRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateUserInfoRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateUserInfoRequest proto.InternalMessageInfo

func (m *UpdateUserInfoRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *UpdateUserInfoRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *UpdateUserInfoRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

type UpdateUserInfoResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	User                 *User    `protobuf:"bytes,3,opt,name=user,proto3" json:"user,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateUserInfoResponse) Reset()         { *m = UpdateUserInfoResponse{} }
func (m *UpdateUserInfoResponse) String() string { return proto.CompactTextString(m) }
func (*UpdateUserInfoResponse) ProtoMessage()    {}
func (*UpdateUserInfoResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_116e343673f7ffaf, []int{8}
}

func (m *UpdateUserInfoResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateUserInfoResponse.Unmarshal(m, b)
}
func (m *UpdateUserInfoResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateUserInfoResponse.Marshal(b, m, deterministic)
}
func (m *UpdateUserInfoResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateUserInfoResponse.Merge(m, src)
}
func (m *UpdateUserInfoResponse) XXX_Size() int {
	return xxx_messageInfo_UpdateUserInfoResponse.Size(m)
}
func (m *UpdateUserInfoResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateUserInfoResponse.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateUserInfoResponse proto.InternalMessageInfo

func (m *UpdateUserInfoResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *UpdateUserInfoResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *UpdateUserInfoResponse) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

func init() {
	proto.RegisterType((*User)(nil), "user.User")
	proto.RegisterType((*RegisterRequest)(nil), "user.RegisterRequest")
	proto.RegisterType((*RegisterResponse)(nil), "user.RegisterResponse")
	proto.RegisterType((*LoginRequest)(nil), "user.LoginRequest")
	proto.RegisterType((*LoginResponse)(nil), "user.LoginResponse")
	proto.RegisterType((*GetUserInfoRequest)(nil), "user.GetUserInfoRequest")
	proto.RegisterType((*GetUserInfoResponse)(nil), "user.GetUserInfoResponse")
	proto.RegisterType((*UpdateUserInfoRequest)(nil), "user.UpdateUserInfoRequest")
	proto.RegisterType((*UpdateUserInfoResponse)(nil), "user.UpdateUserInfoResponse")
}

func init() { proto.RegisterFile("user.proto", fileDescriptor_116e343673f7ffaf) }

var fileDescriptor_116e343673f7ffaf = []byte{
	// 422 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x54, 0x4d, 0x6f, 0xda, 0x40,
	0x10, 0xad, 0x5d, 0x1b, 0xcc, 0x40, 0x69, 0xb5, 0x7c, 0xd4, 0x75, 0x3f, 0x54, 0xf9, 0xd4, 0x4b,
	0x41, 0xa2, 0xc7, 0x9e, 0x88, 0xa2, 0x44, 0x48, 0xc9, 0xc5, 0x11, 0x97, 0x5c, 0x90, 0xb1, 0x07,
	0x64, 0x05, 0x6c, 0xc7, 0xbb, 0x4e, 0x94, 0x3f, 0x90, 0x3f, 0x90, 0x3f, 0x1c, 0xed, 0xda, 0xeb,
	0xd8, 0x0e, 0x48, 0x44, 0x11, 0xb7, 0x7d, 0xf3, 0x86, 0x99, 0xf7, 0x66, 0x06, 0x03, 0xa4, 0x14,
	0x93, 0x51, 0x9c, 0x44, 0x2c, 0x22, 0x1a, 0x7f, 0xdb, 0x8f, 0x0a, 0x68, 0x73, 0x8a, 0x09, 0xe9,
	0x82, 0x1a, 0xf8, 0xa6, 0xf2, 0x5b, 0xf9, 0xd3, 0x72, 0xd4, 0xc0, 0x27, 0x16, 0x18, 0x3c, 0x21,
	0x74, 0xb7, 0x68, 0xaa, 0x22, 0x5a, 0x60, 0xd2, 0x07, 0x1d, 0xb7, 0x6e, 0xb0, 0x31, 0x3f, 0x0a,
	0x22, 0x03, 0xe4, 0x27, 0x80, 0x97, 0xa0, 0xcb, 0xd0, 0x5f, 0xb8, 0xcc, 0xd4, 0x04, 0xd5, 0xca,
	0x23, 0x53, 0xc6, 0xe9, 0x34, 0xf6, 0x25, 0xad, 0x67, 0x74, 0x1e, 0x99, 0x32, 0x7b, 0x01, 0x9f,
	0x1d, 0x5c, 0x07, 0x94, 0x61, 0xe2, 0xe0, 0x6d, 0x8a, 0x94, 0x55, 0x24, 0x28, 0xfb, 0x24, 0xa8,
	0x65, 0x09, 0x16, 0x18, 0xb1, 0x4b, 0xe9, 0x7d, 0x94, 0xf8, 0xb9, 0xb6, 0x02, 0xdb, 0x2b, 0xf8,
	0xf2, 0xd2, 0x80, 0xc6, 0x51, 0x48, 0x91, 0x98, 0xd0, 0xa4, 0xa9, 0xe7, 0x21, 0xa5, 0xa2, 0x81,
	0xe1, 0x48, 0xc8, 0x99, 0x2d, 0x52, 0xea, 0xae, 0xa5, 0x7b, 0x09, 0xc9, 0x2f, 0x10, 0x93, 0x13,
	0xf5, 0xdb, 0x13, 0x18, 0x89, 0x91, 0xf2, 0x11, 0x3a, 0xd9, 0x44, 0xcf, 0xa0, 0x73, 0x11, 0xad,
	0x83, 0xf0, 0x10, 0x17, 0x65, 0xbd, 0x6a, 0x4d, 0xef, 0x03, 0x7c, 0xca, 0xeb, 0xbc, 0x43, 0x6c,
	0x1f, 0x74, 0x16, 0xdd, 0x60, 0x28, 0x37, 0x25, 0x40, 0x61, 0x41, 0xdb, 0x63, 0xe1, 0x2f, 0x90,
	0x73, 0x64, 0x3c, 0x30, 0x0b, 0x57, 0x91, 0x34, 0xf2, 0x15, 0x9a, 0x9c, 0x5d, 0x14, 0x67, 0xd2,
	0xe0, 0x70, 0xe6, 0xdb, 0x01, 0xf4, 0x2a, 0xe9, 0x47, 0x1c, 0xee, 0x12, 0x06, 0x73, 0x71, 0x32,
	0x87, 0x8a, 0x7b, 0xfb, 0x1d, 0xdb, 0x1b, 0x18, 0xd6, 0x7b, 0x1c, 0xcf, 0xd1, 0xe4, 0x49, 0x85,
	0x36, 0x87, 0x57, 0x98, 0xdc, 0x05, 0x1e, 0x92, 0xff, 0x60, 0xc8, 0x33, 0x25, 0x83, 0x2c, 0xbb,
	0xf6, 0xbf, 0xb0, 0x86, 0xf5, 0x70, 0x26, 0xcf, 0xfe, 0x40, 0x26, 0xa0, 0x8b, 0x9b, 0x21, 0x24,
	0x4b, 0x29, 0x1f, 0xa2, 0xd5, 0xab, 0xc4, 0x8a, 0xdf, 0x9c, 0x42, 0xbb, 0xb4, 0x3d, 0x62, 0x66,
	0x59, 0xaf, 0xf7, 0x6f, 0x7d, 0xdb, 0xc1, 0x14, 0x55, 0x2e, 0xa1, 0x5b, 0x1d, 0x1a, 0xf9, 0x9e,
	0x5b, 0xdd, 0xb5, 0x2e, 0xeb, 0xc7, 0x6e, 0x52, 0x96, 0x3b, 0xe9, 0x5e, 0x77, 0x46, 0x63, 0xf1,
	0x9d, 0x1a, 0xf3, 0xc4, 0x65, 0x43, 0xbc, 0xff, 0x3d, 0x07, 0x00, 0x00, 0xff, 0xff, 0x66, 0x57,
	0xf4, 0x2a, 0xc1, 0x04, 0x00, 0x00,
}
