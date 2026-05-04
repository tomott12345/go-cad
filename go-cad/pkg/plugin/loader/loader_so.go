//go:build !windows && !js && cgo

package loader

import (
	goplugin "plugin"

	"fmt"

	"go-cad/pkg/plugin"
)

// LoadSO loads a Go plugin .so file and looks up the exported `NewPlugin`
// symbol, which must have the signature `func() plugin.Plugin`.
//
// This function is only available on Linux/macOS with CGO enabled.
func LoadSO(path string) (plugin.Plugin, error) {
	p, err := goplugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("LoadSO: open %s: %w", path, err)
	}
	sym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("LoadSO: %s does not export NewPlugin: %w", path, err)
	}
	newFn, ok := sym.(func() plugin.Plugin)
	if !ok {
		return nil, fmt.Errorf("LoadSO: %s: NewPlugin has wrong signature (expected func() plugin.Plugin)", path)
	}
	return newFn(), nil
}
