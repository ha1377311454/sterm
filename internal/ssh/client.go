package ssh

import (
	"fmt"
	"os"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// ConnectOptions 保存 SSH 连接的参数。
type ConnectOptions struct {
	Host     string
	Port     int
	User     string
	Password string
	KeyPath  string
}

// Connect 打开交互式 SSH 会话，阻塞直到用户退出。
// 调用方应在 tview.Application.Suspend 内调用，以便暂停 TUI 并使终端处于可用状态。
func Connect(opts ConnectOptions) error {
	client, err := dial(opts, 15*time.Second)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}
		defer term.Restore(fd, oldState)
	}

	w, h, _ := term.GetSize(fd)
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 38400,
		gossh.TTY_OP_OSPEED: 38400,
	}
	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		return fmt.Errorf("request pty: %w", err)
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}

	// 将终端尺寸变化事件转发给远程 shell（平台相关实现）。
	stopResize := watchResize(session, fd)
	defer stopResize()

	return session.Wait()
}

// dial 使用密码和/或密钥认证建立 SSH 客户端连接。
func dial(opts ConnectOptions, timeout time.Duration) (*gossh.Client, error) {
	var auth []gossh.AuthMethod
	var keyErr error

	if opts.KeyPath != "" {
		if signer, err := loadKey(opts.KeyPath); err == nil {
			auth = append(auth, gossh.PublicKeys(signer))
		} else {
			keyErr = fmt.Errorf("load key %q: %w", opts.KeyPath, err)
		}
	}
	if opts.Password != "" {
		auth = append(auth, gossh.Password(opts.Password))
		// 作为需要 keyboard-interactive 的服务器的回退认证方式
		auth = append(auth, gossh.KeyboardInteractive(
			func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = opts.Password
				}
				return answers, nil
			},
		))
	}
	if len(auth) == 0 {
		if keyErr != nil {
			return nil, fmt.Errorf("no usable SSH auth method: %w", keyErr)
		}
		return nil, fmt.Errorf("no SSH auth method configured: set password or key path")
	}

	cfg := &gossh.ClientConfig{
		User:            opts.User,
		Auth:            auth,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         timeout,
	}
	client, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port), cfg)
	if err != nil && keyErr != nil {
		return nil, fmt.Errorf("%w; key auth unavailable: %v", err, keyErr)
	}
	return client, err
}

func loadKey(path string) (gossh.Signer, error) {
	if len(path) > 1 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = home + path[1:]
	}
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return gossh.ParsePrivateKey(key)
}
