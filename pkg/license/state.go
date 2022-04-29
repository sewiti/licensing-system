package license

type State int32

const (
	StateInvalid State = iota
	StateValid
	StateExpired
	StateClosed
)

func (s State) String() string {
	switch s {
	case StateValid:
		return "valid"
	case StateExpired:
		return "expired"
	case StateClosed:
		return "closed"
	default:
		// case StateInvalid:
		return "invalid"
	}
}
