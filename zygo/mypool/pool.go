package mypool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type sig struct{}

//var DefaultExpire time.Duration = 3 //超时时间为3s

var (
	ErrorInvalidCap = errors.New("pool cap can not <= 0")
	ErrorExpired    = errors.New("pool expired can not <= 0")
	ErrorPoolClosed = errors.New("pool is closed")
)

const DEFAULTEXPIRE = 3

type pool struct {
	cap     int32         //容量
	running int32         //正在运行的worker数量
	workers []*Worker     //空闲的worker队列
	expire  time.Duration //超时时间，空闲worker回收
	release chan sig      //关闭信号
	lock    sync.Mutex    //锁机制保证协程池中资源安全
	once    sync.Once     //释放协程的，保证只能使用一次
}

func NewPool(cap int) (*pool, error) {
	p, err := NewTimePool(cap, DEFAULTEXPIRE)
	return p, err
}

func NewTimePool(cap int, expire time.Duration) (*pool, error) {

	if cap <= 0 {
		return nil, ErrorInvalidCap
	}
	if expire <= 0 {
		return nil, ErrorExpired
	}
	p := &pool{
		cap:     int32(cap),
		release: make(chan sig, 1),
		expire:  expire * time.Second,
	}
	go p.expiredWorker()
	return p, nil
}

/* 空闲超时清理 */
func (p *pool) expiredWorker() {
	//定时清理空闲的worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		curentTime := time.Now()
		if p.IsClose() {
			break
		}
		idleWorker := p.workers
		n := -1
		p.lock.Lock()
		//循环空闲的，清除满足条件的
		for i, w := range idleWorker {
			if curentTime.Sub(w.lastTime) <= p.expire {
				//后面的结束时间比前面的晚
				break
			}
			//清理i以及i之前的
			n = i
			w.tasks <- nil
			idleWorker[i] = nil
			if n >= len(idleWorker)-1 {
				p.workers = idleWorker[:0]
			} else {
				p.workers = idleWorker[n+1:]
			}

			fmt.Printf("清理任务结束：running : %d workers:%v\n", p.running, p.workers)
		}
		p.lock.Unlock()
	}
}

/*获取worker -> 提交任务*/
func (p *pool) Submit(task func()) (err error) {
	if len(p.release) > 0 {
		return ErrorPoolClosed
	}
	worker := p.GetWorker()
	worker.tasks <- task
	p.incRunning()
	return nil
}

/*核心 ：获取worker的机制*/
func (p *pool) GetWorker() *Worker {
	//1. 目的获取pool里面的worker
	//2. 如果 有空闲的worker 直接获取

	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	//如果有现成的worker 直接调用
	if n >= 0 {
		p.lock.Lock()
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
	//3. 如果没有空闲的worker，要新建一个worker
	if p.running < p.cap {
		//还不够pool的容量，直接新建一个
		w := &Worker{
			pool:  p,
			tasks: make(chan func(), 1),
		}
		w.run()
		return w
	}
	//4. 如果正在运行的workers 如果大于pool容量，阻塞等待，worker释放
	for {
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n < 0 {
			p.lock.Unlock()
			continue
		}
		w := idleWorkers[n]
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
}

/* 修改池子里运行worker的数量 ,开始运行时使用*/
func (p *pool) incRunning() {
	/* 通过atomic.AddInt32来对其进行原子操作，确保在并发环境下增加其值时不会产生竞态条件 */
	atomic.AddInt32(&p.running, 1)
}

/*修改池子里运行worker的数量，结束运行使用*/
func (p *pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

/* 释放worker */
func (p *pool) PutWorker(worker *Worker) {
	worker.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	p.lock.Unlock()
}

/* 释放协程池 */
func (p *pool) Release() {
	p.once.Do(func() {
		p.lock.Lock()
		for i, _ := range p.workers {
			p.workers[i].tasks = nil
			p.workers[i].pool = nil
			p.workers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
		p.release <- sig{}
	})
}

/* 重新开启协程池 */
func (p *pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = p.release
	go p.expiredWorker()
	return true
}

/* 验证协程池的状态是否关闭 */
func (p *pool) IsClose() bool {
	return len(p.release) > 0
}
