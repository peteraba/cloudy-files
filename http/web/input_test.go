package web_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/http/inandout"
	"github.com/peteraba/cloudy-files/http/web"
)

func header(data map[string][]string) http.Header {
	h := http.Header{}

	for k, v := range data {
		h.Set(k, v[0])
	}

	return h
}

func TestGetIPAddress(t *testing.T) {
	t.Parallel()

	ipStub := "34.241.31.225"
	ipStub2 := "228.182.151.174"

	tests := []struct {
		name    string
		request *http.Request
		want    string
	}{
		{
			name:    "empty",
			request: &http.Request{},
			want:    "",
		},
		{
			name: "remote addr",
			request: &http.Request{
				RemoteAddr: ipStub,
			},
			want: ipStub,
		},
		{
			name: "x-forwarded-for",
			request: &http.Request{
				Header: header(map[string][]string{
					inandout.HeaderXForwardedFor: {ipStub},
				}),
				RemoteAddr: ipStub2,
			},
			want: ipStub,
		},
		{
			name: "x-real-ip",
			request: &http.Request{
				Header: header(map[string][]string{
					inandout.HeaderXForwardedFor: {ipStub2},
					inandout.HeaderXRealIP:       {ipStub},
				}),
				RemoteAddr: ipStub2,
			},
			want: ipStub,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// execute
			ip := web.GetIPAddress(tt.request)

			// assert
			assert.Equal(t, tt.want, ip)
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	type data struct {
		Name string `json:"name"`
	}

	t.Run("empty form", func(t *testing.T) {
		t.Parallel()

		// setup
		req := &http.Request{
			Method: http.MethodPut,
		}

		// execute
		d, err := web.Parse(req, data{})
		require.Error(t, err)

		// assert
		assert.Empty(t, d)
		assert.ErrorContains(t, err, "failed to parse form")
	})

	t.Run("missing content type", func(t *testing.T) {
		t.Parallel()

		// setup
		type data struct {
			Name string `formam:"name"`
		}

		req := &http.Request{
			Header: header(map[string][]string{
				"name": {"John"},
			}),
			Method: http.MethodPut,
			Body:   io.NopCloser(strings.NewReader(`name=John`)),
		}

		// execute
		d, err := web.Parse(req, data{})
		require.Error(t, err)

		// assert
		assert.Empty(t, d)
		assert.ErrorContains(t, err, "content type")
	})
}
