package utils

func Runnable(f func() error) {
	err := f()
	if err != nil {
		panic(err)
	}
}

func Supplier(f func() (interface{}, error)) interface{}  {
	r, err := f()
	if err != nil {
		panic(err)
	}
	return r
}

func Function(v interface{}, f func(v interface{}) (interface{}, error)) interface{} {
	r, err := f(v)
	if err != nil {
		panic(err)
	}
	return r
}

func BiFunction(v, t interface{}, f func(v, t interface{}) (interface{}, error)) interface{} {
	r, err := f(v, t)
	if err != nil {
		panic(err)
	}
	return r
}
