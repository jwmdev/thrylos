// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.2
// source: transactions.proto

package thrylos

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

const (
	BlockchainService_SubmitTransaction_FullMethodName      = "/thrylos.BlockchainService/SubmitTransaction"
	BlockchainService_GetBlock_FullMethodName               = "/thrylos.BlockchainService/GetBlock"
	BlockchainService_GetTransaction_FullMethodName         = "/thrylos.BlockchainService/GetTransaction"
	BlockchainService_GetLastBlock_FullMethodName           = "/thrylos.BlockchainService/GetLastBlock"
	BlockchainService_SubmitTransactionBatch_FullMethodName = "/thrylos.BlockchainService/SubmitTransactionBatch"
	BlockchainService_GetBalance_FullMethodName             = "/thrylos.BlockchainService/GetBalance"
	BlockchainService_GetStats_FullMethodName               = "/thrylos.BlockchainService/GetStats"
	BlockchainService_GetPendingTransactions_FullMethodName = "/thrylos.BlockchainService/GetPendingTransactions"
)

// BlockchainServiceClient is the client API for BlockchainService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlockchainServiceClient interface {
	SubmitTransaction(ctx context.Context, in *TransactionRequest, opts ...grpc.CallOption) (*TransactionResponse, error)
	GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*BlockResponse, error)
	GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*TransactionResponse, error)
	GetLastBlock(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*BlockResponse, error)
	SubmitTransactionBatch(ctx context.Context, in *TransactionBatchRequest, opts ...grpc.CallOption) (*TransactionBatchResponse, error)
	GetBalance(ctx context.Context, in *GetBalanceRequest, opts ...grpc.CallOption) (*BalanceResponse, error)
	GetStats(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (*StatsResponse, error)
	GetPendingTransactions(ctx context.Context, in *GetPendingTransactionsRequest, opts ...grpc.CallOption) (*PendingTransactionsResponse, error)
}

type blockchainServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBlockchainServiceClient(cc grpc.ClientConnInterface) BlockchainServiceClient {
	return &blockchainServiceClient{cc}
}

