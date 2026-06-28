//go:build windows

package ssh

import gossh "golang.org/x/crypto/ssh"

// watchResize 在 Windows 上为空实现（不存在 SIGWINCH）。
func watchResize(_ *gossh.Session, _ int) func() {
	return func() {}
}
