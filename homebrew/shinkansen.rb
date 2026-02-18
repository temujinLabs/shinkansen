class Shinkansen < Formula
  desc "Keyboard-driven TUI for Jira. Fast as a bullet train."
  homepage "https://github.com/temujinlabs/shinkansen"
  url "https://github.com/temujinlabs/shinkansen/archive/refs/tags/v0.1.0.tar.gz"
  sha256 ""
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/shinkansen"
  end

  test do
    assert_match "shinkansen #{version}", shell_output("#{bin}/shinkansen --version")
  end
end
