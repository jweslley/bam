package main

import (
	"os/user"
	"strconv"
	"syscall"
)

func sysProcAttrs(u *user.User) *syscall.SysProcAttr {
	var c *syscall.Credential
	if u != nil {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		c = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}
	return &syscall.SysProcAttr{
		Setpgid:    true,
		Pdeathsig:  syscall.SIGKILL,
		Credential: c,
	}
}
