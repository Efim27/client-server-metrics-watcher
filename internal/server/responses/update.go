package responses

type UpdateMetricResponse struct {
	DefaultResponse
	Hash string `json:"hash,omitempty"`
}

func NewUpdateMetricResponse() UpdateMetricResponse {
	response := UpdateMetricResponse{}
	response.Status = StatusOk

	return response
}

func (response *UpdateMetricResponse) SetHash(hash string) *UpdateMetricResponse {
	if hash == "" {
		return response
	}

	response.Hash = hash
	return response
}
