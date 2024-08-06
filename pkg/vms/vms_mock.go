//go:build !linux

package vms

import (
	"context"
)

func (d *Vm) unmount(ctx context.Context)                   {}
func (d *Vm) mount(ctx context.Context) (err error)         { return }
func chroot(ctx context.Context, target string) (err error) { return }
