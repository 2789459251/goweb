package rpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync/atomic"
	"time"
)

//TCP 客户端 服务端

//cli:连接服务端 -发送请求数据，编码 - 等回复
//server:启动服务 接受请求 ，解码 调用服务 得到响应 发送客户端

type Serializer interface {
	Serialize(i interface{}) ([]byte, error)
	Deserialize(data []byte, i interface{}) error
}
type GobSerializer struct{}

func (c GobSerializer) Serialize(data any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (c GobSerializer) Deserialize(data []byte, target any) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(target)
}

//type Serializer interface {
//	Serialize(data any) []byte
//	Deserialize(data []byte, target any) error
//}
//
//type GobSerializer struct{}
//
//func (c *GobSerializer) Serialize(data any) ([]byte, error) {
//	var buffer bytes.Buffer
//	encoder := gob.NewEncoder(&buffer)
//	if err := encoder.Encode(data); err != nil {
//		return nil, err
//	}
//	return buffer.Bytes(), nil
//}
//
//func (c *GobSerializer) Deserialize(data []byte, target any) error {
//	buffer := bytes.NewBuffer(data)
//	decoder := gob.NewDecoder(buffer)
//	return decoder.Decode(target)
//}

type MyRpcMessage struct {
	//头
	Header *Header
	//消息体
	Data any
}

const mn byte = 0x1d
const version = 0x01

type CompressType byte

const (
	Gzip CompressType = iota
)

type SerializeType byte

const (
	Gob SerializeType = iota
	ProtoBuff
)

type MessageType byte

const (
	msgRequest MessageType = iota
	msgResponse
	msgPing
	msgPong
)

type Header struct {
	MagicNumber   byte
	Version       byte
	FullLength    int32
	MessageType   MessageType
	CompressType  CompressType
	SerializeType SerializeType
	RequestId     int64
}

type MyRpcRequest struct {
	RequestId   int64
	ServiceName string
	MethodName  string
	Args        []any
}

type MyRpcResponse struct {
	RequestId     int64
	Code          int16
	Msg           string
	CompressType  CompressType
	SerializeType SerializeType
	Data          any
}

type MyRpcServer interface {
	Register(name string, service interface{})
	Run()
	Stop()
}

type MyTcpServer struct {
	listener   net.Listener
	Host       string
	Port       int
	Network    string
	serviceMap map[string]interface{}
}

type MsTcpConn struct {
	s       *MyTcpServer
	conn    net.Conn
	rspChan chan *MyRpcResponse
}

/*服务端发送数据*/
func (c *MsTcpConn) writeHandle() {
	ctx := context.Background()
	_, cancel := context.WithTimeout(ctx, time.Duration(3)*time.Second)
	defer cancel()
	select {
	case rsp := <-c.rspChan:
		//编码数据
		err := c.Send(c.conn, rsp)
		if err != nil {
			log.Println(err)
		}
		return
	case <-ctx.Done():
		log.Println("超时了")
		return
	}
}

func (c *MsTcpConn) Send(conn net.Conn, rsp *MyRpcResponse) error {
	headers := make([]byte, 17)
	//魔法数（Magic Number）	1
	//版本（Version）		1
	//消息长度（full length）	4
	//消息类型（messageType）	1
	//压缩类型（compress）	1
	//序列化类型（serialize）	1
	//请求id（requestId）	8
	//请求体（body）
	//magic number
	headers[0] = mn
	//version
	headers[1] = version
	//full length
	//消息类型
	headers[6] = byte(msgResponse)
	//压缩类型
	headers[7] = byte(rsp.CompressType)
	//序列化
	headers[8] = byte(rsp.SerializeType)
	//请求id
	binary.BigEndian.PutUint64(headers[9:], uint64(rsp.RequestId))

	serializer, err := loadSerialize(rsp.SerializeType)
	if err != nil {
		return err
	}
	body, err := serializer.Serialize(rsp)
	if err != nil {
		return err
	}
	body, err = compress(body, rsp.CompressType)
	if err != nil {
		return err
	}
	fullLen := 17 + len(body)
	binary.BigEndian.PutUint32(headers[2:6], uint32(fullLen))
	_, err = conn.Write(headers[:])
	if err != nil {
		return err
	}
	err = binary.Write(c.conn, binary.BigEndian, body[:])
	if err != nil {
		return err
	}
	log.Println("发送数据成功")
	return nil
}

