// Copyright 2015 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package prog

import (
	"testing"
)

func TestResourceCtors(t *testing.T) {
	for _, c := range Syscalls {
		for _, res := range c.InputResources() {
			if len(calcResourceCtors(res.Desc.Kind, true)) == 0 {
				t.Errorf("call %v requires input resource %v, but there are no calls that can create this resource", c.Name, res.Desc.Name)
			}
		}
	}
}

func TestTransitivelyEnabledCalls(t *testing.T) {
	t.Parallel()
	calls := make(map[*Syscall]bool)
	for _, c := range Syscalls {
		calls[c] = true
	}
	if trans := TransitivelyEnabledCalls(calls); len(calls) != len(trans) {
		for c := range calls {
			if !trans[c] {
				t.Logf("disabled %v", c.Name)
			}
		}
		t.Fatalf("can't create some resource")
	}
	delete(calls, SyscallMap["epoll_create"])
	if trans := TransitivelyEnabledCalls(calls); len(calls) != len(trans) {
		t.Fatalf("still must be able to create epoll fd with epoll_create1")
	}
	delete(calls, SyscallMap["epoll_create1"])
	trans := TransitivelyEnabledCalls(calls)
	if len(calls)-5 != len(trans) ||
		trans[SyscallMap["epoll_ctl$EPOLL_CTL_ADD"]] ||
		trans[SyscallMap["epoll_ctl$EPOLL_CTL_MOD"]] ||
		trans[SyscallMap["epoll_ctl$EPOLL_CTL_DEL"]] ||
		trans[SyscallMap["epoll_wait"]] ||
		trans[SyscallMap["epoll_pwait"]] {
		t.Fatalf("epoll fd is not disabled")
	}
}

func TestClockGettime(t *testing.T) {
	t.Parallel()
	calls := make(map[*Syscall]bool)
	for _, c := range Syscalls {
		calls[c] = true
	}
	// Removal of clock_gettime should disable all calls that accept timespec/timeval.
	delete(calls, SyscallMap["clock_gettime"])
	trans := TransitivelyEnabledCalls(calls)
	if len(trans)+10 > len(calls) {
		t.Fatalf("clock_gettime did not disable enough calls: before %v, after %v", len(calls), len(trans))
	}
}