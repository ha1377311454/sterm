package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Host 表示一条已保存的 SSH 连接。
type Host struct {
	Name        string   `yaml:"name"`
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	User        string   `yaml:"user"`
	Password    string   `yaml:"password"`
	KeyPath     string   `yaml:"key_path,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// Options 控制配置、加密密钥和自定义主题目录的位置。
type Options struct {
	ConfigDir string
	KeyFile   string
	ThemeDirs []string
}

func (h Host) TagsStr() string {
	return strings.Join(h.Tags, ", ")
}

// Config 保存全部应用设置。
type Config struct {
	Connections []Host `yaml:"connections"`
	Theme       string `yaml:"theme"`
	filePath    string `yaml:"-"`
	keyFile     string `yaml:"-"`
}

// ConfigDir 返回平台默认的 sterm 配置目录。
func ConfigDir() string {
	return defaultConfigDir()
}

// SkinsDir 返回 skins 子目录。
func SkinsDir() string {
	return filepath.Join(ConfigDir(), "skins")
}

// New 返回带默认值的 Config。
func New() *Config {
	return NewWithOptions(Options{})
}

// NewWithOptions 返回带默认值及路径覆盖选项的 Config。
func NewWithOptions(opts Options) *Config {
	configDir := cleanOrDefault(opts.ConfigDir, defaultConfigDir())
	keyFile := cleanOrDefault(opts.KeyFile, filepath.Join(configDir, "key"))
	return &Config{
		Theme:    "default",
		filePath: filepath.Join(configDir, "config.yaml"),
		keyFile:  keyFile,
	}
}

// Load 从磁盘读取配置，文件不存在时返回默认值。
func Load() (*Config, error) {
	return LoadWithOptions(Options{})
}

// LoadWithOptions 使用显式运行时选项从磁盘读取配置。
func LoadWithOptions(opts Options) (*Config, error) {
	c := NewWithOptions(opts)
	data, err := os.ReadFile(c.filePath)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	for i := range c.Connections {
		if c.Connections[i].Port == 0 {
			c.Connections[i].Port = 22
		}
		pass, err := decryptPassword(c.keyFile, c.Connections[i].Password)
		if err != nil {
			return nil, err
		}
		c.Connections[i].Password = pass
	}
	return c, nil
}

// Save 将配置写入磁盘。
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.filePath), 0o700); err != nil {
		return err
	}
	store := *c
	store.Connections = append([]Host(nil), c.Connections...)
	for i := range store.Connections {
		pass, err := encryptPassword(c.keyFile, store.Connections[i].Password)
		if err != nil {
			return err
		}
		store.Connections[i].Password = pass
	}
	data, err := yaml.Marshal(&store)
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath, data, 0o600)
}

func (c *Config) AddHost(h Host) {
	if h.Port == 0 {
		h.Port = 22
	}
	c.Connections = append(c.Connections, h)
}

func (c *Config) UpdateHost(idx int, h Host) {
	if idx >= 0 && idx < len(c.Connections) {
		if h.Port == 0 {
			h.Port = 22
		}
		c.Connections[idx] = h
	}
}

func (c *Config) DeleteHost(idx int) {
	if idx >= 0 && idx < len(c.Connections) {
		c.Connections = append(c.Connections[:idx], c.Connections[idx+1:]...)
	}
}

func defaultConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "sterm")
	}
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "sterm")
		}
	case "darwin":
		if home != "" {
			return filepath.Join(home, "Library", "Application Support", "sterm")
		}
	}
	if home != "" {
		return filepath.Join(home, ".config", "sterm")
	}
	return "."
}

func cleanOrDefault(path, fallback string) string {
	if strings.TrimSpace(path) == "" {
		return fallback
	}
	return filepath.Clean(path)
}
