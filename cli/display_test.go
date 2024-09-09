package cli_test

import (
	"testing"

	"github.com/peteraba/cloudy-files/cli"
)

func TestStdout_Println(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		sut := cli.NewStdout()

		sut.Println("msg:", "Hello, World!")
	})
}
