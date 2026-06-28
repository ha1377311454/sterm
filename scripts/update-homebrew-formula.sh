#!/usr/bin/env bash
# 根据 GitHub Release 的 checksums.txt 生成 Homebrew Formula。
# 用法: ./scripts/update-homebrew-formula.sh v0.1.0 [checksums.txt]
set -euo pipefail

TAG="${1:?用法: $0 <tag> [checksums.txt]}"
VERSION="${TAG#v}"
CHECKSUMS_FILE="${2:-}"

REPO="ha1377311454/sterm"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

if [[ -z "$CHECKSUMS_FILE" ]]; then
  CHECKSUMS_FILE="$(mktemp)"
  curl -fsSL "${BASE_URL}/checksums.txt" -o "$CHECKSUMS_FILE"
  trap 'rm -f "$CHECKSUMS_FILE"' EXIT
fi

sha() {
  awk -v name="$1" '$2 == name { print $1; exit }' "$CHECKSUMS_FILE"
}

SHA_DARWIN_ARM64="$(sha sterm_darwin_arm64)"
SHA_DARWIN_AMD64="$(sha sterm_darwin_amd64)"
SHA_LINUX_ARM64="$(sha sterm_linux_arm64)"
SHA_LINUX_AMD64="$(sha sterm_linux_amd64)"
BREW_BIN='#{bin}'

cat <<RUBY
class Sterm < Formula
  desc "Terminal SSH connection manager with SFTP and themes"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  on_macos do
    on_arm do
      url "${BASE_URL}/sterm_darwin_arm64"
      sha256 "${SHA_DARWIN_ARM64}"
    end
    on_intel do
      url "${BASE_URL}/sterm_darwin_amd64"
      sha256 "${SHA_DARWIN_AMD64}"
    end
  end

  on_linux do
    on_arm do
      url "${BASE_URL}/sterm_linux_arm64"
      sha256 "${SHA_LINUX_ARM64}"
    end
    on_intel do
      url "${BASE_URL}/sterm_linux_amd64"
      sha256 "${SHA_LINUX_AMD64}"
    end
  end

  def install
    if OS.mac?
      bin.install (Hardware::CPU.arm? ? "sterm_darwin_arm64" : "sterm_darwin_amd64") => "sterm"
    elsif OS.linux?
      bin.install (Hardware::CPU.arm? ? "sterm_linux_arm64" : "sterm_linux_amd64") => "sterm"
    end
  end

  test do
    assert_match "SSH connection manager", shell_output("${BREW_BIN}/sterm --help", 2)
  end
end
RUBY
