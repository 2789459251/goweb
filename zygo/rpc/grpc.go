package rpc

import (
	"google.golang.org/grpc"
	"net"
)

type MyGrpcServer struct {
	listen     net.Listener
	grpcServer *grpc.Server
	registers  []func(grpcServer *grpc.Server)
	ops        []grpc.ServerOption
}

func NewGrpcServer(address string, ops ...MsGrpcOption) (*MyGrpcServer, error) {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	ms := &MyGrpcServer{
		listen: listen,
	}
	for _, op := range ops {
		op.Apply(ms)
	}
	s := grpc.NewServer(ms.ops...)
	ms.grpcServer = s
	return ms, nil
}

func (s *MyGrpcServer) Run() error {
	for _, register := range s.registers {
		register(s.grpcServer)
	}
	return s.grpcServer.Serve(s.listen)
}

func (s *MyGrpcServer) Register(register func(grpServer *grpc.Server)) {
	s.registers = append(s.registers, register)
}

type MsGrpcOption interface {
	Apply(s *MyGrpcServer)
}

type DefaultGrpcOption struct {
	f func(s *MyGrpcServer)
}

func (d DefaultGrpcOption) Apply(s *MyGrpcServer) {
	d.f(s)
}

func WithGrpcOptions(options ...grpc.ServerOption) MsGrpcOption {
	return DefaultGrpcOption{f: func(s *MyGrpcServer) {
		s.ops = append(s.ops, options...)
	}}
}
