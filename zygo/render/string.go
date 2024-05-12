package render

import (
	"fmt"
	"net/http"
	"web/zygo/internal/bytesconv"
)

type String struct {
	Format string
	Data   []any
}

func (s *String) Render(w http.ResponseWriter) error {
	s.WriteContentType(w)
	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(w, s.Format, s.Data...)
		return err
	}
	_, err := w.Write(bytesconv.StringToBytes(s.Format))
	return err
}

func (s *String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/plain;charset=utf-8")
}
