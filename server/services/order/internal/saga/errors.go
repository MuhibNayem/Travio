package saga

// SagaError for client errors
type SagaError struct {
	Message string
}

func (e *SagaError) Error() string {
	return e.Message
}
