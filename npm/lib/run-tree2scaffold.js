"use strict";

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const SUPPORTED_TARGETS = new Map([
  ["darwin-x64", path.join("native", "darwin-x64", "tree2scaffold")],
  ["darwin-arm64", path.join("native", "darwin-arm64", "tree2scaffold")],
  ["linux-x64", path.join("native", "linux-x64", "tree2scaffold")],
]);

function packageRoot() {
  return path.resolve(__dirname, "..");
}

function nativeBinaryPath(platform = process.platform, arch = process.arch) {
  const target = `${platform}-${arch}`;
  const relPath = SUPPORTED_TARGETS.get(target);
  if (!relPath) {
    const supported = Array.from(SUPPORTED_TARGETS.keys()).join(", ");
    throw new Error(`unsupported platform ${target}; supported: ${supported}`);
  }
  return path.join(packageRoot(), relPath);
}

function run(commandName) {
  const command = commandName || "tree2scaffold";

  let binary;
  try {
    binary = nativeBinaryPath();
  } catch (err) {
    console.error(`${command}: ${err.message}`);
    process.exit(1);
  }

  if (!fs.existsSync(binary)) {
    console.error(
      `${command}: missing native binary at ${binary}; run package validation or reinstall the package`
    );
    process.exit(1);
  }

  const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
  if (result.error) {
    console.error(`${command}: failed to execute ${binary}: ${result.error.message}`);
    process.exit(1);
  }
  process.exit(result.status === null ? 1 : result.status);
}

module.exports = { nativeBinaryPath, run };
