class Sterm < Formula
  desc "Terminal SSH connection manager with SFTP and themes"
  homepage "https://github.com/ha1377311454/sterm"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/ha1377311454/sterm/releases/download/v0.1.0/sterm_darwin_arm64"
      sha256 "23602d6016e0eef432e3f68540cd311b28606f479382ef7b70bedb5818452349"
    end
    on_intel do
      url "https://github.com/ha1377311454/sterm/releases/download/v0.1.0/sterm_darwin_amd64"
      sha256 "81cf8bbd9f52d01e5f2e82b66b2c4d6c61b53bb7d7faeff96734840ce2933426"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/ha1377311454/sterm/releases/download/v0.1.0/sterm_linux_arm64"
      sha256 "bba37d5b210ce0a5e34e309bee25a3d2603bfe86a0cad774a4dcab860396ab00"
    end
    on_intel do
      url "https://github.com/ha1377311454/sterm/releases/download/v0.1.0/sterm_linux_amd64"
      sha256 "2ff05d3f7f6a2fe2d5ce6f6ae94a40bc2894325a955ccc4c01af7a1177822d2d"
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
    assert_match "SSH connection manager", shell_output("#{bin}/sterm --help", 2)
  end
end
