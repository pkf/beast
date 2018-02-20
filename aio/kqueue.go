// +build darwin dragonfly freebsd netbsd openbsd solaris

package aio

// #include <sys/types.h>
// #include <sys/event.h>
// #include <sys/time.h>
//
// struct timespec make_timespec(long sec, long nsec) {
//	struct timespec ts = {
//	    .tv_sec = sec,
//	    .tv_nsec = nsec,
//	};
//	return ts;
// }
import "C"

import (
	"syscall"
	"time"
	"unsafe"
)

type Poller int

func NewPoller() (Poller, error) {
	fd, err := C.kqueue()
	if err != nil {
		return 0, err
	}
	return Poller(fd), nil
}

func (p Poller) Add(fd int, flags Flags) error {
	var ev C.struct_kevent
	ev.ident = C.uintptr_t(fd)
	ev.flags = C.EV_ADD
	if flags&In != 0 {
		ev.filter |= C.EVFILT_READ
	}
	if flags&Out != 0 {
		ev.filter |= C.EVFILT_WRITE
	}
	if flags&Err != 0 {
		//todo
	}
	return p.applyEvent(&ev)
}

func (p Poller) Delete(fd int) error {
	var ev C.struct_kevent
	ev.ident = C.uintptr_t(fd)
	ev.filter = C.EVFILT_READ | C.EVFILT_WRITE
	ev.flags = C.EV_DELETE
	return p.applyEvent(&ev)
}

func (p Poller) Wait(duration time.Duration) ([]Event, error) {
	const maxEvents = 64
	inEvents := make([]C.struct_kevent, maxEvents)
	n, err := C.kevent(C.int(p), nil, 0, (*C.struct_kevent)(unsafe.Pointer(&inEvents[0])), maxEvents, p.timespec(duration))
	if err != nil {
		if err == syscall.EINTR {
			err = nil
		}
		return nil, err
	}
	events := make([]Event, int(n))
	for ii := 0; ii < int(n); ii++ {
		inEvent := inEvents[ii]
		var flags Flags
		if inEvent.filter&C.EVFILT_READ != 0 {
			flags |= In
		}
		if inEvent.filter&C.EVFILT_WRITE != 0 {
			flags |= Out
		}
		if inEvent.flags&C.EV_ERROR != 0 {
			flags |= Err
		}
		events[ii] = Event{
			Fd:    int(inEvent.ident),
			Flags: flags,
		}
	}
	return events, nil
}

func (p Poller) timespec(d time.Duration) *C.struct_timespec {
	if d < 0 {
		// a NULL timespec tells kqueue to wait forever
		return nil
	}
	t := syscall.NsecToTimespec(d.Nanoseconds())
	// darwin defines tv_sec as __darwin_time_t, so we
	// need to to some indirection to let C do the implicit
	// type conversion.
	ts := C.make_timespec(C.long(t.Sec), C.long(t.Nsec))
	return &ts
}

func (p Poller) applyEvent(ev *C.struct_kevent) error {
	ok, err := C.kevent(C.int(p), ev, 1, nil, 0, p.timespec(0))
	if ok < 0 {
		return err
	}
	return nil
}