func NewTcpServer(host string, port int) *MyTcpServer {
	return &MyTcpServer{
		Host:    host,
		Port:    port,
		Network: "tcp",
	}
}
func (s *MyTcpServer) Register(name string, service interface{}) {
	if s.serviceMap == nil {
		s.serviceMap = make(map[string]interface{})
	}
	v := reflect.ValueOf(service)
	if v.Kind() != reflect.Pointer {
		panic(errors.New("service not pointer"))
	}
	s.serviceMap[name] = service
}
func (s *MyTcpServer) Run() {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	listen, err := net.Listen(s.Network, addr)
	if err != nil {
		panic(err)
	}
	s.listener = listen
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		msConn := &MsTcpConn{conn: conn, rspChan: make(chan *MyRpcResponse, 1), s: s}
		go s.readHandle(msConn)
		go msConn.writeHandle()
	}
}

func (s *MyTcpServer) Stop() {
	//addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	//listen, err := net.Listen(s.Network, addr)
	//if err != nil {
	//	panic(err)
	//}
	//s.listener = listen
	//for {
	//	conn, err := s.listener.Accept()
	//	if err != nil {
	//		log.Println(err)
	//		continue
	//	}
	//	msConn := &MsTcpConn{conn: conn, rspChan: make(chan *MyRpcResponse, 1), s: s}
	//	go s.readHandle(msConn)
	//	go msConn.writeHandle()
	//}
	err := s.listener.Close()
	if err != nil {
		log.Println(err)
	}
}

/*接受请求 解码请求*/
func (s *MyTcpServer) readHandle(msConn *MsTcpConn) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			msConn.conn.Close()
		}
	}()
	//解码请求
	msg := s.decodeFrame(msConn.conn)
	if msg == nil {
		msConn.rspChan <- nil
		return
	}
	//根据请求,调用服务
	if msg.Header.MessageType == msgRequest {
		//请求id
		req := msg.Data.(*MyRpcRequest)
		//查找注册的服务匹配后进行调用，调用完发送到一个channel当中
		service, ok := s.serviceMap[req.ServiceName]
		rsp := &MyRpcResponse{RequestId: req.RequestId, CompressType: msg.Header.CompressType, SerializeType: msg.Header.SerializeType}
		if !ok {
			rsp.Code = 500
			rsp.Msg = "no service found"
			msConn.rspChan <- rsp
			return
		}
		//查找方法
		v := reflect.ValueOf(service)
		reflectMethod := v.MethodByName(req.MethodName)

		if reflectMethod.IsNil() {
			rsp.Code = 500
			rsp.Msg = "no service method found"
			msConn.rspChan <- rsp
			return
		}

		args := make([]reflect.Value, len(req.Args))
		for i := range req.Args {
			args[i] = reflect.ValueOf(req.Args[i])
		}
		result := reflectMethod.Call(args)
		if len(result) == 0 {
			//无返回结果
			rsp.Code = 200
			msConn.rspChan <- rsp
			return
		}
		resArgs := make([]interface{}, len(result))
		for i := 0; i < len(result); i++ {
			resArgs[i] = result[i].Interface()
		}
		var err error
		if _, ok := result[len(result)-1].Interface().(error); ok {
			err = result[len(result)-1].Interface().(error)
		}

		if err != nil {
			rsp.Code = 500
			rsp.Msg = err.Error()
			log.Println("接收数据出错，服务方法调用出错")
			msConn.rspChan <- rsp
			return
		}
		rsp.Code = 200
		//rpc Method(request)response,error,所以data返回result[0],如果有错误在信息中加上错误信息
		rsp.Data = resArgs[0]
		msConn.rspChan <- rsp
		log.Println("接收数据成功")
		return
	}
}

