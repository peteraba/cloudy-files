package apperr_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/apperr"
)

func TestGetProblem(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want *apperr.Problem
	}{
		{
			name: "assert.AnError",
			args: args{
				err: assert.AnError,
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "Assert.AnError General Error For Testing.",
			},
		},
		{
			name: "access denied",
			args: args{
				err: apperr.ErrAccessDenied,
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Access denied",
				Status: http.StatusForbidden,
				Detail: "Access Denied.",
			},
		},
		{
			name: "not found",
			args: args{
				err: apperr.ErrNotFound,
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Not found",
				Status: http.StatusNotFound,
				Detail: "Not Found.",
			},
		},
		{
			name: "not implemented",
			args: args{
				err: apperr.ErrNotImplemented,
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Not implemented",
				Status: http.StatusNotFound,
				Detail: "Not Implemented.",
			},
		},
		{
			name: "bad request",
			args: args{
				err: apperr.ErrBadRequest(assert.AnError),
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Bad request",
				Status: http.StatusBadRequest,
				Detail: "Assert.AnError General Error For Testing, Err.",
			},
		},
		{
			name: "validation",
			args: args{
				err: apperr.ErrValidation(assert.AnError.Error()),
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Bad request",
				Status: http.StatusBadRequest,
				Detail: "Assert.AnError General Error For Testing, Err.",
			},
		},
		{
			name: "not implemented, wrapped",
			args: args{
				err: fmt.Errorf("foo, err: %w", apperr.ErrNotImplemented),
			},
			want: &apperr.Problem{
				Type:   "",
				Title:  "Not implemented",
				Status: http.StatusNotFound,
				Detail: "Foo, Err.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// execute
			got := apperr.GetProblem(tt.args.err)

			// assert
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProblem_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		problem *apperr.Problem
		want    string
	}{
		{
			name:    "simple",
			problem: apperr.GetProblem(apperr.ErrAccessDenied),
			want:    "403 Access denied Access Denied.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.problem.Error())
		})
	}
}
