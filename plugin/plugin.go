package plugin

type Plugin interface {
	Name() string

	Init()

	Destroy()
}

