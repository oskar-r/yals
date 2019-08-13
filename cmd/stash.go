package cmd

type Stash interface {
	Send(message string) error
}

var stashImpl Stash

func SetupService(s Stash) {
	stashImpl = s
}

func Send(message string) error {
	return stashImpl.Send(message)
}
