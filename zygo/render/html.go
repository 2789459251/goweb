package render

import (
	"html/template"
	"net/http"
	"web/zygo/internal/bytesconv"
)

type HTML struct {
	Data       any
	Name       string
	Templete   *template.Template
	IsTemplate bool
}
type HTMLRender struct {
	Template *template.Template
}

func (h *HTML) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	if h.IsTemplate {
		return h.Templete.ExecuteTemplate(w, h.Name, h.Data)
	}
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}

func (h *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html;charset=utf-8")
}