func (s *MyTcpServer) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func (*MyTcpServer) decodeFrame(conn net.Conn) *MyRpcMessage {
	//读取数据 先读取header部分
	//1+1+4+1+1+1+8 = 17字节
	headers := make([]byte, 17)
	_, err := io.ReadFull(conn, headers)
	if err != nil {
		log.Println(err)
		return nil
	}
	//magic number
	magicNumber := headers[0]
	if magicNumber != mn {
		log.Println("magic number not valid : ", magicNumber)
		return nil
	}
	//version
	version := headers[1]
	//消息长度
	fullLength := headers[2:6]
	//消息类型
	mt := headers[6]
	messageType := MessageType(mt)
	//压缩类型
	compressType := headers[7]
	//序列化类型
	serializeType := headers[8]
	//请求id
	requestId := headers[9:]

	//将body解析出来，包装成request 根据请求内容查找对应的服务，完成调用
	//网络调用 大端
	fl := int32(binary.BigEndian.Uint32(fullLength))
	//请求体长度
	bodyLen := fl - 17
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(conn, body)
	log.Println("读完了")
	if err != nil {
		log.Println(err)
		return nil
	}
	//请求体，先解压
	body, err = unCompress(body, CompressType(compressType))
	if err != nil {
		log.Println(err)
		return nil
	}
	//请求体，反序列化
	serializer, err := loadSerialize(SerializeType(serializeType))
	if err != nil {
		log.Println(err)
		return nil
	}
	header := &Header{}
	header.MagicNumber = magicNumber
	header.FullLength = fl
	header.CompressType = CompressType(compressType)
	header.Version = version
	header.SerializeType = SerializeType(serializeType)
	header.RequestId = int64(binary.BigEndian.Uint64(requestId))
	header.MessageType = messageType

	//判断rpc请求或响应，封装msg
	if messageType == msgRequest {
		msg := &MyRpcMessage{}
		msg.Header = header
		req := &MyRpcRequest{}
		err := serializer.Deserialize(body, req)
		if err != nil {
			log.Println(err)
			return nil
		}
		msg.Data = req
		return msg
	}
	if messageType == msgResponse {
		msg := &MyRpcMessage{}
		msg.Header = header
		rsp := &MyRpcResponse{}
		err := serializer.Deserialize(body, rsp)
		if err != nil {
			log.Println(err)
			return nil
		}
		msg.Data = rsp
		return msg
	}
	return nil
}

func loadSerialize(serializeType SerializeType) (Serializer, error) {
	switch serializeType {
	case Gob:
		//gob
		s := &GobSerializer{}
		return s, nil
	}
	return nil, errors.New("no serializeType")
}

func compress(body []byte, compressType CompressType) ([]byte, error) {
	switch compressType {
	case Gzip:
		//return body, nil
		//gzip
		//创建一个新的 byte 输出流
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)

		_, err := w.Write(body)
		if err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, errors.New("no compressType")
}

func unCompress(body []byte, compressType CompressType) ([]byte, error) {
	switch compressType {
	case Gzip:
		//return body, nil
		//gzip
		reader, err := gzip.NewReader(bytes.NewReader(body))
		defer reader.Close()
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		// 从 Reader 中读取出数据
		if _, err := buf.ReadFrom(reader); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, errors.New("no compressType")
}

//客户端底层逻辑

type MsRpcClient interface {
	Connect() error
	Invoke(context context.Context, serviceName string, methodName string, args []any) (any, error)
	Close() error
}

type MsTcpClient struct {
	conn   net.Conn
	option TcpClientOption
}

type TcpClientOption struct {
	Retries           int
	ConnectionTimeout time.Duration
	SerializeType     SerializeType
	CompressType      CompressType
	Host              string
	Port              int
}

var DefaultOption = TcpClientOption{
	Host:              "127.0.0.1",
	Port:              9222,
	Retries:           3,
	ConnectionTimeout: 5 * time.Second,
	SerializeType:     Gob,
	CompressType:      Gzip,
}

func NewTcpClient(option TcpClientOption) *MsTcpClient {
	return &MsTcpClient{option: option}
}

func (c *MsTcpClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.option.Host, c.option.Port)
	conn, err := net.DialTimeout("tcp", addr, c.option.ConnectionTimeout)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

var reqId int64

// 封装请求 编码 发送消息 读取响应
func (c *MsTcpClient) Invoke(ctx context.Context, serviceName string, methodName string, args []any) (any, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, c.option.ConnectionTimeout)
	defer cancel()

	req := &MyRpcRequest{}
	req.RequestId = atomic.AddInt64(&reqId, 1)
	req.ServiceName = serviceName
	req.MethodName = methodName
	req.Args = args

	headers := make([]byte, 17)
	//magic number
	headers[0] = mn
	//version
	headers[1] = version
	//full length
	//消息类型
	headers[6] = byte(msgRequest)
	//压缩类型
	headers[7] = byte(c.option.CompressType)
	//序列化
	headers[8] = byte(c.option.SerializeType)
	//请求id
	binary.BigEndian.PutUint64(headers[9:], uint64(req.RequestId))

	serializer, err := loadSerialize(c.option.SerializeType)
	if err != nil {
		return nil, err
	}
	body, err := serializer.Serialize(req)
	if err != nil {
		return nil, err
	}
	body, err = compress(body, c.option.CompressType)
	if err != nil {
		return nil, err
	}
	fullLen := 17 + len(body)
	binary.BigEndian.PutUint32(headers[2:6], uint32(fullLen))
	_, err = c.conn.Write(headers[:])
	if err != nil {
		return nil, err
	}
	err = binary.Write(c.conn, binary.BigEndian, body[:])
	if err != nil {
		return nil, err
	}
	rspChan := make(chan *MyRpcResponse)
	go c.readHandle(rspChan)
	rsp := <-rspChan
	return rsp, nil
}

