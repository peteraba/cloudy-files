package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/phuslu/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/store"
)

func TestLocal_Write_and_Read(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T, fileName string, data []byte) *store.Local {
		t.Helper()

		factory := compose.NewTestFactory(appconfig.NewConfig())
		logger := factory.GetLogger()
		localStore := store.NewLocal(logger, fileName)

		if data != nil {
			ctx := context.Background()
			err := localStore.Write(ctx, data)
			require.NoError(t, err)
		}

		return localStore
	}

	cleanUp := func(t *testing.T, fileName string) {
		t.Helper()

		_ = os.Remove(fileName + ".lock")
		_ = os.Remove(fileName)
	}

	t.Run("fail to read if lock is stuck", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("stub data")

		// setup
		sut := setup(t, dataFileName, stubData)
		ctx := context.Background()

		data, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)
		require.Equal(t, stubData, data)

		// execute
		data, err = sut.Read(ctx)
		require.Error(t, err)

		// assert
		assert.Nil(t, data)
		assert.ErrorContains(t, err, "error waiting for lock")
	})

	t.Run("fail to write if lock is stuck", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("stub data")

		// setup
		sut := setup(t, dataFileName, stubData)
		ctx := context.Background()

		data, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)
		require.Equal(t, stubData, data)

		// execute
		err = sut.Write(ctx, stubData)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "error waiting for lock")
	})

	t.Run("simple write-read success", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("stub data")

		// setup
		sut := setup(t, dataFileName, nil)
		ctx := context.Background()

		// execute
		err := sut.Write(ctx, stubData)
		require.NoError(t, err)

		got, err := sut.Read(ctx)
		require.NoError(t, err)

		// assert
		assert.Equal(t, stubData, got)
	})

	t.Run("simultaneous read success", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("stub data")

		// setup
		sut := setup(t, dataFileName, stubData)
		ctx := context.Background()

		channel := make(chan struct{})

		f := func(ch chan struct{}) {
			got, err := sut.Read(ctx)
			require.NoError(t, err)
			require.Equal(t, stubData, got)

			ch <- struct{}{}
		}

		// execute
		go f(channel)
		go f(channel)
		timeout := time.After(3 * time.Second)

		received := 0
		for range 2 {
			select {
			case <-channel:
				received++
			case <-timeout:
				t.Fatal("timeout")
			}
		}

		// assert
		assert.Equal(t, 2, received)
	})
}

func TestLocal_ReadForWrite_and_WriteLocked(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T, fileName string, data []byte) (*store.Local, *log.Logger) { //nolint:unparam // it makes debugging easier if needed
		t.Helper()

		factory := compose.NewTestFactory(appconfig.NewConfig())
		logger := factory.GetLogger()
		sut := store.NewLocal(logger, fileName)

		if data != nil {
			ctx := context.Background()
			err := sut.Write(ctx, data)
			require.NoError(t, err)
		}

		return sut, logger
	}

	cleanUp := func(t *testing.T, fileName string) {
		t.Helper()

		_ = os.Remove(fileName + ".lock")
		_ = os.Remove(fileName)
	}

	t.Run("fail to read for write if lock is stuck", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("stub data")

		// setup
		sut, _ := setup(t, dataFileName, stubData)
		ctx := context.Background()

		data, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)
		require.Equal(t, stubData, data)

		// execute
		data, err = sut.ReadForWrite(ctx)
		require.Error(t, err)

		// assert
		assert.Nil(t, data)
		assert.ErrorContains(t, err, "error waiting for lock")
	})

	t.Run("lock can be unlocked manually", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("stub data")

		// setup
		sut, _ := setup(t, dataFileName, stubData)
		ctx := context.Background()

		_, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)

		err = sut.Unlock(ctx)
		require.NoError(t, err)

		// assert
	})

	t.Run("simple write locked success", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("{}")

		// setup
		sut, _ := setup(t, dataFileName, stubData)
		ctx := context.Background()

		// execute
		got, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)

		err = sut.WriteLocked(ctx, stubData)
		require.NoError(t, err)

		// assert
		assert.Equal(t, stubData, got)
	})

	t.Run("attempt to write between read for write and write will wait", func(t *testing.T) {
		t.Parallel()
		dataFileName := gofakeit.UUID()
		defer cleanUp(t, dataFileName)

		// data
		stubData := []byte("{}")
		expected := []byte("foobar")

		// setup
		sut, _ := setup(t, dataFileName, stubData)
		ctx := context.Background()

		// execute
		got, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)
		require.Equal(t, stubData, got)

		channel := make(chan error)

		go func(ch chan error) {
			// this will not be finished before WriteLocked call is finished
			err2 := sut.Write(ctx, expected)

			ch <- err2
		}(channel)

		time.Sleep(store.DefaultWaitTime)

		err = sut.WriteLocked(ctx, stubData)
		require.NoError(t, err)

		// Ensure that the async write attempt can finish
		time.Sleep(2 * store.DefaultWaitTime)

		err = <-channel
		require.NoError(t, err)

		data, err := sut.Read(ctx)
		require.NoError(t, err)

		// assert
		assert.Equal(t, expected, data)
	})
}
