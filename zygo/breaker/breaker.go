package breaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

/*										熔断器的实现
当连续失败大于阈值 则打开断路器，打开状态，当超过一段时间，设置为半开状态，当连续成功大于阈值，则设置为关闭状态。
*/
//
//type State int
//
//const (
//	StateClosed State = iota
//	StateHalfOpen
//	StateOpen
//)
//
//var (
//	//间隔时间
//	defaultInterval = time.Duration(0 * time.Second)
//	//超时时间
//	defaultTimeout = time.Duration(20 * time.Second)
//	//是否执行熔断
//	defaultReadyToTrip = func(count Counts) bool {
//		return count.ConsecutiveFailures > 5
//	}
//	// 是否成功
//	defaultIsSuccessful = func(err error) bool {
//		return err == nil
//	}
//)
//
///*计数*/
//type Counts struct {
//	Requests             uint32 //请求数量
//	TotalSuccesses       uint32 //总成功数
//	TotalFailures        uint32 //总失败数
//	ConsecutiveSuccesses uint32 //连续成功数量!
//	ConsecutiveFailures  uint32 //连续失败数量
//}
//
//func (c *Counts) onRequest() {
//	c.Requests++
//}
//
//func (c *Counts) onSuccess() {
//	c.TotalSuccesses++
//	c.ConsecutiveSuccesses++
//	c.ConsecutiveFailures = 0
//}
//
//func (c *Counts) onFailure() {
//	c.TotalFailures++
//	c.ConsecutiveFailures++
//	c.ConsecutiveSuccesses = 0
//}
//
//func (c *Counts) Clear() {
//	c.TotalSuccesses = 0
//	c.TotalFailures = 0
//	c.ConsecutiveSuccesses = 0
//	c.ConsecutiveFailures = 0
//	c.Requests = 0
//}
//
///*熔断器的配置设定*/
//type Settings struct {
//	Name          string                                  //名字
//	MaxRequests   uint32                                  //最大请求数
//	Interval      time.Duration                           //间隔时间 -> 与关闭有关
//	Timeout       time.Duration                           //超时时间 -> 与超时有关
//	ReadyToTrip   func(counts Counts) bool                //执行熔断
//	OnStateChange func(name string, from State, to State) //状态变更
//	IsSuccessful  func(err error) bool                    //是否成功
//}
//
///*断路器设置*/
//type CircuitBreaker struct {
//	name          string                                  //名字
//	maxRequests   uint32                                  //最大请求数 当连续请求成功数大于此时 断路器关闭
//	interval      time.Duration                           //间隔时间
//	timeout       time.Duration                           //超时时间
//	readyToTrip   func(counts Counts) bool                //是否执行熔断
//	isSuccessful  func(err error) bool                    //是否成功
//	onStateChange func(name string, from State, to State) //状态变更
//
//	mutex      sync.Mutex
//	state      State     //状态
//	generation uint64    //代 状态变更 new一个
//	counts     Counts    //数量
//	expiry     time.Time //到期时间 检查是否从开到半开
//}
//
//func NewCircuitBreaker(st Settings) *CircuitBreaker {
//
//	cb := new(CircuitBreaker)
//
//	cb.name = st.Name
//	cb.onStateChange = st.OnStateChange
//
//	if st.MaxRequests == 0 {
//		//最少得连续成功一次才熔断
//		cb.maxRequests = 1
//	} else {
//		cb.maxRequests = st.MaxRequests
//	}
//
//	if st.Interval <= 0 {
//		cb.interval = defaultInterval
//	} else {
//		cb.interval = st.Interval
//	}
//
//	if st.Timeout <= 0 {
//		//断路器由     开   ->   半开    的时间
//		cb.timeout = defaultTimeout
//	} else {
//		cb.timeout = st.Timeout
//	}
//
//	//是否需要熔断
//	if st.ReadyToTrip == nil {
//		cb.readyToTrip = defaultReadyToTrip
//	} else {
//		cb.readyToTrip = st.ReadyToTrip
//	}
//
//	//是否成功
//	if st.IsSuccessful == nil {
//		cb.isSuccessful = defaultIsSuccessful
//	} else {
//		cb.isSuccessful = st.IsSuccessful
//	}
//
//	cb.toNewGeneration(time.Now())
//
//	return cb
//
//}
//
//func (cb *CircuitBreaker) Excute(req func() (any, error)) (any, error) {
//	//请求之前 是否执行断路器
//	generation, err1 := cb.beforeReq()
//	if err1 != nil {
//		//有不符合的情况就中断
//		//log.P("对的出口出去了" + err1.Error())
//		return nil, err1
//	}
//
//	defer func() {
//		e := recover()
//		if e != nil {
//			cb.afterReq(generation, false)
//			panic(e)
//		}
//	}()
//
//	//代表一个请求
//	result, err := req()
//	//cb.counts.onRequest()
//
//	//请求之后 状态是否变更
//
//	cb.afterReq(generation, cb.isSuccessful(err))
//	return result, err
//}
//
//func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
//	cb.mutex.Lock()
//	defer cb.mutex.Unlock()
//	cb.generation++
//	cb.counts.Clear()
//	var zero time.Time
//	switch cb.state {
//	case StateClosed:
//		//间隔时间
//		if cb.interval == 0 {
//			cb.expiry = zero
//		} else {
//			cb.expiry = now.Add(cb.interval)
//		}
//	case StateHalfOpen:
//		cb.expiry = now.Add(cb.timeout)
//	case StateOpen:
//		cb.expiry = zero
//	}
//}
//
///*查看状态是否能够发起请求 是否被熔断    返回代数、错误*/
//func (cb *CircuitBreaker) beforeReq() (uint64, error) {
//	cb.mutex.Lock()
//	defer cb.mutex.Unlock()
//	//判断状态
//	now := time.Now()
//	state, generation := cb.currentState(now)
//
//	if state == StateOpen {
//		return generation, errors.New("当前断路器是打开状态")
//	}
//	if state == StateHalfOpen && (cb.counts.Requests > cb.maxRequests) {
//		//!!1
//		if cb.expiry.Before(now) {
//			cb.toNewGeneration(now)
//		}
//		return generation, errors.New("请求数量过多")
//	}
//
//	cb.counts.onRequest()
//	return generation, nil
//}
//
//func (cb *CircuitBreaker) afterReq(before uint64, success bool) error {
//	cb.mutex.Lock()
//	defer cb.mutex.Unlock()
//
//	now := time.Now()
//	state, generation := cb.currentState(now)
//	//意味 熔断器创出的 后面的请求在新的断路器验证、工作
//	if generation != before {
//		return nil
//	}
//	if success {
//		//是否需要 断路器 半开 -> 关 ？
//		cb.onSuccess(state, now)
//	} else {
//		//是否需要 断路器 半开 -> 开 ？
//		cb.onFailure(state, now)
//	}
//	return nil
//}
//
///*不修改状态，只是关着的熔断器过期后 开新一代断路器*/
//func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
//	switch cb.state {
//	case StateClosed:
//		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
//			cb.toNewGeneration(now)
//		}
//	case StateOpen:
//		if cb.expiry.Before(now) {
//			cb.setState(StateHalfOpen, now)
//		}
//
//	}
//	return cb.state, cb.generation
//}
//
//func (cb *CircuitBreaker) setState(state State, now time.Time) {
//	if cb.state == state {
//		return
//	}
//
//	prev := cb.state
//	cb.state = state
//	//状态变更之后，重新计数
//	cb.toNewGeneration(now)
//
//	if cb.onStateChange != nil {
//		cb.onStateChange(cb.name, prev, state)
//	}
//}
//
///*能产生请求的状态只有关和半开*/
//func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
//	switch state {
//	case StateClosed:
//		cb.counts.onSuccess()
//	case StateHalfOpen:
//		cb.counts.onSuccess()
//		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
//			cb.setState(StateClosed, now)
//		}
//	}
//}
//
///*能产生请求的状态只有关和半开 断路器 半开 -> 开 ？*/
//// 当连续失败大于阈值 则打开断路器，打开状态，当超过一段时间，设置为半开状态，设置为半开状态，当连续成功大于阈值，则设置为关闭状态。
//func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
//	switch state {
//	case StateClosed:
//		cb.counts.onFailure()
//		//是否开启熔断方法执行判断
//		if cb.readyToTrip(cb.counts) {
//			cb.setState(StateOpen, now)
//		}
//
//	case StateHalfOpen:
//		//状态变更 -> 清0 ->不用计数了
//		cb.setState(StateOpen, now)
//
//	}
//}
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrOpenState       = errors.New("circuit breaker is open")
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}

