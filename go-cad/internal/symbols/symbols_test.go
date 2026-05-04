package symbols

import (
        "testing"

        "go-cad/internal/document"
)

func TestNamesNotEmpty(t *testing.T) {
        n := Names()
        if len(n) == 0 {
                t.Error("expected at least one built-in symbol")
        }
        for _, name := range n {
                if name == "" {
                        t.Error("found empty symbol name in Names()")
                }
        }
}

func TestKnownSymbolsPresent(t *testing.T) {
        expected := []string{
                "CENTER_MARK",
                "NORTH_ARROW",
                "REVISION_TRIANGLE",
                "DATUM_TRIANGLE",
                "SURFACE_FINISH",
        }
        set := make(map[string]bool)
        for _, n := range Names() {
                set[n] = true
        }
        for _, want := range expected {
                if !set[want] {
                        t.Errorf("expected symbol %q in Names(), not found", want)
                }
        }
}

func TestEntitiesReturnsNonEmpty(t *testing.T) {
        for _, name := range Names() {
                ents := Entities(name)
                if len(ents) == 0 {
                        t.Errorf("symbol %q has no entities", name)
                }
        }
}

func TestEntitiesUnknownReturnsNil(t *testing.T) {
        ents := Entities("NO_SUCH_SYMBOL_XYZ")
        if ents != nil && len(ents) != 0 {
                t.Errorf("expected nil for unknown symbol, got %v", ents)
        }
}

func TestNoDuplicateNames(t *testing.T) {
        seen := make(map[string]int)
        for _, n := range Names() {
                seen[n]++
                if seen[n] > 1 {
                        t.Errorf("duplicate symbol name: %q", n)
                }
        }
}

func TestRegisterInstallsAllSymbols(t *testing.T) {
        doc := document.New()
        Register(doc)
        blocks := doc.Blocks()
        registered := make(map[string]bool)
        for _, b := range blocks {
                registered[b.Name] = true
        }
        for _, name := range Names() {
                if !registered[name] {
                        t.Errorf("Register() did not install symbol %q into document", name)
                }
        }
}

func TestRegisterIdempotent(t *testing.T) {
        doc := document.New()
        Register(doc)
        n1 := len(doc.Blocks())
        Register(doc)
        n2 := len(doc.Blocks())
        if n1 != n2 {
                t.Errorf("Register() is not idempotent: %d blocks after first call, %d after second", n1, n2)
        }
}

func TestEntitiesHaveValidTypes(t *testing.T) {
        valid := map[string]bool{
                document.TypeLine:     true,
                document.TypeCircle:   true,
                document.TypeArc:      true,
                document.TypePolyline: true,
                document.TypeText:     true,
        }
        for _, name := range Names() {
                for i, e := range Entities(name) {
                        if !valid[e.Type] {
                                t.Errorf("symbol %q entity %d has unexpected type %q", name, i, e.Type)
                        }
                }
        }
}
