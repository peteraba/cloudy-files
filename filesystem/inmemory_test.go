package filesystem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/filesystem"
	"github.com/peteraba/cloudy-files/util"
)

func TestInMemory_Write_and_Read(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) *filesystem.InMemory {
		t.Helper()

		sut := filesystem.NewInMemory(util.NewSpy())

		return sut
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		dataStub := []byte("bar")
		sut := setup(t)

		// exercise
		err := sut.Write(ctx, nameStub, dataStub)
		require.NoError(t, err)

		actualData, err := sut.Read(ctx, nameStub)
		require.NoError(t, err)

		// assert
		assert.Equal(t, dataStub, actualData)
	})

	t.Run("fail if file is missing", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		sut := setup(t)

		// exercise
		actualData, err := sut.Read(ctx, nameStub)
		require.Error(t, err)
		require.Empty(t, actualData)

		// assert
		assert.ErrorContains(t, err, "error reading file")
	})

	t.Run("fail if write fails", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		dataStub := []byte("bar")
		sut := setup(t)

		spy := sut.GetSpy()
		spy.Register("Write", 0, assert.AnError, nameStub, dataStub)

		// exercise
		err := sut.Write(ctx, nameStub, dataStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if read fails", func(t *testing.T) {
		t.Parallel()

		// setup
		name := "foo"
		dataStub := []byte("bar")
		sut := setup(t)

		spy := sut.GetSpy()
		spy.Register("Read", 0, assert.AnError, name)

		// exercise
		err := sut.Write(ctx, name, dataStub)
		require.NoError(t, err)

		actualData, err := sut.Read(ctx, name)
		require.Error(t, err)
		require.Empty(t, actualData)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
