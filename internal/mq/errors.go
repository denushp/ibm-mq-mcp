package mq

import "fmt"

const (
	CategoryUnavailable       = "unavailable"
	CategoryConnectionFailure = "mq_connection_failure"
	CategoryAuthorization     = "authorization_failure"
	CategoryObjectNotFound    = "object_not_found"
	CategoryPCFCommandFailure = "pcf_command_failure"
	CategoryBrowseFailure     = "browse_failure"
	CategoryPutFailure        = "put_failure"
)

type OperationError struct {
	Category string
	CompCode int32
	Reason   int32
	Detail   string
}

func (e *OperationError) Error() string {
	if e.CompCode == 0 && e.Reason == 0 {
		return fmt.Sprintf("%s: %s", e.Category, e.Detail)
	}
	return fmt.Sprintf("%s: %s (CC=%d RC=%d)", e.Category, e.Detail, e.CompCode, e.Reason)
}

func unavailableError(detail string) error {
	return &OperationError{
		Category: CategoryUnavailable,
		Detail:   detail,
	}
}
