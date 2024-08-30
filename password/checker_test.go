package password_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/util"
)

func TestChecker_IsOK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("fail on password too long", func(t *testing.T) {
		t.Parallel()

		// setup
		sut := password.NewChecker()

		// execute
		err := sut.IsOK(ctx, strings.Repeat("a", 73))
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrPasswordTooLong)
	})

	type fields struct {
		minimumEntropy float64
	}
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty password",
			fields: fields{
				minimumEntropy: 0.1,
			},
			args:    args{password: ""},
			wantErr: util.ErrorContains("insecure password"),
		},
		{
			name: "weak password, low bar",
			fields: fields{
				minimumEntropy: 20.0,
			},
			args:    args{password: "helloWorld"},
			wantErr: assert.NoError,
		},
		{
			name: "weak password, normal bar",
			fields: fields{
				minimumEntropy: 60.0,
			},
			args:    args{password: "helloWorld"},
			wantErr: util.ErrorContains("insecure password"),
		},
		{
			name: "medium password, normal bar",
			fields: fields{
				minimumEntropy: 60.0,
			},
			args:    args{password: "helloWorld123"},
			wantErr: assert.NoError,
		},
		{
			name: "medium password, high bar",
			fields: fields{
				minimumEntropy: 100.0,
			},
			args:    args{password: "helloWorld123"},
			wantErr: util.ErrorContains("insecure password"),
		},
		{
			name: "high password, high bar",
			fields: fields{
				minimumEntropy: 100.0,
			},
			args:    args{password: "6LRFjZse6IYiBNGZlhrVEckQqt9i"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			sut := password.NewCheckerWithEntropy(tt.fields.minimumEntropy)

			// execute
			actual := sut.IsOK(ctx, tt.args.password)

			tt.wantErr(t, actual, fmt.Sprintf("IsOK(%v)", tt.args.password))
		})
	}
}