type Counts struct {
	Requests             uint32 //请求数量
	TotalSuccesses       uint32 //总成功数
	TotalFailures        uint32 //总失败数
	ConsecutiveSuccesses uint32 //连续成功数量
	ConsecutiveFailures  uint32 //连续失败数量
}

func (c *Counts) onRequest() {
	c.Requests++
}

func (c *Counts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) clear() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

type Settings struct {
	Name          string                                  //名字
	MaxRequests   uint32                                  //最大请求数
	Interval      time.Duration                           //间隔时间
	Timeout       time.Duration                           //超时时间
	ReadyToTrip   func(counts Counts) bool                //执行熔断
	OnStateChange func(name string, from State, to State) //状态变更
	IsSuccessful  func(err error) bool                    //是否成功
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	name          string                                  //名字
	maxRequests   uint32                                  //最大请求数 当连续请求成功数大于此时 断路器关闭
	interval      time.Duration                           //间隔时间
	timeout       time.Duration                           //超时时间
	readyToTrip   func(counts Counts) bool                //是否执行熔断
	isSuccessful  func(err error) bool                    //是否成功
	onStateChange func(name string, from State, to State) //状态变更

	mutex      sync.Mutex
	state      State     //状态
	generation uint64    //代 状态变更 new一个
	counts     Counts    //数量
	expiry     time.Time //到期时间 检查是否从开到半开
}

