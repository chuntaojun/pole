package utils

type Equals interface {
	Equal(i interface{}) bool
}

type Comparator interface {
	Compare(i interface{}) int
}
