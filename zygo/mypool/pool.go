package mypool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"web/zygo/config"
)

type sig struct{}

//var DefaultExpire time.Duration = 3 //超时时间为3s

var (
	ErrorInvalidCap = errors.New("Pool cap can not <= 0")
	ErrorExpired    = errors.New("Pool expired can not <= 0")
	ErrorPoolClosed = errors.New("Pool is closed")
)

const DEFAULTEXPIRE = 3

type Pool struct {
	cap          int32         //容量
	running      int32         //正在运行的worker数量
	workers      []*Worker     //空闲的worker队列
	expire       time.Duration //超时时间，空闲worker回收
	release      chan sig      //关闭信号
	lock         sync.Mutex    //锁机制保证协程池中资源安全
	once         sync.Once     //释放协程的，保证只能使用一次
	workercache  sync.Pool     //workerCache 缓存
	cond         *sync.Cond    //用于协程通信
	PanicHandler func()        //PanicHandler

}

func NewPoolConf() (*Pool, error) {
	cap, ok := config.Conf.Pool["cap"]
	if !ok {
		return nil, errors.New("cap config not exist")
	}
	return NewTimePool(int(cap.(int64)), DEFAULTEXPIRE)
}
func NewPool(cap int) (*Pool, error) {
	p, err := NewTimePool(cap, DEFAULTEXPIRE)
	return p, err
}

func NewTimePool(cap int, expire time.Duration) (*Pool, error) {

	if cap <= 0 {
		return nil, ErrorInvalidCap
	}
	if expire <= 0 {
		return nil, ErrorExpired
	}
	p := &Pool{
		cap:     int32(cap),
		release: make(chan sig, 1),
		expire:  expire * time.Second,
	}
	p.workercache.New = func() any {
		return &Worker{
			pool:  p,
			tasks: make(chan func(), 1),
		}
	}
	p.cond = sync.NewCond(&p.lock)
	go p.expiredWorker()
	return p, nil
}

/* 空闲超时清理 */
func (p *Pool) expiredWorker() {
	//定时清理过期的空闲worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClose() {
			break
		}
		//循环空闲的workers 如果当前时间和worker的最后运行任务的时间 差值大于expire 进行清理
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n >= 0 {
			var clearN = -1
			for i, w := range idleWorkers {
				if time.Now().Sub(w.lastTime) <= p.expire {
					break
				}
				clearN = i
				w.tasks <- nil
				idleWorkers[i] = nil
			}
			// 3 2
			if clearN != -1 {
				if clearN >= len(idleWorkers)-1 {
					p.workers = idleWorkers[:0]
				} else {
					// len=3 0,1 del 2
					p.workers = idleWorkers[clearN+1:]
				}
				fmt.Printf("清除完成,running:%d, workers:%v \n", p.running, p.workers)
			}
		}
		p.lock.Unlock()
	}
}

/*获取worker -> 提交任务*/
func (p *Pool) Submit(task func()) (err error) {
	if len(p.release) > 0 {
		return ErrorPoolClosed
	}
	worker := p.GetWorker()
	worker.tasks <- task
	return nil
}

/*核心 ：获取worker的机制*/
func (p *Pool) GetWorker() (w *Worker) {
	//1. 目的获取pool里面的worker
	//2. 如果 有空闲的worker 直接获取
	var readyWorker = func() {
		w = p.workercache.Get().(*Worker)
		w.run()
	}
	//🔓
	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	//如果有现成的worker 直接调用
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return
	}
	//3. 如果没有空闲的worker，要新建一个worker
	if p.running < p.cap {
		//	//	//先从缓存中拿
		p.lock.Unlock()
		readyWorker()
		return
	}
	p.lock.Unlock()
	w = p.waitIdleWorker()
	return
}

/*阻塞等待worker被释放*/
func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait()
	fmt.Println("被唤醒")
	idledWorker := p.workers
	n := len(idledWorker) - 1
	if n < 0 {
		p.lock.Unlock()
		if p.running < p.cap {
			//还不够pool的容量，直接新建一个
			c := p.workercache.Get()
			var w *Worker
			if c == nil {
				w = &Worker{
					pool:  p,
					tasks: make(chan func(), 1),
				}
			} else {
				w = c.(*Worker)
			}
			w.run()
			return w
		}
		return p.waitIdleWorker()
	}
	worker := idledWorker[n]
	p.workers = idledWorker[:n]
	idledWorker[n] = nil
	p.lock.Unlock()
	return worker
}

/* 修改池子里运行worker的数量 ,开始运行时使用*/
func (p *Pool) incRunning() {
	/* 通过atomic.AddInt32来对其进行原子操作，确保在并发环境下增加其值时不会产生竞态条件 */
	atomic.AddInt32(&p.running, 1)
}

/*修改池子里运行worker的数量，结束运行使用*/
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

/* 释放worker */
func (p *Pool) PutWorker(worker *Worker) {
	worker.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	p.cond.Signal()
	p.lock.Unlock()
}

/* 释放协程池 */
func (p *Pool) Release() {
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
func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = p.release
	go p.expiredWorker()
	return true
}

/* 验证协程池的状态是否关闭 */
func (p *Pool) IsClose() bool {
	return len(p.release) > 0
}

func (p *Pool) Running() string {
	return fmt.Sprintf("%d", atomic.LoadInt32(&p.running))
}
func (p *Pool) Free() int {
	return int((p.cap - p.running))
}
