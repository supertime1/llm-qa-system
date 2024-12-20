// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.1
// source: medical_service.proto

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
	MedicalQAService_GenerateDraftAnswer_FullMethodName = "/backend.MedicalQAService/GenerateDraftAnswer"
)

// MedicalQAServiceClient is the client API for MedicalQAService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Main QA service definition
type MedicalQAServiceClient interface {
	// Generate a draft answer for medical questions
	GenerateDraftAnswer(ctx context.Context, in *QuestionRequest, opts ...grpc.CallOption) (*QuestionResponse, error)
}

type medicalQAServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMedicalQAServiceClient(cc grpc.ClientConnInterface) MedicalQAServiceClient {
	return &medicalQAServiceClient{cc}
}

func (c *medicalQAServiceClient) GenerateDraftAnswer(ctx context.Context, in *QuestionRequest, opts ...grpc.CallOption) (*QuestionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QuestionResponse)
	err := c.cc.Invoke(ctx, MedicalQAService_GenerateDraftAnswer_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MedicalQAServiceServer is the server API for MedicalQAService service.
// All implementations must embed UnimplementedMedicalQAServiceServer
// for forward compatibility.
//
// Main QA service definition
type MedicalQAServiceServer interface {
	// Generate a draft answer for medical questions
	GenerateDraftAnswer(context.Context, *QuestionRequest) (*QuestionResponse, error)
	mustEmbedUnimplementedMedicalQAServiceServer()
}

// UnimplementedMedicalQAServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMedicalQAServiceServer struct{}

func (UnimplementedMedicalQAServiceServer) GenerateDraftAnswer(context.Context, *QuestionRequest) (*QuestionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateDraftAnswer not implemented")
}
func (UnimplementedMedicalQAServiceServer) mustEmbedUnimplementedMedicalQAServiceServer() {}
func (UnimplementedMedicalQAServiceServer) testEmbeddedByValue()                          {}

// UnsafeMedicalQAServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MedicalQAServiceServer will
// result in compilation errors.
type UnsafeMedicalQAServiceServer interface {
	mustEmbedUnimplementedMedicalQAServiceServer()
}

func RegisterMedicalQAServiceServer(s grpc.ServiceRegistrar, srv MedicalQAServiceServer) {
	// If the following call pancis, it indicates UnimplementedMedicalQAServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MedicalQAService_ServiceDesc, srv)
}

func _MedicalQAService_GenerateDraftAnswer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuestionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MedicalQAServiceServer).GenerateDraftAnswer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MedicalQAService_GenerateDraftAnswer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MedicalQAServiceServer).GenerateDraftAnswer(ctx, req.(*QuestionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MedicalQAService_ServiceDesc is the grpc.ServiceDesc for MedicalQAService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MedicalQAService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "backend.MedicalQAService",
	HandlerType: (*MedicalQAServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenerateDraftAnswer",
			Handler:    _MedicalQAService_GenerateDraftAnswer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "medical_service.proto",
}

const (
	MedicalService_SubmitQuestion_FullMethodName    = "/backend.MedicalService/SubmitQuestion"
	MedicalService_GetQuestionStatus_FullMethodName = "/backend.MedicalService/GetQuestionStatus"
	MedicalService_GetAnswerHistory_FullMethodName  = "/backend.MedicalService/GetAnswerHistory"
	MedicalService_GetPendingReviews_FullMethodName = "/backend.MedicalService/GetPendingReviews"
	MedicalService_SubmitReview_FullMethodName      = "/backend.MedicalService/SubmitReview"
)

// MedicalServiceClient is the client API for MedicalService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Backend Medical Service
type MedicalServiceClient interface {
	// Patient operations
	SubmitQuestion(ctx context.Context, in *PatientQuestionRequest, opts ...grpc.CallOption) (*PatientQuestionResponse, error)
	GetQuestionStatus(ctx context.Context, in *QuestionStatusRequest, opts ...grpc.CallOption) (*QuestionStatusResponse, error)
	GetAnswerHistory(ctx context.Context, in *AnswerHistoryRequest, opts ...grpc.CallOption) (*AnswerHistoryResponse, error)
	// Doctor operations
	GetPendingReviews(ctx context.Context, in *PendingReviewsRequest, opts ...grpc.CallOption) (*PendingReviewsResponse, error)
	SubmitReview(ctx context.Context, in *ReviewSubmissionRequest, opts ...grpc.CallOption) (*ReviewSubmissionResponse, error)
}

type medicalServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMedicalServiceClient(cc grpc.ClientConnInterface) MedicalServiceClient {
	return &medicalServiceClient{cc}
}

func (c *medicalServiceClient) SubmitQuestion(ctx context.Context, in *PatientQuestionRequest, opts ...grpc.CallOption) (*PatientQuestionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PatientQuestionResponse)
	err := c.cc.Invoke(ctx, MedicalService_SubmitQuestion_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *medicalServiceClient) GetQuestionStatus(ctx context.Context, in *QuestionStatusRequest, opts ...grpc.CallOption) (*QuestionStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QuestionStatusResponse)
	err := c.cc.Invoke(ctx, MedicalService_GetQuestionStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *medicalServiceClient) GetAnswerHistory(ctx context.Context, in *AnswerHistoryRequest, opts ...grpc.CallOption) (*AnswerHistoryResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AnswerHistoryResponse)
	err := c.cc.Invoke(ctx, MedicalService_GetAnswerHistory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *medicalServiceClient) GetPendingReviews(ctx context.Context, in *PendingReviewsRequest, opts ...grpc.CallOption) (*PendingReviewsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PendingReviewsResponse)
	err := c.cc.Invoke(ctx, MedicalService_GetPendingReviews_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *medicalServiceClient) SubmitReview(ctx context.Context, in *ReviewSubmissionRequest, opts ...grpc.CallOption) (*ReviewSubmissionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReviewSubmissionResponse)
	err := c.cc.Invoke(ctx, MedicalService_SubmitReview_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MedicalServiceServer is the server API for MedicalService service.
// All implementations must embed UnimplementedMedicalServiceServer
// for forward compatibility.
//
// Backend Medical Service
type MedicalServiceServer interface {
	// Patient operations
	SubmitQuestion(context.Context, *PatientQuestionRequest) (*PatientQuestionResponse, error)
	GetQuestionStatus(context.Context, *QuestionStatusRequest) (*QuestionStatusResponse, error)
	GetAnswerHistory(context.Context, *AnswerHistoryRequest) (*AnswerHistoryResponse, error)
	// Doctor operations
	GetPendingReviews(context.Context, *PendingReviewsRequest) (*PendingReviewsResponse, error)
	SubmitReview(context.Context, *ReviewSubmissionRequest) (*ReviewSubmissionResponse, error)
	mustEmbedUnimplementedMedicalServiceServer()
}

// UnimplementedMedicalServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMedicalServiceServer struct{}

func (UnimplementedMedicalServiceServer) SubmitQuestion(context.Context, *PatientQuestionRequest) (*PatientQuestionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitQuestion not implemented")
}
func (UnimplementedMedicalServiceServer) GetQuestionStatus(context.Context, *QuestionStatusRequest) (*QuestionStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQuestionStatus not implemented")
}
func (UnimplementedMedicalServiceServer) GetAnswerHistory(context.Context, *AnswerHistoryRequest) (*AnswerHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAnswerHistory not implemented")
}
func (UnimplementedMedicalServiceServer) GetPendingReviews(context.Context, *PendingReviewsRequest) (*PendingReviewsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPendingReviews not implemented")
}
func (UnimplementedMedicalServiceServer) SubmitReview(context.Context, *ReviewSubmissionRequest) (*ReviewSubmissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitReview not implemented")
}
func (UnimplementedMedicalServiceServer) mustEmbedUnimplementedMedicalServiceServer() {}
func (UnimplementedMedicalServiceServer) testEmbeddedByValue()                        {}

// UnsafeMedicalServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MedicalServiceServer will
// result in compilation errors.
type UnsafeMedicalServiceServer interface {
	mustEmbedUnimplementedMedicalServiceServer()
}

func RegisterMedicalServiceServer(s grpc.ServiceRegistrar, srv MedicalServiceServer) {
	// If the following call pancis, it indicates UnimplementedMedicalServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MedicalService_ServiceDesc, srv)
}

func _MedicalService_SubmitQuestion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatientQuestionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MedicalServiceServer).SubmitQuestion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MedicalService_SubmitQuestion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MedicalServiceServer).SubmitQuestion(ctx, req.(*PatientQuestionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MedicalService_GetQuestionStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuestionStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MedicalServiceServer).GetQuestionStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MedicalService_GetQuestionStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MedicalServiceServer).GetQuestionStatus(ctx, req.(*QuestionStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MedicalService_GetAnswerHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnswerHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MedicalServiceServer).GetAnswerHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MedicalService_GetAnswerHistory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MedicalServiceServer).GetAnswerHistory(ctx, req.(*AnswerHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MedicalService_GetPendingReviews_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PendingReviewsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MedicalServiceServer).GetPendingReviews(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MedicalService_GetPendingReviews_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MedicalServiceServer).GetPendingReviews(ctx, req.(*PendingReviewsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MedicalService_SubmitReview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReviewSubmissionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MedicalServiceServer).SubmitReview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MedicalService_SubmitReview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MedicalServiceServer).SubmitReview(ctx, req.(*ReviewSubmissionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MedicalService_ServiceDesc is the grpc.ServiceDesc for MedicalService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MedicalService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "backend.MedicalService",
	HandlerType: (*MedicalServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitQuestion",
			Handler:    _MedicalService_SubmitQuestion_Handler,
		},
		{
			MethodName: "GetQuestionStatus",
			Handler:    _MedicalService_GetQuestionStatus_Handler,
		},
		{
			MethodName: "GetAnswerHistory",
			Handler:    _MedicalService_GetAnswerHistory_Handler,
		},
		{
			MethodName: "GetPendingReviews",
			Handler:    _MedicalService_GetPendingReviews_Handler,
		},
		{
			MethodName: "SubmitReview",
			Handler:    _MedicalService_SubmitReview_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "medical_service.proto",
}
