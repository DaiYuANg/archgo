package dix

import (
	"errors"
	"strings"

	do "github.com/samber/do/v2"
)

type StopReport struct {
	HookError      error
	ShutdownReport *do.ShutdownReport
}

func (r *StopReport) HasErrors() bool {
	return r != nil && r.Err() != nil
}

func (r *StopReport) Err() error {
	if r == nil {
		return nil
	}

	errs := make([]error, 0, 2)
	if r.HookError != nil {
		errs = append(errs, r.HookError)
	}
	if r.ShutdownReport != nil && len(r.ShutdownReport.Errors) > 0 {
		errs = append(errs, r.ShutdownReport)
	}
	return errors.Join(errs...)
}

func (r *StopReport) Error() string {
	if err := r.Err(); err != nil {
		parts := make([]string, 0, 2)
		if r.HookError != nil {
			parts = append(parts, r.HookError.Error())
		}
		if r.ShutdownReport != nil && len(r.ShutdownReport.Errors) > 0 {
			parts = append(parts, r.ShutdownReport.Error())
		}
		return strings.Join(parts, "\n")
	}
	return ""
}
