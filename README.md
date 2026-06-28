# sterm

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](go.mod)

**English** | A terminal SSH connection manager with SFTP file transfer, encrypted password storage, and customizable themes.

**中文** | 终端 SSH 连接管理器，支持 SFTP 文件传输、密码加密存储和可定制主题。

---

## Features / 功能特性

| | English | 中文 |
|---|---------|------|
| 📋 | Host list with add / edit / delete | 主机增删改查 |
| 🔍 | Filter hosts by substring (`/`) | 子串过滤搜索（`/`） |
| 🔐 | Interactive SSH sessions | 交互式 SSH 连接 |
| 📁 | Dual-pane SFTP browser | 双栏 SFTP 文件浏览器 |
| 🔒 | AES-GCM encrypted passwords | 密码 AES-GCM 加密存储 |
| 🎨 | Built-in & custom themes | 内置与自定义主题 |

---

## Screenshot / 截图

> Add a terminal screenshot or GIF here before publishing.
>
> 发布前请在此添加终端截图或 GIF 演示。

---

## Installation / 安装

### From source / 源码构建

```bash
git clone git@github.com:ha1377311454/sterm.git
cd sterm
make build
./sterm
```

### Go install / Go 安装

```bash
go install github.com/ha1377311454/sterm@latest
```

### Cross-compile / 交叉编译

```bash
make release   # outputs to dist/
```

---

## Usage / 使用方法

```bash
sterm
```

### CLI flags / 命令行参数

| Flag | Description | 说明 |
|------|-------------|------|
| `--config-dir` | Config directory | 配置目录 |
| `--key-file` | AES encryption key file | AES 加密密钥文件 |
| `--theme-dir` | Custom theme directory (repeatable) | 自定义主题目录（可多次指定） |

---

## Keybindings / 快捷键

### Host list / 主机列表

| Key | Action | 操作 |
|-----|--------|------|
| `Enter` | Connect | 连接 |
| `a` | Add host | 添加 |
| `e` | Edit host | 编辑 |
| `d` | Delete host | 删除 |
| `f` | Open SFTP | 打开 SFTP |
| `/` | Filter | 过滤搜索 |
| `t` | Change theme | 切换主题 |
| `q` | Quit | 退出 |
| `Esc` | Clear filter | 清除过滤 |

### SFTP browser / SFTP 浏览器

| Key | Action | 操作 |
|-----|--------|------|
| `Tab` | Switch pane | 切换面板 |
| `Enter` | Enter directory | 进入目录 |
| `u` | Upload | 上传 |
| `d` | Download | 下载 |
| `g` | Go to path | 跳转路径 |
| `m` | Make directory | 新建目录 |
| `Esc` | Close | 关闭 |

---

## Configuration / 配置

### Config paths / 配置路径

| OS | Path |
|----|------|
| Linux | `~/.config/sterm/` |
| macOS | `~/Library/Application Support/sterm/` |
| Windows | `%APPDATA%\sterm\` |

Files / 文件:

- `config.yaml` — host definitions / 主机定义
- `key` — AES encryption key / AES 加密密钥
- `skins/` — custom themes / 自定义主题

### Example / 示例

See [`examples/config.yaml`](examples/config.yaml).

```yaml
theme: default
connections:
  - name: my-server
    host: 192.168.1.100
    port: 22
    user: root
    password: your-password
    key_path: ~/.ssh/id_rsa
    tags: [prod]
```

Passwords are encrypted with AES-GCM before being written to disk.

密码在写入磁盘前会使用 AES-GCM 加密。

### Built-in themes / 内置主题

`default`, `teal`, `nord`, `dracula`, `monokai`, `catppuccin`

Place custom `*.yaml` skin files in the config `skins/` directory or pass `--theme-dir`.

将自定义 `*.yaml` 主题文件放入配置目录的 `skins/` 下，或通过 `--theme-dir` 指定。

---

## Development / 开发

```bash
make test    # run tests / 运行测试
make lint    # golangci-lint (optional) / 代码检查
make tidy    # tidy modules / 整理依赖
make help    # list targets / 查看所有目标
```

---

## Security / 安全说明

- **Do not commit** `config.yaml` or `key` to version control.
- SSH **host key verification is not enabled yet**. See [SECURITY.md](SECURITY.md).

- **切勿**将 `config.yaml` 或 `key` 提交到版本库。
- SSH **主机密钥校验尚未启用**，详见 [SECURITY.md](SECURITY.md)。

---

## License / 许可证

[MIT License](LICENSE) © 2026 [ha1377311454](https://github.com/ha1377311454)
