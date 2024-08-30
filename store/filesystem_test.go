package store_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/store"
)

const defaultFileMode = 0o755

func TestFileSystem_Write_and_Read(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T, path string) *store.FileSystem {
		t.Helper()

		factory := compose.NewFactory()
		logger := factory.GetLogger()

		path = filepath.Join("/tmp", path)
		err := os.MkdirAll(path, defaultFileMode)
		require.NoError(t, err)

		sut := store.NewFileSystem(logger, path)

		return sut
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		stubFileName := "bar"
		stubData := []byte("bar")

		// setup
		sut := setup(t, gofakeit.UUID())
		ctx := context.Background()

		// execute
		err := sut.Write(ctx, stubFileName, stubData)
		require.NoError(t, err)

		data, err := sut.Read(ctx, stubFileName)
		require.NoError(t, err)

		// assert
		require.Equal(t, stubData, data)
	})
}
