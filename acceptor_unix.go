// Copyright 2019 Andy Pan. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build linux freebsd dragonfly darwin

package gnet

import (
	"os"

	"golang.org/x/sys/unix"
)

func (svr *server) acceptNewConnection(fd int) error {
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return os.NewSyscallError("accept", err)
	}
	if err := unix.SetNonblock(nfd, true); err != nil {
		return os.NewSyscallError("fcntl nonblock", err)
	}
	el := svr.subEventLoopSet.next(nfd)
	c := newTCPConn(nfd, el, sa)
	_ = el.poller.Trigger(func() (err error) {
		if err = el.poller.AddRead(nfd); err != nil {
			return
		}
		el.connections[nfd] = c
		el.calibrateCallback(el, 1)
		err = el.loopOpen(c)
		return
	})
	return nil
}
