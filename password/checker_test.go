package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/password"
)

func TestChecker_IsOK(t *testing.T) {
	t.Parallel()

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
		wantErr string
	}{
		{
			name: "empty password",
			fields: fields{
				minimumEntropy: 0.1,
			},
			args:    args{password: ""},
			wantErr: "insecure password",
		},
		{
			name: "weak password, low bar",
			fields: fields{
				minimumEntropy: 20.0,
			},
			args:    args{password: "helloWorld"},
			wantErr: "",
		},
		{
			name: "weak password, normal bar",
			fields: fields{
				minimumEntropy: 60.0,
			},
			args:    args{password: "helloWorld"},
			wantErr: "insecure password",
		},
		{
			name: "medium password, normal bar",
			fields: fields{
				minimumEntropy: 60.0,
			},
			args:    args{password: "helloWorld123"},
			wantErr: "",
		},
		{
			name: "medium password, high bar",
			fields: fields{
				minimumEntropy: 100.0,
			},
			args:    args{password: "helloWorld123"},
			wantErr: "insecure password",
		},
		{
			name: "high password, high bar",
			fields: fields{
				minimumEntropy: 100.0,
			},
			args:    args{password: "6LRFjZse6IYiBNGZlhrVEckQqt9i"},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sut := password.NewCheckerWithEntropy(tt.fields.minimumEntropy)

			err := sut.IsOK(tt.args.password)

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
