// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: shep_remote_muster/shep_remote_muster.proto

package shep_remote_muster

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PulseClient is the client API for Pulse service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PulseClient interface {
	HeartBeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatReply, error)
}

type pulseClient struct {
	cc grpc.ClientConnInterface
}

func NewPulseClient(cc grpc.ClientConnInterface) PulseClient {
	return &pulseClient{cc}
}

func (c *pulseClient) HeartBeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatReply, error) {
	out := new(HeartbeatReply)
	err := c.cc.Invoke(ctx, "/muster.Pulse/HeartBeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PulseServer is the server API for Pulse service.
// All implementations must embed UnimplementedPulseServer
// for forward compatibility
type PulseServer interface {
	HeartBeat(context.Context, *HeartbeatRequest) (*HeartbeatReply, error)
	mustEmbedUnimplementedPulseServer()
}

// UnimplementedPulseServer must be embedded to have forward compatible implementations.
type UnimplementedPulseServer struct {
}

func (UnimplementedPulseServer) HeartBeat(context.Context, *HeartbeatRequest) (*HeartbeatReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HeartBeat not implemented")
}
func (UnimplementedPulseServer) mustEmbedUnimplementedPulseServer() {}

// UnsafePulseServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PulseServer will
// result in compilation errors.
type UnsafePulseServer interface {
	mustEmbedUnimplementedPulseServer()
}

func RegisterPulseServer(s grpc.ServiceRegistrar, srv PulseServer) {
	s.RegisterService(&Pulse_ServiceDesc, srv)
}

func _Pulse_HeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PulseServer).HeartBeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/muster.Pulse/HeartBeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PulseServer).HeartBeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Pulse_ServiceDesc is the grpc.ServiceDesc for Pulse service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Pulse_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "muster.Pulse",
	HandlerType: (*PulseServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HeartBeat",
			Handler:    _Pulse_HeartBeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shep_remote_muster/shep_remote_muster.proto",
}

// LogClient is the client API for Log service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogClient interface {
	SyncLogBuffers(ctx context.Context, opts ...grpc.CallOption) (Log_SyncLogBuffersClient, error)
}

type logClient struct {
	cc grpc.ClientConnInterface
}

func NewLogClient(cc grpc.ClientConnInterface) LogClient {
	return &logClient{cc}
}

func (c *logClient) SyncLogBuffers(ctx context.Context, opts ...grpc.CallOption) (Log_SyncLogBuffersClient, error) {
	stream, err := c.cc.NewStream(ctx, &Log_ServiceDesc.Streams[0], "/muster.Log/SyncLogBuffers", opts...)
	if err != nil {
		return nil, err
	}
	x := &logSyncLogBuffersClient{stream}
	return x, nil
}

type Log_SyncLogBuffersClient interface {
	Send(*SyncLogRequest) error
	CloseAndRecv() (*SyncLogReply, error)
	grpc.ClientStream
}

type logSyncLogBuffersClient struct {
	grpc.ClientStream
}

func (x *logSyncLogBuffersClient) Send(m *SyncLogRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *logSyncLogBuffersClient) CloseAndRecv() (*SyncLogReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(SyncLogReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// LogServer is the server API for Log service.
// All implementations must embed UnimplementedLogServer
// for forward compatibility
type LogServer interface {
	SyncLogBuffers(Log_SyncLogBuffersServer) error
	mustEmbedUnimplementedLogServer()
}

// UnimplementedLogServer must be embedded to have forward compatible implementations.
type UnimplementedLogServer struct {
}

func (UnimplementedLogServer) SyncLogBuffers(Log_SyncLogBuffersServer) error {
	return status.Errorf(codes.Unimplemented, "method SyncLogBuffers not implemented")
}
func (UnimplementedLogServer) mustEmbedUnimplementedLogServer() {}

// UnsafeLogServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogServer will
// result in compilation errors.
type UnsafeLogServer interface {
	mustEmbedUnimplementedLogServer()
}

func RegisterLogServer(s grpc.ServiceRegistrar, srv LogServer) {
	s.RegisterService(&Log_ServiceDesc, srv)
}

func _Log_SyncLogBuffers_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LogServer).SyncLogBuffers(&logSyncLogBuffersServer{stream})
}

type Log_SyncLogBuffersServer interface {
	SendAndClose(*SyncLogReply) error
	Recv() (*SyncLogRequest, error)
	grpc.ServerStream
}

type logSyncLogBuffersServer struct {
	grpc.ServerStream
}

