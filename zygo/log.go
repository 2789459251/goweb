package zygo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//格式化操作

// 颜色样式
const (
	//背景颜色颜色
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// 标准输出
var DefaultWriter io.Writer = os.Stdout

// 默认打印,返回打印语句
var defaultFormatter = func(params *LogFormatterParams) string {
	//设置颜色
	var statusCodeColor = params.StatusCodeColor()
	//将颜色转换为原来
	var reset = params.ResetColor()
	//打印时间 状态码 时间 ip 方法 路径
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	//return fmt.Sprintf("[zygo] %v |%s %3d %s| %13v | %15s | %-7s %#v",
	//	params.TimeStamp.Format("2006/01/02--15:04:05"),
	//	statusCodeColor, params.StatusCode, reset, params.Latency, params.ClientIP, params.Method, params.Path)
	if params.IsDisplayColor {
		return fmt.Sprintf("%s [zygo] %s |%s %v %s| %s %3d %s |%s %13v %s| %15s  |%s %-7s %s %s %#v %s",
			yellow, reset, blue, params.TimeStamp.Format("2006/01/02 - 15:04:05"), reset,
			statusCodeColor, params.StatusCode, reset,
			red, params.Latency, reset,
			params.ClientIP,
			magenta, params.Method, reset,
			cyan, params.Path, reset,
		)
	} else {
		return fmt.Sprintf("[zygo] %v |%s %3d %s| %13v | %15s | %-7s %#v",
			params.TimeStamp.Format("2006/01/02--15:04:05"),
			statusCodeColor, params.StatusCode, reset, params.Latency, params.ClientIP, params.Method, params.Path)
	}

}

// 中间件
type LoggerConfig struct {
	Formatter LogerFormatter
	out       io.Writer
}

// 抽象出
type LogerFormatter = func(params *LogFormatterParams) string

// 提取出信息
type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	IsDisplayColor bool
}

func LoggingWithConfig(conf LoggerConfig, next HandlerFunc) HandlerFunc {
	formatter := conf.Formatter
	// 没有规定打印配置，就执行默认
	if formatter == nil {
		formatter = defaultFormatter
	}

	out := conf.out
	if out == nil {
		out = DefaultWriter
	}
	return func(ctx *Context) {
		r := ctx.R

		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery
		next(ctx)
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := r.Method
		status := ctx.StatusCode

		param := &LogFormatterParams{
			Request:        r,
			IsDisplayColor: false,
			TimeStamp:      stop,
			StatusCode:     status,
			Latency:        latency,
			ClientIP:       clientIP,
			Method:         method,
			Path:           path,
		}
		if raw != "" {
			path = path + "?" + raw
		}

		//打印时间 状态码 时间 ip 方法 路径
		fmt.Fprintln(out, formatter(param))
	}
}

func Logging(next HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggerConfig{}, next)
}
