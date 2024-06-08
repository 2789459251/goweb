package mypool

import (
	"time"
	"web/zygo/mylog"
)

type Worker struct {
	pool  *Pool
	tasks chan func()
	//执行任务的最后时间、便于进行超时释放操作
	lastTime time.Time
}

func (w *Worker) run() {
	/* 开启协程，运行任务 */
	w.pool.incRunning()
	go w.running()
}

func (w *Worker) running() {
	defer func() {
		w.pool.decRunning()
		w.pool.workercache.Put(w)
		if err := recover(); err != nil {
			//捕获任务发生的panic
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				mylog.Default().Error(err)
			}
		}
		w.pool.cond.Signal()
	}()
	for f := range w.tasks {
		if f == nil {
			return
		}
		f()
		/*任务运行完成，worker空闲*/
		w.pool.PutWorker(w)
	}
}
