package rpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type MyHttpClient struct {
	client      *http.Client
	ServicesMap map[string]MyService
}

func NewHttpClient() *MyHttpClient {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return &MyHttpClient{client: client, ServicesMap: make(map[string]MyService)} //请求分发，协程安全

}
func (s *MyHttpClientSession) Get(url string, args map[string]any) ([]byte, error) {
	if args != nil && len(args) > 0 {
		url = url + "?" + s.toValues(args)
	}
	log.Println(url)
	//req := &http.Request{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	//
	//response, err := cli.Do(req)
	//body := response.Body
	//bufio.NewReader(body)
	return s.responsehandle(req)
}

func (s *MyHttpClientSession) responsehandle(req *http.Request) ([]byte, error) {
	s.ReqHandler(req)
	//在调用之前  session设置jaeger 信息header
	resp, err := s.MyHttpClient.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status " + resp.Status)
	}
	reader := bufio.NewReader(resp.Body)
	var buf = make([]byte, 127)
	var body []byte
	for {
		n, err := reader.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		body = append(body, buf[:n]...)
		if n < len(buf) {
			break
		}
	}
	return body, nil
}

func (*MyHttpClient) toValues(args map[string]any) string {
	if args != nil && len(args) > 0 {
		params := url.Values{}
		for k, v := range args {
			params.Set(k, fmt.Sprintf("%v", v))
		}
		return params.Encode()
	}
	return ""
}
func (s *MyHttpClientSession) PostForm(url string, args map[string]any) ([]byte, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(s.toValues(args)))
	if err != nil {
		return nil, err
	}
	return s.responsehandle(req)
}
func (s *MyHttpClientSession) PostJson(url string, args map[string]any) ([]byte, error) {
	marshal, _ := json.Marshal(args)
	req, err := http.NewRequest("POST ", url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	return s.responsehandle(req)
}

func (s *MyHttpClientSession) GetRequest(method string, url string, args map[string]any) (*http.Request, error) {
	if args != nil && len(args) > 0 {
		url = url + "?" + s.toValues(args)
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *MyHttpClient) FormRequest(method string, url string, args map[string]any) (*http.Request, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	return req, nil
}
func (s *MyHttpClientSession) Response(req *http.Request) ([]byte, error) {
	return s.responsehandle(req)
}

func (c *MyHttpClient) JsonRequest(method string, url string, args map[string]any) (*http.Request, error) {
	jsonStr, _ := json.Marshal(args)
	req, err := http.NewRequest(method, url, bytes.NewReader(jsonStr))
	if err != nil {
		return nil, err
	}
	return req, nil
}

const (
	HTTP  = "http"
	HTTPS = "https"
)
const (
	GET      = "GET"
	POSTForm = "POST_FORM"
	POSTJson = "POST_JSON"
)

type HttpConfig struct {
	Protocol string
	Host     string
	Port     int
}
type MyService interface {
	Env() HttpConfig
}

func (c *MyHttpClient) RegisterHttpService(name string, service MyService) {
	c.ServicesMap[name] = service
}

type MyHttpClientSession struct {
	*MyHttpClient
	ReqHandler func(req *http.Request)
}

func (c *MyHttpClient) NewSession() *MyHttpClientSession {
	return &MyHttpClientSession{
		c,
		nil,
	}
}

// A 调用 B服务
func (s *MyHttpClientSession) Do(service string, method string) MyService {

	myService, ok := s.MyHttpClient.ServicesMap[service]
	if !ok {
		panic(errors.New("service " + service + " not exist"))
	}
	//找到service里的Field给其中要调用的方法、赋值
	t := reflect.TypeOf(myService)
	v := reflect.ValueOf(myService)
	if t.Kind() != reflect.Ptr {
		panic(errors.New("service " + service + " not ptr"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	fieldIndex := -1
	for i := 0; i < tVar.NumField(); i++ {
		name := tVar.Field(i).Name
		if name == method {
			fieldIndex = i
			break
		}
	}
	if fieldIndex == -1 {
		panic(errors.New("method " + method + " not exist"))
	}
	tag := tVar.Field(fieldIndex).Tag
	rpcInfo := tag.Get("myrpc")
	if rpcInfo == "" {
		panic(errors.New("myrpc is empty"))
	}
	split := strings.Split(rpcInfo, ",")
	if len(split) != 2 {
		panic(errors.New("rpcInfo " + rpcInfo + " format error"))
	}
	methodType := split[0]
	path := split[1]
	httpConfig := myService.Env()
	f := func(args map[string]any) ([]byte, error) {
		if methodType == GET {
			return s.Get(httpConfig.Prefix()+path, args)
		}
		if methodType == POSTForm {
			return s.PostForm(httpConfig.Prefix()+path, args)
		}
		if methodType == POSTJson {
			return s.PostJson(httpConfig.Prefix()+path, args)
		}
		return nil, errors.New("no match method type")
	}
	fValue := reflect.ValueOf(f)
	vVar.Field(fieldIndex).Set(fValue)
	return myService
}

func (c HttpConfig) Prefix() string {
	if c.Protocol == "" {
		c.Protocol = HTTP
	}
	switch c.Protocol {
	case HTTP:
		return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
	case HTTPS:
		return fmt.Sprintf("https://%s:%d", c.Host, c.Port)
	}
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}
