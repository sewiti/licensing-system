package license

type State int32

const (
	StateInvalid State = iota
	StateValid
	StateExpired
	StateClosed
)
