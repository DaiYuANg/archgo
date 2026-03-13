package clientx

import "testing"

func TestHookFuncsDispatch(t *testing.T) {
	var dialCalled bool
	var ioCalled bool

	h := HookFuncs{
		OnDialFunc: func(event DialEvent) {
			dialCalled = event.Protocol == ProtocolTCP
		},
		OnIOFunc: func(event IOEvent) {
			ioCalled = event.Protocol == ProtocolHTTP
		},
	}

	EmitDial([]Hook{h}, DialEvent{Protocol: ProtocolTCP})
	EmitIO([]Hook{h}, IOEvent{Protocol: ProtocolHTTP})

	if !dialCalled {
		t.Fatal("expected dial hook to be called")
	}
	if !ioCalled {
		t.Fatal("expected io hook to be called")
	}
}

func TestEmitHookPanicIsolation(t *testing.T) {
	dialCalled := false
	ioCalled := false

	hooks := []Hook{
		HookFuncs{
			OnDialFunc: func(event DialEvent) {
				panic("dial hook panic")
			},
			OnIOFunc: func(event IOEvent) {
				panic("io hook panic")
			},
		},
		HookFuncs{
			OnDialFunc: func(event DialEvent) {
				dialCalled = true
			},
			OnIOFunc: func(event IOEvent) {
				ioCalled = true
			},
		},
	}

	EmitDial(hooks, DialEvent{Protocol: ProtocolTCP})
	EmitIO(hooks, IOEvent{Protocol: ProtocolHTTP})

	if !dialCalled {
		t.Fatal("expected subsequent dial hook to be called after panic")
	}
	if !ioCalled {
		t.Fatal("expected subsequent io hook to be called after panic")
	}
}
