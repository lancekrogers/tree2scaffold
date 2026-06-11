// Package env abstracts the host-environment probes that are not portable
// across build targets. Under GOOS=wasip1 there is no process model, so the
// exec-based probes (clipboard, `go version`, `git config`) are unavailable and
// report ErrUnsupported; callers MUST fall back to sensible defaults rather than
// treat that as a hard error. The implementation is selected by build tags:
// env_exec.go for native builds, env_wasip1.go for WASI.
package env

import "errors"

// ErrUnsupported signals a host probe that cannot run in the current build or
// runtime (e.g. an exec-based probe under GOOS=wasip1). Treat it as a soft
// failure and fall back to a default.
var ErrUnsupported = errors.New("env: operation unsupported in this runtime")

// Environment abstracts the non-portable host probes used while scaffolding.
// It is intentionally small so each build target can implement it trivially.
type Environment interface {
	// GoVersion returns the host Go toolchain's major.minor (e.g. "1.24"), or
	// ("", ErrUnsupported) when no toolchain probe is available.
	GoVersion() (string, error)

	// GitRemoteOriginURL returns the working directory repo's origin URL, or
	// ("", ErrUnsupported) when VCS probing is unavailable.
	GitRemoteOriginURL() (string, error)

	// Getwd returns the current working directory. Portable on native AND wasip1.
	Getwd() (string, error)

	// Clipboard returns the clipboard contents, or (nil, ErrUnsupported) where a
	// clipboard is unavailable (e.g. under WASI).
	Clipboard() ([]byte, error)
}
