package zyerror

type MyError struct {
	err    error
	ErrFuc ErrFuc
}

func Default() *MyError {
	return &MyError{}
}

func (e *MyError) Error() string {
	return e.err.Error()
}

func (m *MyError) Put(err error) {
	m.check(err)
}

func (m *MyError) check(err error) error {
	if err != nil {
		m.err = err
		panic(m)
	}
	return nil
}

type ErrFuc func(err *MyError)

/* 用户自定义错误调用方法 */
func (m *MyError) Result(fuc ErrFuc) {
	m.ErrFuc = fuc
}

func (m *MyError) ExcuResult() {
	m.ErrFuc(m)
}
