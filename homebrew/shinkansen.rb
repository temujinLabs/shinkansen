class Shinkansen < Formula
  desc "Keyboard-driven TUI for Jira. Fast as a bullet train."
  homepage "https://shinkansen.temujinlabs.com"
  url "https://github.com/temujinLabs/lab/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "4ad0b92e969013c4ffa53c7df26c6452971c6098a7526c35d5bbb93b7e9fe05f"
  license "MIT"

  depends_on "go" => :build

  def install
    cd "shinkansen" do
      ldflags = %W[
        -s -w
        -X main.version=#{version}
      ]
      system "go", "build", *std_go_args(ldflags:), "./cmd/shinkansen"
    end
  end

  test do
    assert_match "shinkansen #{version}", shell_output("#{bin}/shinkansen --version")
  end
end
