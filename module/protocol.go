package module

type Header struct {
	Token      int64
	RemoteAddr string
	Proto      string
	Method     string
	Path       string
	Headers    []string
}

type Request struct {
	Header Header `json:"Header"`
	Body   []byte `json:"Body"`
}
