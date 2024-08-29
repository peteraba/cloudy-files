package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/phuslu/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/store"
)

func TestFile_Write_and_Read(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T, fileName string, data []byte) *store.File {
		t.Helper()

		factory := compose.NewFactory()
		logger := factory.GetLogger()
		fileStore := store.NewFile(logger, fileName)

		if data != nil {
			err := fileStore.Write(data)
			require.NoError(t, err)
		}

		return fileStore
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

		data, err := sut.ReadForWrite()
		require.NoError(t, err)
		require.Equal(t, stubData, data)

		// execute
		data, err = sut.Read()
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

		data, err := sut.ReadForWrite()
		require.NoError(t, err)
		require.Equal(t, stubData, data)

		// execute
		err = sut.Write(stubData)
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

		// execute
		err := sut.Write(stubData)
		require.NoError(t, err)

		got, err := sut.Read()
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

		channel := make(chan struct{})

		f := func(ch chan struct{}) {
			got, err := sut.Read()
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

func TestFile_ReadForWrite_and_WriteLocked(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T, fileName string, data []byte) (*store.File, log.Logger) { //nolint:unparam // it makes debugging easier if needed
		t.Helper()

		factory := compose.NewFactory()
		logger := factory.GetLogger()
		sut := store.NewFile(logger, fileName)

		if data != nil {
			err := sut.Write(data)
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

		data, err := sut.ReadForWrite()
		require.NoError(t, err)
		require.Equal(t, stubData, data)

		// execute
		data, err = sut.ReadForWrite()
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

		_, err := sut.ReadForWrite()
		require.NoError(t, err)

		err = sut.Unlock()
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

		// execute
		got, err := sut.ReadForWrite()
		require.NoError(t, err)

		err = sut.WriteLocked(stubData)
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

		// execute
		got, err := sut.ReadForWrite()
		require.NoError(t, err)
		require.Equal(t, stubData, got)

		channel := make(chan error)

		go func(ch chan error) {
			// this will not be finished before WriteLocked call is finished
			err2 := sut.Write(expected)

			ch <- err2
		}(channel)

		time.Sleep(store.DefaultWaitTime)

		err = sut.WriteLocked(stubData)
		require.NoError(t, err)

		// Ensure that the async write attempt can finish
		time.Sleep(2 * store.DefaultWaitTime)

		err = <-channel
		require.NoError(t, err)

		data, err := sut.Read()
		require.NoError(t, err)

		// assert
		assert.Equal(t, expected, data)
	})
}
