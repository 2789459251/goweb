package binding

import (
	"encoding/xml"
	"net/http"
)

type XMLBinding struct {
}

func (b *XMLBinding) Name() string {
	return "xml"
}
func (b *XMLBinding) Bind(req *http.Request, obj any) error {
	if req.Body == nil {
		return nil
	}
	decode := xml.NewDecoder(req.Body)
	err := decode.Decode(obj)
	if err != nil {
		return err
	}

	return Validator_.ValidateStruct(obj)
}
