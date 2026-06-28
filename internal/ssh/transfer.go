package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

// SFTPClient 封装 SSH 与 SFTP 连接。
type SFTPClient struct {
	ssh  *gossh.Client
	sftp *sftp.Client
}

// ProgressFunc 接收已传输字节数、总字节数和已用时间。
type ProgressFunc func(done, total int64, elapsed time.Duration)

// NewSFTPClient 与指定主机建立 SFTP 会话。
func NewSFTPClient(opts ConnectOptions) (*SFTPClient, error) {
	client, err := dial(opts, 15*time.Second)
	if err != nil {
		return nil, err
	}
	sc, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}
	return &SFTPClient{ssh: client, sftp: sc}, nil
}

// Close 关闭 SFTP 与 SSH 连接。
func (c *SFTPClient) Close() {
	if c.sftp != nil {
		c.sftp.Close()
	}
	if c.ssh != nil {
		c.ssh.Close()
	}
}

// RemoteFile 保存远程文件的元数据。
type RemoteFile struct {
	Name    string
	Size    int64
	IsDir   bool
	ModTime time.Time
	Mode    os.FileMode
}

// ListDir 返回指定远程路径下的目录项。
func (c *SFTPClient) ListDir(path string) ([]RemoteFile, error) {
	entries, err := c.sftp.ReadDir(path)
	if err != nil {
		return nil, err
	}
	files := make([]RemoteFile, 0, len(entries)+1)
	if path != "/" {
		files = append(files, RemoteFile{Name: "..", IsDir: true})
	}
	for _, e := range entries {
		files = append(files, RemoteFile{
			Name:    e.Name(),
			Size:    e.Size(),
			IsDir:   e.IsDir(),
			ModTime: e.ModTime(),
			Mode:    e.Mode(),
		})
	}
	return files, nil
}

// Download 将远程文件复制到 localDir。
func (c *SFTPClient) Download(remotePath, localDir string) error {
	return c.DownloadWithProgress(remotePath, localDir, nil)
}

// DownloadWithProgress 将远程文件复制到 localDir 并报告进度。
func (c *SFTPClient) DownloadWithProgress(remotePath, localDir string, progress ProgressFunc) error {
	name := filepath.Base(remotePath)
	dst, err := os.Create(filepath.Join(localDir, name))
	if err != nil {
		return err
	}
	defer dst.Close()

	src, err := c.sftp.Open(remotePath)
	if err != nil {
		return err
	}
	defer src.Close()

	var total int64
	if info, statErr := src.Stat(); statErr == nil {
		total = info.Size()
	}
	_, err = copyWithProgress(dst, src, total, progress)
	return err
}

// Upload 将本地文件复制到 remoteDir。
func (c *SFTPClient) Upload(localPath, remoteDir string) error {
	return c.UploadWithProgress(localPath, remoteDir, nil)
}

// UploadWithProgress 将本地文件复制到 remoteDir 并报告进度。
func (c *SFTPClient) UploadWithProgress(localPath, remoteDir string, progress ProgressFunc) error {
	name := filepath.Base(localPath)
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	var total int64
	if info, statErr := src.Stat(); statErr == nil {
		total = info.Size()
	}
	dst, err := c.sftp.Create(path.Join(remoteDir, name))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = copyWithProgress(dst, src, total, progress)
	return err
}

// MkdirAll 创建远程目录路径，包括缺失的父目录。
func (c *SFTPClient) MkdirAll(remotePath string) error {
	return c.sftp.MkdirAll(remotePath)
}

// WorkingDir 返回远程工作目录。
func (c *SFTPClient) WorkingDir() (string, error) {
	return c.sftp.Getwd()
}

func copyWithProgress(dst io.Writer, src io.Reader, total int64, progress ProgressFunc) (int64, error) {
	buf := make([]byte, 32*1024)
	start := time.Now()
	var done int64
	if progress != nil {
		progress(0, total, 0)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				done += int64(nw)
				if progress != nil {
					progress(done, total, time.Since(start))
				}
			}
			if ew != nil {
				return done, ew
			}
			if nr != nw {
				return done, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				if progress != nil {
					progress(done, total, time.Since(start))
				}
				return done, nil
			}
			return done, er
		}
	}
}

// FormatSize 返回人类可读的文件大小。
func FormatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(1024), 0
	for n := size / 1024; n >= 1024; n /= 1024 {
		div *= 1024
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
