package model

import (
	"bytes"
	"errors"
	"strconv"
)

type Limit int

const Unlimited Limit = 0 // Less than or equals means unlimited.

// Allows reports whether v is allowed, i.e., doesn't exceed the limit.
func (l Limit) Allows(v int) bool {
	if l <= Unlimited {
		return true
	}
	return v <= int(l)
}

func (l Limit) MarshalJSON() ([]byte, error) {
	if l <= Unlimited {
		return []byte("-1"), nil
	}
	return []byte(strconv.Itoa(int(l))), nil
}

func (l *Limit) UnmarshalJSON(bs []byte) error {
	if l == nil {
		return errors.New("limit is nil")
	}
	if bytes.Equal(bs, []byte("null")) {
		*l = Unlimited
		return nil
	}
	v, err := strconv.Atoi(string(bs))
	if err != nil {
		return err
	}
	if v < int(Unlimited) {
		v = int(Unlimited)
	}
	*l = Limit(v)
	return nil
}
