package clientx

import (
	"time"
)

type Hook interface {
	OnDial(event DialEvent)
	OnIO(event IOEvent)
}

type HookFuncs struct {
	OnDialFunc func(event DialEvent)
	OnIOFunc   func(event IOEvent)
}

func (h HookFuncs) OnDial(event DialEvent) {
	if h.OnDialFunc != nil {
		h.OnDialFunc(event)
	}
}

func (h HookFuncs) OnIO(event IOEvent) {
	if h.OnIOFunc != nil {
		h.OnIOFunc(event)
	}
}

type DialEvent struct {
	Protocol Protocol
	Op       string
	Network  string
	Addr     string
	Duration time.Duration
	Err      error
}

type IOEvent struct {
	Protocol Protocol
	Op       string
	Addr     string
	Bytes    int
	Duration time.Duration
	Err      error
}

func EmitDial(hooks []Hook, event DialEvent) {
	for _, h := range hooks {
		emitDialSafe(h, event)
	}
}

func EmitIO(hooks []Hook, event IOEvent) {
	for _, h := range hooks {
		emitIOSafe(h, event)
	}
}

func emitDialSafe(h Hook, event DialEvent) {
	if h == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	h.OnDial(event)
}

func emitIOSafe(h Hook, event IOEvent) {
	if h == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	h.OnIO(event)
}
