package zygo

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"web/zygo/zyerror"
)

func detailMsg(err any) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v\n", err))
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}

func Recovery(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				if err2 := err.(error); err2 != nil {
					var myerror *zyerror.MyError
					if errors.As(err2, &myerror) {
						myerror.ExcuResult()
						//return
					}
				}
				ctx.Logger.Error(detailMsg(err))
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		next(ctx)
	}
}
