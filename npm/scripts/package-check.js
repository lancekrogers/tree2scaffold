#!/usr/bin/env node
"use strict";

const fs = require("fs");
const os = require("os");
const path = require("path");
const { spawnSync } = require("child_process");
const { nativeBinaryPath } = require("../lib/run-tree2scaffold");

const root = path.resolve(__dirname, "..", "..");
const pkg = require(path.join(root, "package.json"));

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: root,
    encoding: "utf8",
    ...options,
  });
  if (result.status !== 0) {
    throw new Error(
      `${command} ${args.join(" ")} failed\nstdout:\n${result.stdout || ""}\nstderr:\n${result.stderr || ""}`
    );
  }
  return result;
}

function assertPackageShape() {
  const expectedBins = {
    tree2scaffold: "npm/bin/tree2scaffold.js",
    t2s: "npm/bin/t2s.js",
  };

  for (const [name, relPath] of Object.entries(expectedBins)) {
    if (pkg.bin?.[name] !== relPath) {
      throw new Error(`package.json bin.${name} must point to ${relPath}`);
    }

    const filePath = path.join(root, relPath);
    const content = fs.readFileSync(filePath, "utf8");
    if (!content.startsWith("#!/usr/bin/env node")) {
      throw new Error(`${relPath} must start with #!/usr/bin/env node`);
    }
  }

  if (!Array.isArray(pkg.files) || !pkg.files.includes("npm/")) {
    throw new Error('package.json files must include "npm/"');
  }
}

function stageCurrentBinary() {
  const binary = nativeBinaryPath();
  fs.mkdirSync(path.dirname(binary), { recursive: true });
  run("go", [
    "build",
    "-ldflags",
    "-X main.version=npm-check -X main.commit=packagecheck -X main.date=2026-07-01T00:00:00Z",
    "-o",
    binary,
    "./cmd/tree2scaffold",
  ]);
  fs.chmodSync(binary, 0o755);
  return binary;
}

function assertWrapperSmoke() {
  for (const wrapper of ["tree2scaffold", "t2s"]) {
    const relPath = pkg.bin[wrapper];
    const result = run(process.execPath, [relPath, "--version"]);
    if (!result.stdout.includes("tree2scaffold npm-check")) {
      throw new Error(`${wrapper} --version did not reach staged native binary`);
    }
  }

  const rootDir = fs.mkdtempSync(path.join(os.tmpdir(), "tree2scaffold-npm-"));
  try {
    run(process.execPath, [pkg.bin.tree2scaffold, "-root", rootDir, "-yes"], {
      input: "demo/\n└── README.md\n",
    });
    if (!fs.existsSync(path.join(rootDir, "README.md"))) {
      throw new Error("wrapper pipe smoke test did not create README.md");
    }
  } finally {
    fs.rmSync(rootDir, { recursive: true, force: true });
  }
}

function assertUnsupportedPlatform() {
  try {
    nativeBinaryPath("freebsd", "x64");
    throw new Error("unsupported platform did not throw");
  } catch (err) {
    if (!err.message.includes("unsupported platform freebsd-x64")) {
      throw err;
    }
  }
}

function assertPackContents() {
  const result = run("npm", ["pack", "--dry-run", "--json"]);
  const pack = JSON.parse(result.stdout)[0];
  const files = pack.files.map((file) => file.path).sort();

  const required = [
    "LICENSE",
    "README.md",
    "npm/bin/t2s.js",
    "npm/bin/tree2scaffold.js",
    "npm/lib/run-tree2scaffold.js",
    "npm/scripts/package-check.js",
    "package.json",
  ];

  for (const file of required) {
    if (!files.includes(file)) {
      throw new Error(`npm pack output is missing ${file}`);
    }
  }

  if (!files.some((file) => file.startsWith("npm/native/"))) {
    throw new Error("npm pack output is missing staged native binary");
  }

  const forbiddenPrefixes = [".git/", "dist/", "test/", "specs/", "ai_docs/"];
  for (const file of files) {
    if (forbiddenPrefixes.some((prefix) => file.startsWith(prefix))) {
      throw new Error(`npm pack output includes forbidden file ${file}`);
    }
  }
}

assertPackageShape();
stageCurrentBinary();
try {
  assertWrapperSmoke();
  assertUnsupportedPlatform();
  assertPackContents();
} finally {
  fs.rmSync(path.join(root, "npm", "native"), { recursive: true, force: true });
}

console.log("npm package validation passed");
