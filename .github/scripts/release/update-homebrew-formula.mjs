import fs from "node:fs";

const version = process.argv[2];
if (!version) {
  throw new Error("Usage: update-homebrew-formula.mjs <version>");
}

const { assets } = JSON.parse(fs.readFileSync(0, "utf8"));
const sha256 = (name) => {
  const digest = assets.find((asset) => asset.name === name)?.digest;
  if (!digest?.startsWith("sha256:")) {
    throw new Error(`Missing SHA-256 digest for ${name}`);
  }
  return digest.slice("sha256:".length);
};

const macosArm = sha256("nasduck-aarch64-apple-darwin.tar.gz");
const macosIntel = sha256("nasduck-x86_64-apple-darwin.tar.gz");
const linuxArm = sha256("nasduck-aarch64-unknown-linux-gnu.tar.gz");
const linuxIntel = sha256("nasduck-x86_64-unknown-linux-gnu.tar.gz");

const formula = `class Nasduck < Formula
  desc "Control, inspect, test, and record terminal sessions"
  homepage "https://github.com/shiplaone2q/nasduck"
  version "${version}"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/shiplaone2q/nasduck/releases/download/${version}/nasduck-aarch64-apple-darwin.tar.gz"
      sha256 "${macosArm}"
    end

    on_intel do
      url "https://github.com/shiplaone2q/nasduck/releases/download/${version}/nasduck-x86_64-apple-darwin.tar.gz"
      sha256 "${macosIntel}"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/shiplaone2q/nasduck/releases/download/${version}/nasduck-aarch64-unknown-linux-gnu.tar.gz"
      sha256 "${linuxArm}"
    end

    on_intel do
      url "https://github.com/shiplaone2q/nasduck/releases/download/${version}/nasduck-x86_64-unknown-linux-gnu.tar.gz"
      sha256 "${linuxIntel}"
    end
  end

  def install
    bin.install "nasduck"
  end

  test do
    system bin/"nasduck", "--version"
  end
end
`;

fs.writeFileSync("Formula/nasduck.rb", formula);