func (x *logSyncLogBuffersServer) SendAndClose(m *SyncLogReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *logSyncLogBuffersServer) Recv() (*SyncLogRequest, error) {
	m := new(SyncLogRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Log_ServiceDesc is the grpc.ServiceDesc for Log service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Log_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "muster.Log",
	HandlerType: (*LogServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SyncLogBuffers",
			Handler:       _Log_SyncLogBuffers_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "shep_remote_muster/shep_remote_muster.proto",
}

// ControlClient is the client API for Control service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ControlClient interface {
	ApplyControl(ctx context.Context, opts ...grpc.CallOption) (Control_ApplyControlClient, error)
}

type controlClient struct {
	cc grpc.ClientConnInterface
}

func NewControlClient(cc grpc.ClientConnInterface) ControlClient {
	return &controlClient{cc}
}

func (c *controlClient) ApplyControl(ctx context.Context, opts ...grpc.CallOption) (Control_ApplyControlClient, error) {
	stream, err := c.cc.NewStream(ctx, &Control_ServiceDesc.Streams[0], "/muster.Control/ApplyControl", opts...)
	if err != nil {
		return nil, err
	}
	x := &controlApplyControlClient{stream}
	return x, nil
}

type Control_ApplyControlClient interface {
	Send(*ControlRequest) error
	CloseAndRecv() (*ControlReply, error)
	grpc.ClientStream
}

type controlApplyControlClient struct {
	grpc.ClientStream
}

func (x *controlApplyControlClient) Send(m *ControlRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *controlApplyControlClient) CloseAndRecv() (*ControlReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ControlReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ControlServer is the server API for Control service.
// All implementations must embed UnimplementedControlServer
// for forward compatibility
type ControlServer interface {
	ApplyControl(Control_ApplyControlServer) error
	mustEmbedUnimplementedControlServer()
}

// UnimplementedControlServer must be embedded to have forward compatible implementations.
type UnimplementedControlServer struct {
}

func (UnimplementedControlServer) ApplyControl(Control_ApplyControlServer) error {
	return status.Errorf(codes.Unimplemented, "method ApplyControl not implemented")
}
func (UnimplementedControlServer) mustEmbedUnimplementedControlServer() {}

// UnsafeControlServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ControlServer will
// result in compilation errors.
type UnsafeControlServer interface {
	mustEmbedUnimplementedControlServer()
}

func RegisterControlServer(s grpc.ServiceRegistrar, srv ControlServer) {
	s.RegisterService(&Control_ServiceDesc, srv)
}

func _Control_ApplyControl_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ControlServer).ApplyControl(&controlApplyControlServer{stream})
}

type Control_ApplyControlServer interface {
	SendAndClose(*ControlReply) error
	Recv() (*ControlRequest, error)
	grpc.ServerStream
}

type controlApplyControlServer struct {
	grpc.ServerStream
}

func (x *controlApplyControlServer) SendAndClose(m *ControlReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *controlApplyControlServer) Recv() (*ControlRequest, error) {
	m := new(ControlRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Control_ServiceDesc is the grpc.ServiceDesc for Control service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Control_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "muster.Control",
	HandlerType: (*ControlServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ApplyControl",
			Handler:       _Control_ApplyControl_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "shep_remote_muster/shep_remote_muster.proto",
}

// CoordinateClient is the client API for Coordinate service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CoordinateClient interface {
	CoordinateLog(ctx context.Context, in *CoordinateLogRequest, opts ...grpc.CallOption) (*CoordinateLogReply, error)
	CompleteRun(ctx context.Context, in *CompleteRunRequest, opts ...grpc.CallOption) (*CompleteRunReply, error)
}

type coordinateClient struct {
	cc grpc.ClientConnInterface
}

func NewCoordinateClient(cc grpc.ClientConnInterface) CoordinateClient {
	return &coordinateClient{cc}
}

func (c *coordinateClient) CoordinateLog(ctx context.Context, in *CoordinateLogRequest, opts ...grpc.CallOption) (*CoordinateLogReply, error) {
	out := new(CoordinateLogReply)
	err := c.cc.Invoke(ctx, "/muster.Coordinate/CoordinateLog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coordinateClient) CompleteRun(ctx context.Context, in *CompleteRunRequest, opts ...grpc.CallOption) (*CompleteRunReply, error) {
	out := new(CompleteRunReply)
	err := c.cc.Invoke(ctx, "/muster.Coordinate/CompleteRun", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CoordinateServer is the server API for Coordinate service.
// All implementations must embed UnimplementedCoordinateServer
// for forward compatibility
type CoordinateServer interface {
	CoordinateLog(context.Context, *CoordinateLogRequest) (*CoordinateLogReply, error)
	CompleteRun(context.Context, *CompleteRunRequest) (*CompleteRunReply, error)
	mustEmbedUnimplementedCoordinateServer()
}

// UnimplementedCoordinateServer must be embedded to have forward compatible implementations.
type UnimplementedCoordinateServer struct {
}

func (UnimplementedCoordinateServer) CoordinateLog(context.Context, *CoordinateLogRequest) (*CoordinateLogReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CoordinateLog not implemented")
}
func (UnimplementedCoordinateServer) CompleteRun(context.Context, *CompleteRunRequest) (*CompleteRunReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CompleteRun not implemented")
}
func (UnimplementedCoordinateServer) mustEmbedUnimplementedCoordinateServer() {}

// UnsafeCoordinateServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CoordinateServer will
// result in compilation errors.
type UnsafeCoordinateServer interface {
	mustEmbedUnimplementedCoordinateServer()
}

func RegisterCoordinateServer(s grpc.ServiceRegistrar, srv CoordinateServer) {
	s.RegisterService(&Coordinate_ServiceDesc, srv)
}

func _Coordinate_CoordinateLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CoordinateLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordinateServer).CoordinateLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/muster.Coordinate/CoordinateLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordinateServer).CoordinateLog(ctx, req.(*CoordinateLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coordinate_CompleteRun_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompleteRunRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoordinateServer).CompleteRun(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/muster.Coordinate/CompleteRun",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoordinateServer).CompleteRun(ctx, req.(*CompleteRunRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Coordinate_ServiceDesc is the grpc.ServiceDesc for Coordinate service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Coordinate_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "muster.Coordinate",
	HandlerType: (*CoordinateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CoordinateLog",
			Handler:    _Coordinate_CoordinateLog_Handler,
		},
		{
			MethodName: "CompleteRun",
			Handler:    _Coordinate_CompleteRun_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shep_remote_muster/shep_remote_muster.proto",
}
