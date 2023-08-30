// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.18.1
// source: sp_external.proto

package protos

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

// ExternalClient is the client API for External service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExternalClient interface {
	// Should return everything that is needed to build the initial view in the console
	GetAll(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (*GetAllResponse, error)
	// Temporary method to test gRPC-Web streaming
	GetAllStream(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (External_GetAllStreamClient, error)
	// Returns pipelines (_wasm_bytes field is stripped)
	GetPipelines(ctx context.Context, in *GetPipelinesRequest, opts ...grpc.CallOption) (*GetPipelinesResponse, error)
	// Returns a single pipeline (_wasm_bytes field is stripped)
	GetPipeline(ctx context.Context, in *GetPipelineRequest, opts ...grpc.CallOption) (*GetPipelineResponse, error)
	// Create a new pipeline; id must be left empty on create
	CreatePipeline(ctx context.Context, in *CreatePipelineRequest, opts ...grpc.CallOption) (*CreatePipelineResponse, error)
	// Update an existing pipeline; id must be set
	UpdatePipeline(ctx context.Context, in *UpdatePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Delete a pipeline
	DeletePipeline(ctx context.Context, in *DeletePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Attach a pipeline to an audience
	AttachPipeline(ctx context.Context, in *AttachPipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Detach a pipeline from an audience
	DetachPipeline(ctx context.Context, in *DetachPipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Pause a pipeline; noop if pipeline is already paused
	PausePipeline(ctx context.Context, in *PausePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Resume a pipeline; noop if pipeline is not paused
	ResumePipeline(ctx context.Context, in *ResumePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Create a new notification config
	CreateNotification(ctx context.Context, in *CreateNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Update an existing notification config
	UpdateNotification(ctx context.Context, in *UpdateNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Delete a notification config
	DeleteNotification(ctx context.Context, in *DeleteNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Returns all notification configs
	GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsResponse, error)
	// Returns a single notification config
	GetNotification(ctx context.Context, in *GetNotificationRequest, opts ...grpc.CallOption) (*GetNotificationResponse, error)
	// Attach a notification config to a pipeline
	AttachNotification(ctx context.Context, in *AttachNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Detach a notification config from a pipeline
	DetachNotification(ctx context.Context, in *DetachNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Delete an un-attached audience
	DeleteAudience(ctx context.Context, in *DeleteAudienceRequest, opts ...grpc.CallOption) (*StandardResponse, error)
	// Returns all metric counters
	GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (External_GetMetricsClient, error)
	// Test method
	Test(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error)
}

type externalClient struct {
	cc grpc.ClientConnInterface
}

func NewExternalClient(cc grpc.ClientConnInterface) ExternalClient {
	return &externalClient{cc}
}

func (c *externalClient) GetAll(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (*GetAllResponse, error) {
	out := new(GetAllResponse)
	err := c.cc.Invoke(ctx, "/protos.External/GetAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) GetAllStream(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (External_GetAllStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &External_ServiceDesc.Streams[0], "/protos.External/GetAllStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &externalGetAllStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type External_GetAllStreamClient interface {
	Recv() (*GetAllResponse, error)
	grpc.ClientStream
}

type externalGetAllStreamClient struct {
	grpc.ClientStream
}

func (x *externalGetAllStreamClient) Recv() (*GetAllResponse, error) {
	m := new(GetAllResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *externalClient) GetPipelines(ctx context.Context, in *GetPipelinesRequest, opts ...grpc.CallOption) (*GetPipelinesResponse, error) {
	out := new(GetPipelinesResponse)
	err := c.cc.Invoke(ctx, "/protos.External/GetPipelines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) GetPipeline(ctx context.Context, in *GetPipelineRequest, opts ...grpc.CallOption) (*GetPipelineResponse, error) {
	out := new(GetPipelineResponse)
	err := c.cc.Invoke(ctx, "/protos.External/GetPipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) CreatePipeline(ctx context.Context, in *CreatePipelineRequest, opts ...grpc.CallOption) (*CreatePipelineResponse, error) {
	out := new(CreatePipelineResponse)
	err := c.cc.Invoke(ctx, "/protos.External/CreatePipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) UpdatePipeline(ctx context.Context, in *UpdatePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/UpdatePipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) DeletePipeline(ctx context.Context, in *DeletePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/DeletePipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) AttachPipeline(ctx context.Context, in *AttachPipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/AttachPipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) DetachPipeline(ctx context.Context, in *DetachPipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/DetachPipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) PausePipeline(ctx context.Context, in *PausePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/PausePipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) ResumePipeline(ctx context.Context, in *ResumePipelineRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/ResumePipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) CreateNotification(ctx context.Context, in *CreateNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/CreateNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) UpdateNotification(ctx context.Context, in *UpdateNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/UpdateNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) DeleteNotification(ctx context.Context, in *DeleteNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/DeleteNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsResponse, error) {
	out := new(GetNotificationsResponse)
	err := c.cc.Invoke(ctx, "/protos.External/GetNotifications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) GetNotification(ctx context.Context, in *GetNotificationRequest, opts ...grpc.CallOption) (*GetNotificationResponse, error) {
	out := new(GetNotificationResponse)
	err := c.cc.Invoke(ctx, "/protos.External/GetNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) AttachNotification(ctx context.Context, in *AttachNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/AttachNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) DetachNotification(ctx context.Context, in *DetachNotificationRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/DetachNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) DeleteAudience(ctx context.Context, in *DeleteAudienceRequest, opts ...grpc.CallOption) (*StandardResponse, error) {
	out := new(StandardResponse)
	err := c.cc.Invoke(ctx, "/protos.External/DeleteAudience", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *externalClient) GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (External_GetMetricsClient, error) {
	stream, err := c.cc.NewStream(ctx, &External_ServiceDesc.Streams[1], "/protos.External/GetMetrics", opts...)
	if err != nil {
		return nil, err
	}
	x := &externalGetMetricsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type External_GetMetricsClient interface {
	Recv() (*GetMetricsResponse, error)
	grpc.ClientStream
}

type externalGetMetricsClient struct {
	grpc.ClientStream
}

func (x *externalGetMetricsClient) Recv() (*GetMetricsResponse, error) {
	m := new(GetMetricsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *externalClient) Test(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error) {
	out := new(TestResponse)
	err := c.cc.Invoke(ctx, "/protos.External/Test", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExternalServer is the server API for External service.
// All implementations must embed UnimplementedExternalServer
// for forward compatibility
type ExternalServer interface {
	// Should return everything that is needed to build the initial view in the console
	GetAll(context.Context, *GetAllRequest) (*GetAllResponse, error)
	// Temporary method to test gRPC-Web streaming
	GetAllStream(*GetAllRequest, External_GetAllStreamServer) error
	// Returns pipelines (_wasm_bytes field is stripped)
	GetPipelines(context.Context, *GetPipelinesRequest) (*GetPipelinesResponse, error)
	// Returns a single pipeline (_wasm_bytes field is stripped)
	GetPipeline(context.Context, *GetPipelineRequest) (*GetPipelineResponse, error)
	// Create a new pipeline; id must be left empty on create
	CreatePipeline(context.Context, *CreatePipelineRequest) (*CreatePipelineResponse, error)
	// Update an existing pipeline; id must be set
	UpdatePipeline(context.Context, *UpdatePipelineRequest) (*StandardResponse, error)
	// Delete a pipeline
	DeletePipeline(context.Context, *DeletePipelineRequest) (*StandardResponse, error)
	// Attach a pipeline to an audience
	AttachPipeline(context.Context, *AttachPipelineRequest) (*StandardResponse, error)
	// Detach a pipeline from an audience
	DetachPipeline(context.Context, *DetachPipelineRequest) (*StandardResponse, error)
	// Pause a pipeline; noop if pipeline is already paused
	PausePipeline(context.Context, *PausePipelineRequest) (*StandardResponse, error)
	// Resume a pipeline; noop if pipeline is not paused
	ResumePipeline(context.Context, *ResumePipelineRequest) (*StandardResponse, error)
	// Create a new notification config
	CreateNotification(context.Context, *CreateNotificationRequest) (*StandardResponse, error)
	// Update an existing notification config
	UpdateNotification(context.Context, *UpdateNotificationRequest) (*StandardResponse, error)
	// Delete a notification config
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*StandardResponse, error)
	// Returns all notification configs
	GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsResponse, error)
	// Returns a single notification config
	GetNotification(context.Context, *GetNotificationRequest) (*GetNotificationResponse, error)
	// Attach a notification config to a pipeline
	AttachNotification(context.Context, *AttachNotificationRequest) (*StandardResponse, error)
	// Detach a notification config from a pipeline
	DetachNotification(context.Context, *DetachNotificationRequest) (*StandardResponse, error)
	// Delete an un-attached audience
	DeleteAudience(context.Context, *DeleteAudienceRequest) (*StandardResponse, error)
	// Returns all metric counters
	GetMetrics(*GetMetricsRequest, External_GetMetricsServer) error
	// Test method
	Test(context.Context, *TestRequest) (*TestResponse, error)
	mustEmbedUnimplementedExternalServer()
}

// UnimplementedExternalServer must be embedded to have forward compatible implementations.
type UnimplementedExternalServer struct {
}

func (UnimplementedExternalServer) GetAll(context.Context, *GetAllRequest) (*GetAllResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAll not implemented")
}
func (UnimplementedExternalServer) GetAllStream(*GetAllRequest, External_GetAllStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method GetAllStream not implemented")
}
func (UnimplementedExternalServer) GetPipelines(context.Context, *GetPipelinesRequest) (*GetPipelinesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPipelines not implemented")
}
func (UnimplementedExternalServer) GetPipeline(context.Context, *GetPipelineRequest) (*GetPipelineResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPipeline not implemented")
}
func (UnimplementedExternalServer) CreatePipeline(context.Context, *CreatePipelineRequest) (*CreatePipelineResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePipeline not implemented")
}
func (UnimplementedExternalServer) UpdatePipeline(context.Context, *UpdatePipelineRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePipeline not implemented")
}
func (UnimplementedExternalServer) DeletePipeline(context.Context, *DeletePipelineRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePipeline not implemented")
}
func (UnimplementedExternalServer) AttachPipeline(context.Context, *AttachPipelineRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AttachPipeline not implemented")
}
func (UnimplementedExternalServer) DetachPipeline(context.Context, *DetachPipelineRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DetachPipeline not implemented")
}
func (UnimplementedExternalServer) PausePipeline(context.Context, *PausePipelineRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PausePipeline not implemented")
}
func (UnimplementedExternalServer) ResumePipeline(context.Context, *ResumePipelineRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResumePipeline not implemented")
}
func (UnimplementedExternalServer) CreateNotification(context.Context, *CreateNotificationRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateNotification not implemented")
}
func (UnimplementedExternalServer) UpdateNotification(context.Context, *UpdateNotificationRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNotification not implemented")
}
func (UnimplementedExternalServer) DeleteNotification(context.Context, *DeleteNotificationRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNotification not implemented")
}
func (UnimplementedExternalServer) GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNotifications not implemented")
}
func (UnimplementedExternalServer) GetNotification(context.Context, *GetNotificationRequest) (*GetNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNotification not implemented")
}
func (UnimplementedExternalServer) AttachNotification(context.Context, *AttachNotificationRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AttachNotification not implemented")
}
func (UnimplementedExternalServer) DetachNotification(context.Context, *DetachNotificationRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DetachNotification not implemented")
}
func (UnimplementedExternalServer) DeleteAudience(context.Context, *DeleteAudienceRequest) (*StandardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAudience not implemented")
}
func (UnimplementedExternalServer) GetMetrics(*GetMetricsRequest, External_GetMetricsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetMetrics not implemented")
}
func (UnimplementedExternalServer) Test(context.Context, *TestRequest) (*TestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Test not implemented")
}
func (UnimplementedExternalServer) mustEmbedUnimplementedExternalServer() {}

// UnsafeExternalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExternalServer will
// result in compilation errors.
type UnsafeExternalServer interface {
	mustEmbedUnimplementedExternalServer()
}

func RegisterExternalServer(s grpc.ServiceRegistrar, srv ExternalServer) {
	s.RegisterService(&External_ServiceDesc, srv)
}

func _External_GetAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).GetAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/GetAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).GetAll(ctx, req.(*GetAllRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_GetAllStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetAllRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExternalServer).GetAllStream(m, &externalGetAllStreamServer{stream})
}

type External_GetAllStreamServer interface {
	Send(*GetAllResponse) error
	grpc.ServerStream
}

type externalGetAllStreamServer struct {
	grpc.ServerStream
}

func (x *externalGetAllStreamServer) Send(m *GetAllResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _External_GetPipelines_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPipelinesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).GetPipelines(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/GetPipelines",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).GetPipelines(ctx, req.(*GetPipelinesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_GetPipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).GetPipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/GetPipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).GetPipeline(ctx, req.(*GetPipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_CreatePipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).CreatePipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/CreatePipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).CreatePipeline(ctx, req.(*CreatePipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_UpdatePipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).UpdatePipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/UpdatePipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).UpdatePipeline(ctx, req.(*UpdatePipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_DeletePipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletePipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).DeletePipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/DeletePipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).DeletePipeline(ctx, req.(*DeletePipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_AttachPipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AttachPipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).AttachPipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/AttachPipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).AttachPipeline(ctx, req.(*AttachPipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_DetachPipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DetachPipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).DetachPipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/DetachPipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).DetachPipeline(ctx, req.(*DetachPipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_PausePipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PausePipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).PausePipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/PausePipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).PausePipeline(ctx, req.(*PausePipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_ResumePipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResumePipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).ResumePipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/ResumePipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).ResumePipeline(ctx, req.(*ResumePipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_CreateNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).CreateNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/CreateNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).CreateNotification(ctx, req.(*CreateNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_UpdateNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).UpdateNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/UpdateNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).UpdateNotification(ctx, req.(*UpdateNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_DeleteNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).DeleteNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/DeleteNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).DeleteNotification(ctx, req.(*DeleteNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_GetNotifications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNotificationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).GetNotifications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/GetNotifications",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).GetNotifications(ctx, req.(*GetNotificationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_GetNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).GetNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/GetNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).GetNotification(ctx, req.(*GetNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_AttachNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AttachNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).AttachNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/AttachNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).AttachNotification(ctx, req.(*AttachNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_DetachNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DetachNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).DetachNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/DetachNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).DetachNotification(ctx, req.(*DetachNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_DeleteAudience_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAudienceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).DeleteAudience(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/DeleteAudience",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).DeleteAudience(ctx, req.(*DeleteAudienceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _External_GetMetrics_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetMetricsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExternalServer).GetMetrics(m, &externalGetMetricsServer{stream})
}

type External_GetMetricsServer interface {
	Send(*GetMetricsResponse) error
	grpc.ServerStream
}

type externalGetMetricsServer struct {
	grpc.ServerStream
}

func (x *externalGetMetricsServer) Send(m *GetMetricsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _External_Test_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExternalServer).Test(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.External/Test",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExternalServer).Test(ctx, req.(*TestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// External_ServiceDesc is the grpc.ServiceDesc for External service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var External_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protos.External",
	HandlerType: (*ExternalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAll",
			Handler:    _External_GetAll_Handler,
		},
		{
			MethodName: "GetPipelines",
			Handler:    _External_GetPipelines_Handler,
		},
		{
			MethodName: "GetPipeline",
			Handler:    _External_GetPipeline_Handler,
		},
		{
			MethodName: "CreatePipeline",
			Handler:    _External_CreatePipeline_Handler,
		},
		{
			MethodName: "UpdatePipeline",
			Handler:    _External_UpdatePipeline_Handler,
		},
		{
			MethodName: "DeletePipeline",
			Handler:    _External_DeletePipeline_Handler,
		},
		{
			MethodName: "AttachPipeline",
			Handler:    _External_AttachPipeline_Handler,
		},
		{
			MethodName: "DetachPipeline",
			Handler:    _External_DetachPipeline_Handler,
		},
		{
			MethodName: "PausePipeline",
			Handler:    _External_PausePipeline_Handler,
		},
		{
			MethodName: "ResumePipeline",
			Handler:    _External_ResumePipeline_Handler,
		},
		{
			MethodName: "CreateNotification",
			Handler:    _External_CreateNotification_Handler,
		},
		{
			MethodName: "UpdateNotification",
			Handler:    _External_UpdateNotification_Handler,
		},
		{
			MethodName: "DeleteNotification",
			Handler:    _External_DeleteNotification_Handler,
		},
		{
			MethodName: "GetNotifications",
			Handler:    _External_GetNotifications_Handler,
		},
		{
			MethodName: "GetNotification",
			Handler:    _External_GetNotification_Handler,
		},
		{
			MethodName: "AttachNotification",
			Handler:    _External_AttachNotification_Handler,
		},
		{
			MethodName: "DetachNotification",
			Handler:    _External_DetachNotification_Handler,
		},
		{
			MethodName: "DeleteAudience",
			Handler:    _External_DeleteAudience_Handler,
		},
		{
			MethodName: "Test",
			Handler:    _External_Test_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetAllStream",
			Handler:       _External_GetAllStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetMetrics",
			Handler:       _External_GetMetrics_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "sp_external.proto",
}