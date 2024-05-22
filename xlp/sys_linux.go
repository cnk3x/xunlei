package xlp

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/cnk3x/xunlei/embeds"
)

func setupProcAttr(c *exec.Cmd, uid, gid uint32) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Pdeathsig = syscall.SIGKILL

	if uid != 0 || gid != 0 {
		c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
	}

	return
}

func setupWrapProcAttr(c *exec.Cmd, root string) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Pdeathsig = syscall.SIGINT
	c.SysProcAttr.Chroot = root
	return
}

func checkEnv() (err error) {
	if stat, _ := os.Stat("/.dockerenv"); stat != nil {
		return os.Remove("/.dockerenv")
	}
	return embeds.Extract(SYNOPKG_PKGDEST)
}

func lookupUg(userOrId, groupOrId string) (uid, gid uint32, err error) {
	var u *user.User
	var g *user.Group

	if userOrId != "" {
		u, err = user.Lookup(userOrId)

		if err != nil {
			u, err = user.LookupId(userOrId)
		}

		if err != nil {
			uid, err = parseUint(userOrId)
		}

		if err != nil {
			return
		}
	}

	if groupOrId != "" {
		g, err = user.LookupGroup(groupOrId)

		if err != nil {
			g, err = user.LookupGroupId(groupOrId)
		}

		if err != nil {
			gid, err = parseUint(groupOrId)
		}

		if err != nil {
			return
		}
	}

	if u != nil {
		uid, _ = parseUint(u.Uid)
		gid, _ = parseUint(u.Gid)
	}

	if g != nil {
		gid, _ = parseUint(u.Gid)
	}

	return
}

func parseUint(s string) (uint32, error) {
	if out, err := strconv.ParseUint(s, 10, 32); err != nil {
		return 0, err
	} else {
		return uint32(out), nil
	}
}

func sysMount(source, endpoint string) (err error) {
	return syscall.Mount(source, endpoint, "auto", syscall.MS_BIND|syscall.MS_SLAVE|syscall.MS_REC, "")
}

func sysUnmount(endpoint string) error {
	return syscall.Unmount(endpoint, syscall.MNT_DETACH)
}

func sysChroot(newRoot string) (err error) {
	return syscall.Chroot(newRoot)
}

var _ = setupWrapProcAttr
