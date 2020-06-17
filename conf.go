package main


type Conf struct {
}

func Init() *Conf {
	n := new(Conf)
	return n
}

func (n *Conf) Start() {
}

func (n *Conf) initDiscovery() {
}