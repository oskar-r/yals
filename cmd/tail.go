package cmd

type Tail interface {
	Start()
	Cleanup()
}

var tailImpl Tail

func SetupTail(t Tail) {
	tailImpl = t
}

func Start() {
	tailImpl.Start()
}

func Cleanup() {
	tailImpl.Cleanup()
}