func NewCircuitBreaker(st Settings) *CircuitBreaker {
	cb := new(CircuitBreaker)

	cb.name = st.Name
	cb.onStateChange = st.OnStateChange

	if st.MaxRequests == 0 {
		cb.maxRequests = 1
	} else {
		cb.maxRequests = st.MaxRequests
	}

	if st.Interval <= 0 {
		cb.interval = defaultInterval
	} else {
		cb.interval = st.Interval
	}

	if st.Timeout <= 0 {
		cb.timeout = defaultTimeout
	} else {
		cb.timeout = st.Timeout
	}

	if st.ReadyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	} else {
		cb.readyToTrip = st.ReadyToTrip
	}

	if st.IsSuccessful == nil {
		cb.isSuccessful = defaultIsSuccessful
	} else {
		cb.isSuccessful = st.IsSuccessful
	}

	cb.toNewGeneration(time.Now())

	return cb
}

const defaultInterval = time.Duration(0) * time.Second
const defaultTimeout = time.Duration(60) * time.Second

// 连续失败五次 执行熔断
func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

func defaultIsSuccessful(err error) bool {
	return err == nil
}

// Name returns the name of the CircuitBreaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// State returns the current state of the CircuitBreaker.
func (cb *CircuitBreaker) State() State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns internal counters
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

func (cb *CircuitBreaker) Execute(req func() (any, error)) (any, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result, err := req()
	cb.afterRequest(generation, cb.isSuccessful(err))
	return result, err
}

func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrOpenState
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.onRequest()
	return generation, nil
}

func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.onSuccess()
	case StateHalfOpen:
		cb.counts.onSuccess()
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.onFailure()
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts.clear()

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}
