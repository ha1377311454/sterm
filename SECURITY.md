# Security Policy / 安全策略

## Reporting a Vulnerability / 报告漏洞

If you discover a security issue, please **do not** open a public GitHub issue.
Email the maintainer privately or use GitHub **Security Advisories** on this repository.

如发现安全问题，**请勿**公开提交 Issue，请通过私有渠道或 GitHub Security Advisories 联系维护者。

## Sensitive Data / 敏感数据

- `config.yaml` may contain SSH credentials. **Never commit it to git.**
- The AES key file (`key`) encrypts stored passwords locally. Keep it private with file mode `0600`.
- Config directory is created with mode `0700`.

- `config.yaml` 可能包含 SSH 凭据，**切勿**提交到 Git。
- 本地 AES 密钥文件（`key`）用于加密存储的密码，请保持私有，权限应为 `0600`。
- 配置目录以 `0700` 权限创建。

## Known Limitations / 已知限制

- SSH host key verification is **not enabled** (`InsecureIgnoreHostKey`). Connections are vulnerable to MITM attacks. Use only on trusted networks until host key checking is implemented.

- SSH 主机密钥校验**尚未启用**，连接可能遭受中间人攻击。在实现主机密钥校验前，请仅在可信网络中使用。
