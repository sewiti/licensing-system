package license

import (
	"fmt"
	"syscall"
)

// Identifier returns info about system.
// It is equivalent to:
//  echo "$(uname -n) | $(uname -srm) | $(uname -v)"
func Identifier() (string, error) {
	var un syscall.Utsname
	err := syscall.Uname(&un)
	if err != nil {
		return "", err
	}
	n := int8ToStr(un.Nodename[:])
	s := int8ToStr(un.Sysname[:])
	r := int8ToStr(un.Release[:])
	m := int8ToStr(un.Machine[:])
	v := int8ToStr(un.Version[:])
	return fmt.Sprintf("%s | %s %s %s | %s", n, s, r, m, v), nil
}

func int8ToStr(bs []int8) string {
	if len(bs) == 0 {
		return ""
	}
	buf := make([]byte, len(bs))
	for i, v := range bs {
		if v == 0 {
			return string(buf[:i])
		}
		buf[i] = byte(v)
	}
	return string(buf)
}
