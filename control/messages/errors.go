package messages

type ErrInvalidResponse struct{}

func (e *ErrInvalidResponse) Error() string {
	return "invalid response"
}
