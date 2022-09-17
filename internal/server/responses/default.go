package responses

import "encoding/json"

type DefaultResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func NewDefaultResponse() DefaultResponse {
	response := DefaultResponse{}
	response.Status = StatusOk

	return response
}

func (response *DefaultResponse) SetStatus(newStatus string) *DefaultResponse {
	response.Status = newStatus
	return response
}

func (response *DefaultResponse) SetStatusError(responseError error) *DefaultResponse {
	response.Status = StatusError
	response.Error = responseError.Error()
	return response
}

func (response DefaultResponse) GetJSONBytes() []byte {
	jsonBytes, _ := json.Marshal(response)
	return jsonBytes
}

func (response DefaultResponse) GetJSONString() string {
	return string(response.GetJSONBytes())
}
