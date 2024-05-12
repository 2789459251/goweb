package render

import (
	"encoding/xml"
	"net/http"
)

type Xml struct {
	Data any
}

func (x *Xml) Render(w http.ResponseWriter) error {
	x.WriteContentType(w)
	err := xml.NewEncoder(w).Encode(x.Data)
	return err
}

func (x *Xml) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/xml;charset=utf-8")
}
