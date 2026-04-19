package p2p

import (
	"reflect"
	"testing"
)

func TestBuildIdentityView_PreservesDidAndListenAddrs(t *testing.T) {
	listenAddrs := []string{
		"/ip4/127.0.0.1/tcp/9000",
		"/ip6/::/tcp/9000",
	}

	view := buildIdentityView("did:lango:v2:abcdef1234567890", "peer-123", "secure-store", listenAddrs)

	if got, want := view["did"], "did:lango:v2:abcdef1234567890"; got != want {
		t.Fatalf("did = %v, want %v", got, want)
	}
	if got, want := view["peerId"], "peer-123"; got != want {
		t.Fatalf("peerId = %v, want %v", got, want)
	}
	if got, want := view["keyStorage"], "secure-store"; got != want {
		t.Fatalf("keyStorage = %v, want %v", got, want)
	}

	gotAddrs, ok := view["listenAddrs"].([]string)
	if !ok {
		t.Fatalf("listenAddrs type = %T, want []string", view["listenAddrs"])
	}
	if !reflect.DeepEqual(gotAddrs, listenAddrs) {
		t.Fatalf("listenAddrs = %v, want %v", gotAddrs, listenAddrs)
	}
}
