//go:build wasip1

package env

import "os"

// wasiEnv is the WASI Environment. WASI Preview 1 has no process model, so the
// exec-based probes report ErrUnsupported; the working directory still works via
// the runtime's preopened directories. This file imports no os/exec, so the
// wasip1 binary never links process-spawning code.
type wasiEnv struct{}

// New returns the WASI Environment.
func New() Environment { return wasiEnv{} }

func (wasiEnv) GoVersion() (string, error)          { return "", ErrUnsupported }
func (wasiEnv) GitRemoteOriginURL() (string, error) { return "", ErrUnsupported }
func (wasiEnv) Getwd() (string, error)              { return os.Getwd() }
func (wasiEnv) Clipboard() ([]byte, error)          { return nil, ErrUnsupported }
