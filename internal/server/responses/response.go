// Package responses - шаблон ответов сервера в формате JSON
package responses

const (
	StatusOk    = "ok"
	StatusError = "error"
)

type Response interface {
	GetJSONBytes() []byte
	GetJSONString() string
}
