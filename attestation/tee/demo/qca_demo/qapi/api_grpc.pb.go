// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package qapi

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

// QcaClient is the client API for Qca service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QcaClient interface {
	GetReport(ctx context.Context, in *GetReportRequest, opts ...grpc.CallOption) (*GetReportReply, error)
}

type qcaClient struct {
	cc grpc.ClientConnInterface
}

func NewQcaClient(cc grpc.ClientConnInterface) QcaClient {
	return &qcaClient{cc}
}

func (c *qcaClient) GetReport(ctx context.Context, in *GetReportRequest, opts ...grpc.CallOption) (*GetReportReply, error) {
	out := new(GetReportReply)
	err := c.cc.Invoke(ctx, "/Qca/GetReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QcaServer is the server API for Qca service.
// All implementations must embed UnimplementedQcaServer
// for forward compatibility
type QcaServer interface {
	GetReport(context.Context, *GetReportRequest) (*GetReportReply, error)
	mustEmbedUnimplementedQcaServer()
}

// UnimplementedQcaServer must be embedded to have forward compatible implementations.
type UnimplementedQcaServer struct {
}

func (UnimplementedQcaServer) GetReport(context.Context, *GetReportRequest) (*GetReportReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReport not implemented")
}
func (UnimplementedQcaServer) mustEmbedUnimplementedQcaServer() {}

// UnsafeQcaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QcaServer will
// result in compilation errors.
type UnsafeQcaServer interface {
	mustEmbedUnimplementedQcaServer()
}

func RegisterQcaServer(s grpc.ServiceRegistrar, srv QcaServer) {
	s.RegisterService(&Qca_ServiceDesc, srv)
}

func _Qca_GetReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QcaServer).GetReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Qca/GetReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QcaServer).GetReport(ctx, req.(*GetReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Qca_ServiceDesc is the grpc.ServiceDesc for Qca service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Qca_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Qca",
	HandlerType: (*QcaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetReport",
			Handler:    _Qca_GetReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "qca_demo/qapi/api.proto",
}
