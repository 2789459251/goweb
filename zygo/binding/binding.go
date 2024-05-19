package binding

import "net/http"

type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

var JSON = JSONBinding{}
var XML = XMLBinding{}
