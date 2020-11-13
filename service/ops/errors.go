package ops

type ErrOpStopped struct{}

func (e *ErrOpStopped) Error() string {
	return "gate has 'stopped' status"
}
