//go:build linux

package os

import "os/exec"

func newPlatformIsolator() OSIsolator {
	ll := NewLandlockIsolator()
	sc := NewSeccompIsolator()

	var active []OSIsolator
	if ll.Available() {
		active = append(active, ll)
	}
	if sc.Available() {
		active = append(active, sc)
	}

	switch len(active) {
	case 0:
		return &noopIsolator{}
	case 1:
		return active[0]
	default:
		return &compositeIsolator{isolators: active}
	}
}

func probePlatform(caps *PlatformCapabilities) {
	ll := NewLandlockIsolator()
	caps.HasLandlock = ll.Available()
	if li, ok := ll.(*landlockIsolator); ok {
		caps.LandlockABI = li.abiVersion
	}

	sc := NewSeccompIsolator()
	caps.HasSeccomp = sc.Available()

	out, err := exec.Command("uname", "-r").Output()
	if err == nil && len(out) > 0 {
		caps.KernelVersion = string(out[:len(out)-1])
	}
}
