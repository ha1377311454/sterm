# Homebrew 维护说明

## 用户安装

```bash
brew install ha1377311454/tap/sterm
```

Tap 仓库：https://github.com/ha1377311454/homebrew-tap

## 手动更新 Formula

发新版本后，在 sterm 仓库根目录执行：

```bash
./scripts/update-homebrew-formula.sh v0.2.0 > /tmp/sterm.rb
# 复制到 homebrew-tap/Formula/sterm.rb 并提交
```

## 自动更新（Release 工作流）

在 sterm 仓库 **Settings → Secrets → Actions** 添加：

| Secret | 说明 |
|--------|------|
| `HOMEBREW_TAP_TOKEN` | GitHub PAT，需 `repo` 权限，用于 push 到 `homebrew-tap` |

每次 push `v*` tag 触发 Release 后，`bump-homebrew` job 会自动更新 tap 仓库中的 formula。

未配置该 Secret 时，Release 仍正常完成，仅跳过 tap 更新。
