package aio

type Flags int

const (
	In  = 1
	Out = 2
	Err = 4
)

type Event struct {
	Fd    int
	Flags Flags
}
