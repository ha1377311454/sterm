//go:build !windows

package ssh

import (
	"os"
	"os/signal"
	"syscall"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// watchResize 监听 SIGWINCH 信号，并将终端窗口尺寸变化转发给 SSH 会话。
// 返回 stop 函数，调用方必须在 defer 中调用。
func watchResize(session *gossh.Session, fd int) func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if w, h, err := term.GetSize(fd); err == nil {
				_ = session.WindowChange(h, w)
			}
		}
	}()
	return func() {
		signal.Stop(ch)
		close(ch)
	}
}
