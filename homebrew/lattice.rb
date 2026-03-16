class Lattice < Formula
  desc "Mental models engine — apply Munger's latticework to any problem"
  homepage "https://github.com/cyperx84/lattice"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/cyperx84/lattice/releases/download/v#{version}/lattice-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/cyperx84/lattice/releases/download/v#{version}/lattice-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    url "https://github.com/cyperx84/lattice/releases/download/v#{version}/lattice-linux-amd64.tar.gz"
    sha256 "PLACEHOLDER"
  end

  def install
    bin.install "lattice"
  end

  test do
    assert_match "lattice", shell_output("#{bin}/lattice --version")
  end
end
