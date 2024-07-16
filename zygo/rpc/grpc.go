package rpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

type MyGrpcServer struct {
	listen     net.Listener
	grpcServer *grpc.Server
	registers  []func(grpcServer *grpc.Server)
	ops        []grpc.ServerOption
}

func NewGrpcServer(address string, ops ...MyGrpcOption) (*MyGrpcServer, error) {
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

type MyGrpcOption interface {
	Apply(s *MyGrpcServer)
}

type DefaultGrpcOption struct {
	f func(s *MyGrpcServer)
}

func (d DefaultGrpcOption) Apply(s *MyGrpcServer) {
	d.f(s)
}

func WithGrpcOptions(options ...grpc.ServerOption) MyGrpcOption {
	return DefaultGrpcOption{f: func(s *MyGrpcServer) {
		s.ops = append(s.ops, options...)
	}}
}

type MyGrpcClient struct {
	Conn *grpc.ClientConn
}

func NewGrpcClient(config *MyGrpcClientConfig) (*MyGrpcClient, error) {
	var ctx = context.Background()
	var dialOptions = config.dialOptions

	if config.Block {
		//阻塞
		if config.DialTimeout > time.Duration(0) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)
			defer cancel()
		}
		dialOptions = append(dialOptions, grpc.WithBlock())
	}
	if config.KeepAlive != nil {
		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(*config.KeepAlive))
	}
	conn, err := grpc.DialContext(ctx, config.Address, dialOptions...)
	if err != nil {
		return nil, err
	}
	return &MyGrpcClient{
		Conn: conn,
	}, nil
}

type MyGrpcClientConfig struct {
	Address     string
	Block       bool
	DialTimeout time.Duration
	ReadTimeout time.Duration
	Direct      bool
	KeepAlive   *keepalive.ClientParameters
	dialOptions []grpc.DialOption
}

func DefaultGrpcClientConfig() *MyGrpcClientConfig {
	return &MyGrpcClientConfig{
		dialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
		DialTimeout: time.Second * 3,
		ReadTimeout: time.Second * 2,
		Block:       true,
	}
}
