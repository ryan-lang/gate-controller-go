package control

type ErrUnsolicitedFrame struct{}

func (e *ErrUnsolicitedFrame) Error() string {
	return "unsolicited frame"
}

type ErrUnexpectedFrame struct{}

func (e *ErrUnexpectedFrame) Error() string {
	return "packet did not pass transaction filter"
}

type ErrNotRunning struct{}

func (e *ErrNotRunning) Error() string {
	return "controller is not running"
}
