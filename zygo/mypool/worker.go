package mypool

import (
	"time"
)

type Worker struct {
	pool  *pool
	tasks chan func()
	//执行任务的最后时间、便于进行超时释放操作
	lastTime time.Time
}

func (w *Worker) run() {
	/* 开启协程，运行任务 */
	go w.running()
}

func (w *Worker) running() {
	for f := range w.tasks {
		if f == nil {
			return
		}
		f()
		/*任务运行完成，worker空闲*/
		w.pool.PutWorker(w)
		w.pool.decRunning()
	}
}
