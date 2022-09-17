package responses

const (
	StatusOk    = "ok"
	StatusError = "error"
)

type Response interface {
	GetJSONBytes() []byte
	GetJSONString() string
}
