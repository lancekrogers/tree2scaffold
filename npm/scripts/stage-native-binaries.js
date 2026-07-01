#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const root = path.resolve(__dirname, "..", "..");
const releaseDir = path.join(root, "dist", "release");
const nativeDir = path.join(root, "npm", "native");

const targets = [
  { archive: "darwin_amd64", npm: "darwin-x64" },
  { archive: "darwin_arm64", npm: "darwin-arm64" },
  { archive: "linux_amd64", npm: "linux-x64" },
];

function run(command, args) {
  const result = spawnSync(command, args, { cwd: root, encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(
      `${command} ${args.join(" ")} failed\nstdout:\n${result.stdout || ""}\nstderr:\n${result.stderr || ""}`
    );
  }
  return result;
}

function findArchive(target) {
  if (!fs.existsSync(releaseDir)) {
    throw new Error(`release directory does not exist: ${releaseDir}`);
  }

  const suffix = `_${target.archive}.tar.gz`;
  const matches = fs.readdirSync(releaseDir).filter((name) => name.endsWith(suffix));
  if (matches.length !== 1) {
    throw new Error(`expected exactly one release archive ending in ${suffix}, found ${matches.length}`);
  }
  return path.join(releaseDir, matches[0]);
}

fs.rmSync(nativeDir, { recursive: true, force: true });

for (const target of targets) {
  const archive = findArchive(target);
  const destination = path.join(nativeDir, target.npm);
  fs.mkdirSync(destination, { recursive: true });
  run("tar", ["-xzf", archive, "-C", destination, "tree2scaffold"]);
  fs.chmodSync(path.join(destination, "tree2scaffold"), 0o755);
}

console.log(`staged native binaries in ${nativeDir}`);