func (c *MsTcpClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

/*读取响应放进管道*/
func (c *MsTcpClient) readHandle(rspChan chan *MyRpcResponse) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			c.conn.Close()
		}
	}()
	for {
		//解码
		msg := c.decodeFrame(c.conn)
		if msg == nil {
			log.Println("未解析出任何数据")
			rspChan <- nil
			return
		}
		//根据请求
		if msg.Header.MessageType == msgResponse {
			rsp := msg.Data.(*MyRpcResponse)
			rspChan <- rsp
			return
		}
	}
}

func (*MsTcpClient) decodeFrame(conn net.Conn) *MyRpcMessage {
	//读取数据 先读取header部分
	//1+1+4+1+1+1+8 = 17字节
	headers := make([]byte, 17)
	_, err := io.ReadFull(conn, headers)
	if err != nil {
		log.Println(err)
		return nil
	}
	//magic number
	magicNumber := headers[0]
	if magicNumber != mn {
		log.Println("magic number not valid : ", magicNumber)
		return nil
	}
	//version
	version := headers[1]
	//
	fullLength := headers[2:6]
	//
	mt := headers[6]
	messageType := MessageType(mt)
	//压缩类型
	compressType := headers[7]
	//序列化类型
	serializeType := headers[8]
	//请求id
	requestId := headers[9:]

	//将body解析出来，包装成request 根据请求内容查找对应的服务，完成调用
	//网络调用 大端
	fl := int32(binary.BigEndian.Uint32(fullLength))
	bodyLen := fl - 17
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(conn, body)
	log.Println("读完了")
	if err != nil {
		log.Println(err)
		return nil
	}
	//先解压
	body, err = unCompress(body, CompressType(compressType))
	if err != nil {
		log.Println(err)
		return nil
	}
	//反序列化
	serializer, err := loadSerialize(SerializeType(serializeType))
	if err != nil {
		log.Println(err)
		return nil
	}
	header := &Header{}
	header.MagicNumber = magicNumber
	header.FullLength = fl
	header.CompressType = CompressType(compressType)
	header.Version = version
	header.SerializeType = SerializeType(serializeType)
	header.RequestId = int64(binary.BigEndian.Uint64(requestId))
	header.MessageType = messageType

	if messageType == msgRequest {
		msg := &MyRpcMessage{}
		msg.Header = header
		req := &MyRpcRequest{}
		err := serializer.Deserialize(body, req)
		if err != nil {
			log.Println(err)
			return nil
		}
		msg.Data = req
		return msg
	}
	if messageType == msgResponse {
		msg := &MyRpcMessage{}
		msg.Header = header
		rsp := &MyRpcResponse{}
		err := serializer.Deserialize(body, rsp)
		if err != nil {
			log.Println(err)
			return nil
		}
		msg.Data = rsp
		return msg
	}
	return nil
}

type MsTcpClientProxy struct {
	client *MsTcpClient
	option TcpClientOption
}

func NewMsTcpClientProxy(option TcpClientOption) *MsTcpClientProxy {
	return &MsTcpClientProxy{option: option}
}

func (p *MsTcpClientProxy) Call(ctx context.Context, serviceName string, methodName string, args []any) (any, error) {
	client := NewTcpClient(p.option)
	p.client = client
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	for i := 0; i < p.option.Retries; i++ {
		result, err := client.Invoke(ctx, serviceName, methodName, args)
		if err != nil {
			if i >= p.option.Retries-1 {
				log.Println(errors.New("already retry all time"))
				client.Close()
				return nil, err
			}
			continue
		}
		client.Close()
		return result, nil
	}
	return nil, errors.New("retry time is 0")
}
