package zygo

import (
	"log"
	"net"
	"strings"
	"time"
)

// 中间件
type LoggerConfig struct{}

func LoggingWithConfig(conf LoggerConfig, next HandlerFunc) HandlerFunc {
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

		if raw != "" {
			path = path + "?" + raw
		}

		//打印时间 状态码 时间 ip 方法 路径
		log.Printf("[zygo] %v | %3d | %13v | %15s | %-7s %#v",
			stop.Format("2006/01/02--15:04:05"),
			status, latency, clientIP, method, path)
	}
}

func Logging(next HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggerConfig{}, next)
}
