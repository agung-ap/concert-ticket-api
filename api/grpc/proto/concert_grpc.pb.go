// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.20.3
// source: api/grpc/proto/concert.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ConcertService_GetConcert_FullMethodName    = "/concert.ConcertService/GetConcert"
	ConcertService_ListConcerts_FullMethodName  = "/concert.ConcertService/ListConcerts"
	ConcertService_CreateConcert_FullMethodName = "/concert.ConcertService/CreateConcert"
	ConcertService_UpdateConcert_FullMethodName = "/concert.ConcertService/UpdateConcert"
)

// ConcertServiceClient is the client API for ConcertService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ConcertServiceClient interface {
	GetConcert(ctx context.Context, in *GetConcertRequest, opts ...grpc.CallOption) (*Concert, error)
	ListConcerts(ctx context.Context, in *ListConcertsRequest, opts ...grpc.CallOption) (*ListConcertsResponse, error)
	CreateConcert(ctx context.Context, in *CreateConcertRequest, opts ...grpc.CallOption) (*Concert, error)
	UpdateConcert(ctx context.Context, in *UpdateConcertRequest, opts ...grpc.CallOption) (*Concert, error)
}

type concertServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewConcertServiceClient(cc grpc.ClientConnInterface) ConcertServiceClient {
	return &concertServiceClient{cc}
}

func (c *concertServiceClient) GetConcert(ctx context.Context, in *GetConcertRequest, opts ...grpc.CallOption) (*Concert, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Concert)
	err := c.cc.Invoke(ctx, ConcertService_GetConcert_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *concertServiceClient) ListConcerts(ctx context.Context, in *ListConcertsRequest, opts ...grpc.CallOption) (*ListConcertsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListConcertsResponse)
	err := c.cc.Invoke(ctx, ConcertService_ListConcerts_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *concertServiceClient) CreateConcert(ctx context.Context, in *CreateConcertRequest, opts ...grpc.CallOption) (*Concert, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Concert)
	err := c.cc.Invoke(ctx, ConcertService_CreateConcert_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *concertServiceClient) UpdateConcert(ctx context.Context, in *UpdateConcertRequest, opts ...grpc.CallOption) (*Concert, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Concert)
	err := c.cc.Invoke(ctx, ConcertService_UpdateConcert_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ConcertServiceServer is the server API for ConcertService service.
// All implementations must embed UnimplementedConcertServiceServer
// for forward compatibility.
type ConcertServiceServer interface {
	GetConcert(context.Context, *GetConcertRequest) (*Concert, error)
	ListConcerts(context.Context, *ListConcertsRequest) (*ListConcertsResponse, error)
	CreateConcert(context.Context, *CreateConcertRequest) (*Concert, error)
	UpdateConcert(context.Context, *UpdateConcertRequest) (*Concert, error)
	mustEmbedUnimplementedConcertServiceServer()
}

// UnimplementedConcertServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedConcertServiceServer struct{}

func (UnimplementedConcertServiceServer) GetConcert(context.Context, *GetConcertRequest) (*Concert, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConcert not implemented")
}
func (UnimplementedConcertServiceServer) ListConcerts(context.Context, *ListConcertsRequest) (*ListConcertsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListConcerts not implemented")
}
func (UnimplementedConcertServiceServer) CreateConcert(context.Context, *CreateConcertRequest) (*Concert, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateConcert not implemented")
}
func (UnimplementedConcertServiceServer) UpdateConcert(context.Context, *UpdateConcertRequest) (*Concert, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConcert not implemented")
}
func (UnimplementedConcertServiceServer) mustEmbedUnimplementedConcertServiceServer() {}
func (UnimplementedConcertServiceServer) testEmbeddedByValue()                        {}

// UnsafeConcertServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ConcertServiceServer will
// result in compilation errors.
type UnsafeConcertServiceServer interface {
	mustEmbedUnimplementedConcertServiceServer()
}

func RegisterConcertServiceServer(s grpc.ServiceRegistrar, srv ConcertServiceServer) {
	// If the following call pancis, it indicates UnimplementedConcertServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ConcertService_ServiceDesc, srv)
}

func _ConcertService_GetConcert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConcertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConcertServiceServer).GetConcert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConcertService_GetConcert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConcertServiceServer).GetConcert(ctx, req.(*GetConcertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConcertService_ListConcerts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListConcertsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConcertServiceServer).ListConcerts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConcertService_ListConcerts_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConcertServiceServer).ListConcerts(ctx, req.(*ListConcertsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConcertService_CreateConcert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateConcertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConcertServiceServer).CreateConcert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConcertService_CreateConcert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConcertServiceServer).CreateConcert(ctx, req.(*CreateConcertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ConcertService_UpdateConcert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConcertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConcertServiceServer).UpdateConcert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ConcertService_UpdateConcert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConcertServiceServer).UpdateConcert(ctx, req.(*UpdateConcertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ConcertService_ServiceDesc is the grpc.ServiceDesc for ConcertService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ConcertService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "concert.ConcertService",
	HandlerType: (*ConcertServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConcert",
			Handler:    _ConcertService_GetConcert_Handler,
		},
		{
			MethodName: "ListConcerts",
			Handler:    _ConcertService_ListConcerts_Handler,
		},
		{
			MethodName: "CreateConcert",
			Handler:    _ConcertService_CreateConcert_Handler,
		},
		{
			MethodName: "UpdateConcert",
			Handler:    _ConcertService_UpdateConcert_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/grpc/proto/concert.proto",
}