func (c *blockchainServiceClient) SubmitTransaction(ctx context.Context, in *TransactionRequest, opts ...grpc.CallOption) (*TransactionResponse, error) {
	out := new(TransactionResponse)
	err := c.cc.Invoke(ctx, BlockchainService_SubmitTransaction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*BlockResponse, error) {
	out := new(BlockResponse)
	err := c.cc.Invoke(ctx, BlockchainService_GetBlock_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*TransactionResponse, error) {
	out := new(TransactionResponse)
	err := c.cc.Invoke(ctx, BlockchainService_GetTransaction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetLastBlock(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*BlockResponse, error) {
	out := new(BlockResponse)
	err := c.cc.Invoke(ctx, BlockchainService_GetLastBlock_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) SubmitTransactionBatch(ctx context.Context, in *TransactionBatchRequest, opts ...grpc.CallOption) (*TransactionBatchResponse, error) {
	out := new(TransactionBatchResponse)
	err := c.cc.Invoke(ctx, BlockchainService_SubmitTransactionBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetBalance(ctx context.Context, in *GetBalanceRequest, opts ...grpc.CallOption) (*BalanceResponse, error) {
	out := new(BalanceResponse)
	err := c.cc.Invoke(ctx, BlockchainService_GetBalance_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetStats(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (*StatsResponse, error) {
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, BlockchainService_GetStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetPendingTransactions(ctx context.Context, in *GetPendingTransactionsRequest, opts ...grpc.CallOption) (*PendingTransactionsResponse, error) {
	out := new(PendingTransactionsResponse)
	err := c.cc.Invoke(ctx, BlockchainService_GetPendingTransactions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlockchainServiceServer is the server API for BlockchainService service.
// All implementations must embed UnimplementedBlockchainServiceServer
// for forward compatibility
type BlockchainServiceServer interface {
	SubmitTransaction(context.Context, *TransactionRequest) (*TransactionResponse, error)
	GetBlock(context.Context, *GetBlockRequest) (*BlockResponse, error)
	GetTransaction(context.Context, *GetTransactionRequest) (*TransactionResponse, error)
	GetLastBlock(context.Context, *EmptyRequest) (*BlockResponse, error)
	SubmitTransactionBatch(context.Context, *TransactionBatchRequest) (*TransactionBatchResponse, error)
	GetBalance(context.Context, *GetBalanceRequest) (*BalanceResponse, error)
	GetStats(context.Context, *GetStatsRequest) (*StatsResponse, error)
	GetPendingTransactions(context.Context, *GetPendingTransactionsRequest) (*PendingTransactionsResponse, error)
	mustEmbedUnimplementedBlockchainServiceServer()
}

// UnimplementedBlockchainServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBlockchainServiceServer struct {
}

func (UnimplementedBlockchainServiceServer) SubmitTransaction(context.Context, *TransactionRequest) (*TransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTransaction not implemented")
}
func (UnimplementedBlockchainServiceServer) GetBlock(context.Context, *GetBlockRequest) (*BlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (UnimplementedBlockchainServiceServer) GetTransaction(context.Context, *GetTransactionRequest) (*TransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransaction not implemented")
}
func (UnimplementedBlockchainServiceServer) GetLastBlock(context.Context, *EmptyRequest) (*BlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLastBlock not implemented")
}
func (UnimplementedBlockchainServiceServer) SubmitTransactionBatch(context.Context, *TransactionBatchRequest) (*TransactionBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTransactionBatch not implemented")
}
func (UnimplementedBlockchainServiceServer) GetBalance(context.Context, *GetBalanceRequest) (*BalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (UnimplementedBlockchainServiceServer) GetStats(context.Context, *GetStatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedBlockchainServiceServer) GetPendingTransactions(context.Context, *GetPendingTransactionsRequest) (*PendingTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPendingTransactions not implemented")
}
func (UnimplementedBlockchainServiceServer) mustEmbedUnimplementedBlockchainServiceServer() {}

// UnsafeBlockchainServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlockchainServiceServer will
// result in compilation errors.
type UnsafeBlockchainServiceServer interface {
	mustEmbedUnimplementedBlockchainServiceServer()
}

func RegisterBlockchainServiceServer(s grpc.ServiceRegistrar, srv BlockchainServiceServer) {
	s.RegisterService(&BlockchainService_ServiceDesc, srv)
}

func _BlockchainService_SubmitTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).SubmitTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_SubmitTransaction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).SubmitTransaction(ctx, req.(*TransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_GetBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetBlock(ctx, req.(*GetBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_GetTransaction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetTransaction(ctx, req.(*GetTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetLastBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetLastBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_GetLastBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetLastBlock(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_SubmitTransactionBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).SubmitTransactionBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_SubmitTransactionBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).SubmitTransactionBatch(ctx, req.(*TransactionBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_GetBalance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetBalance(ctx, req.(*GetBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_GetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetStats(ctx, req.(*GetStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetPendingTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPendingTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetPendingTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockchainService_GetPendingTransactions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetPendingTransactions(ctx, req.(*GetPendingTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BlockchainService_ServiceDesc is the grpc.ServiceDesc for BlockchainService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BlockchainService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "thrylos.BlockchainService",
	HandlerType: (*BlockchainServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitTransaction",
			Handler:    _BlockchainService_SubmitTransaction_Handler,
		},
		{
			MethodName: "GetBlock",
			Handler:    _BlockchainService_GetBlock_Handler,
		},
		{
			MethodName: "GetTransaction",
			Handler:    _BlockchainService_GetTransaction_Handler,
		},
		{
			MethodName: "GetLastBlock",
			Handler:    _BlockchainService_GetLastBlock_Handler,
		},
		{
			MethodName: "SubmitTransactionBatch",
			Handler:    _BlockchainService_SubmitTransactionBatch_Handler,
		},
		{
			MethodName: "GetBalance",
			Handler:    _BlockchainService_GetBalance_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _BlockchainService_GetStats_Handler,
		},
		{
			MethodName: "GetPendingTransactions",
			Handler:    _BlockchainService_GetPendingTransactions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "transactions.proto",
}
