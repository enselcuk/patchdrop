//go:build !windows

package discovery

import "syscall"

// reuseControl lets several PatchDrop instances (or a restart) bind the
// discovery port at once; broadcast datagrams are delivered to all of them.
func reuseControl(network, address string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		if serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); serr != nil {
			return
		}
		serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	})
	if err != nil {
		return err
	}
	return serr
}
