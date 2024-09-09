package web_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/http"
	"github.com/peteraba/cloudy-files/util"
)

func TestNegotiateContentType(t *testing.T) {
	t.Parallel()

	type args struct {
		accept         string
		supportedTypes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty accept returns first supported type",
			args: args{
				accept: "",
				supportedTypes: []string{
					http.ContentTypePlain,
				},
			},
			want: http.ContentTypePlain,
		},
		{
			name: "return first found accepted type which is supported type",
			args: args{
				accept: "dummy, text/plain; charset=utf-8, text/html; charset=utf-8",
				supportedTypes: []string{
					http.ContentTypeHTML,
					http.ContentTypeJSON,
				},
			},
			want: http.ContentTypeHTML,
		},
		{
			name: "return first supported type in case nothing supported is matched",
			args: args{
				accept: "dummy, text/plai; charset=utf-8, text/htm; charset=utf-8",
				supportedTypes: []string{
					http.ContentTypeHTML,
					http.ContentTypeJSON,
				},
			},
			want: http.ContentTypeHTML,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			actual := util.NegotiateContentType(tt.args.accept, tt.args.supportedTypes)

			// assert
			assert.Equal(t, tt.want, actual)
		})
	}
}
