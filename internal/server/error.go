package server

import (
	"errors"

	"github.com/apex/log"
	"github.com/sewiti/licensing-system/internal/core"
)

func logError(err error, scope string) {
	sErr := &core.SensitiveError{}
	if errors.As(err, &sErr) {
		log.WithError(sErr.Err).Errorf("%s: %s", scope, sErr.Message)
	} else {
		log.WithError(err).Error(scope)
	}
}
