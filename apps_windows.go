// +build !linux !darwin

package main

import (
	"os/user"
	"syscall"
)

func sysProcAttrs(u *user.User) *syscall.SysProcAttr {
	return nil
}
