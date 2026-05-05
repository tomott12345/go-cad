package dxf

import (
	"io"
	"strings"

	"github.com/tomott12345/go-cad/internal/document"
)

// Write exports doc as a DXF R2000 (AC1015) stream to w.
func Write(doc *document.Document, w io.Writer) error {
	_, err := io.WriteString(w, doc.ExportDXF())
	return err
}

// WriteR12 exports doc as a DXF R12 (AC1009) stream to w.
func WriteR12(doc *document.Document, w io.Writer) error {
	_, err := io.WriteString(w, doc.ExportDXFR12())
	return err
}

// String exports doc as a DXF R2000 string (convenience wrapper).
func String(doc *document.Document) string { return doc.ExportDXF() }

// StringR12 exports doc as a DXF R12 string (convenience wrapper).
func StringR12(doc *document.Document) string { return doc.ExportDXFR12() }

// ReadString is a convenience wrapper around Read that accepts a string.
func ReadString(dxfText string) (*document.Document, []string, error) {
	return Read(strings.NewReader(dxfText))
}
