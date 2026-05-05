//go:build windows || js || !cgo

package loader

import (
	"fmt"

	"github.com/tomott12345/go-cad/pkg/plugin"
)

// LoadSO is not available on this platform.
// Use the subprocess JSON-RPC transport instead.
func LoadSO(path string) (plugin.Plugin, error) {
	return nil, fmt.Errorf("LoadSO: .so plugin loading is not supported on this platform (use subprocess transport)")
}
