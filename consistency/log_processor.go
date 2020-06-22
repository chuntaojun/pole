package consistency

type LogProcessor interface {

	OnRequest()

	OnApply()

	OnError()

	Group() string

}

type LogProcessor4AP interface {

	LogProcessor

}

type LogProcessor4CP interface {

	LogProcessor



}

