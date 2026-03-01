package customerror

import (
	"github.com/morikuni/failure"
)

var (
	ItemNotFound      = failure.StringCode("ItemNotFound")
	FailedSendMessage = failure.StringCode("FailedSendMessage")
	DocDBError        = failure.StringCode("DocDBError")
)

type Error struct {
	StatusCode int
	Message    string
}

func HandleError(err error) *Error {
	if err == nil {
		return nil
	}

	code, ok := failure.CodeOf(err)
	if !ok {
		return &Error{
			StatusCode: 1,
			Message:    err.Error(),
		}
	}
	switch code {
	case ItemNotFound:
		return &Error{
			StatusCode: 1,
			Message:    "the requested item was not found",
		}
	default:
		return &Error{
			StatusCode: 2,
			Message:    err.Error(),
		}
	}
}
