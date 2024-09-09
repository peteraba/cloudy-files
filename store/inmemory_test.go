package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestInMemory_Read_and_Write(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T, data repo.UserModelMap) *store.InMemory {
		t.Helper()

		sut := store.NewInMemory(util.NewSpy())

		if data != nil {
			err := sut.Marshal(ctx, data)
			require.NoError(t, err)
		}

		return sut
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		dataStub := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, dataStub)

		dataToRead, err := json.Marshal(dataStub)
		require.NoError(t, err)

		// execute
		err = sut.Write(ctx, dataToRead)
		require.NoError(t, err)

		actualData, err := sut.Read(ctx)
		require.NoError(t, err)

		// assert
		require.Equal(t, dataToRead, actualData)
	})

	t.Run("fail if write is set up to fail", func(t *testing.T) {
		t.Parallel()

		// data
		dataStub := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, dataStub)

		dataToRead, err := json.Marshal(dataStub)
		require.NoError(t, err)

		spy := sut.GetSpy()
		spy.Register("Write", 0, assert.AnError, dataToRead)

		// execute
		err = sut.Write(ctx, dataToRead)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if read is set up to fail", func(t *testing.T) {
		t.Parallel()

		// data
		dataStub := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, dataStub)

		spy := sut.GetSpy()
		spy.Register("Read", 0, assert.AnError)

		dataToRead, err := json.Marshal(dataStub)
		require.NoError(t, err)

		// execute
		err = sut.Write(ctx, dataToRead)
		require.NoError(t, err)

		actualData, err := sut.Read(ctx)
		require.Error(t, err)
		require.Empty(t, actualData)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestInMemory_ReadForWrite_and_WriteLocked_Unlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T, data repo.UserModelMap) *store.InMemory {
		t.Helper()

		sut := store.NewInMemory(util.NewSpy())

		err := sut.Marshal(ctx, data)
		require.NoError(t, err)

		return sut
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		userModelMap := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}
		userModelMap2 := repo.UserModelMap{
			"bar": {
				Name:     "bar",
				Email:    "bar@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, userModelMap)

		dataToRead, err := json.Marshal(userModelMap)
		require.NoError(t, err)

		dataToWrite, err := json.Marshal(userModelMap2)
		require.NoError(t, err)

		// execute
		actualData, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)

		err = sut.WriteLocked(ctx, dataToWrite)
		require.NoError(t, err)

		err = sut.Unlock(ctx)
		require.NoError(t, err)

		actualData2, err := sut.Read(ctx)
		require.NoError(t, err)

		// assert
		assert.Equal(t, dataToRead, actualData)
		assert.Equal(t, dataToWrite, actualData2)
	})

	t.Run("fail if ReadForWrite is set up to fail", func(t *testing.T) {
		t.Parallel()

		// data
		userModelMap := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, userModelMap)

		spy := sut.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		actualData, err := sut.ReadForWrite(ctx)
		require.Error(t, err)
		require.Empty(t, actualData)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked is set up to fail", func(t *testing.T) {
		t.Parallel()

		// data
		userModelMap := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}
		userModelMap2 := repo.UserModelMap{
			"bar": {
				Name:     "bar",
				Email:    "bar@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, userModelMap)

		dataToRead, err := json.Marshal(userModelMap)
		require.NoError(t, err)

		dataToWrite, err := json.Marshal(userModelMap2)
		require.NoError(t, err)

		spy := sut.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, dataToWrite)

		// execute
		actualData, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)

		err = sut.WriteLocked(ctx, dataToWrite)
		require.Error(t, err)

		// assert
		assert.Equal(t, dataToRead, actualData)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if Unlock is set up to fail", func(t *testing.T) {
		t.Parallel()

		// data
		userModelMap := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}
		userModelMap2 := repo.UserModelMap{
			"bar": {
				Name:     "bar",
				Email:    "bar@example.com",
				Password: "password",
				IsAdmin:  true,
			},
		}

		// setup
		sut := setup(t, userModelMap)

		spy := sut.GetSpy()
		spy.Register("Unlock", 0, assert.AnError)

		dataToRead, err := json.Marshal(userModelMap)
		require.NoError(t, err)

		dataToWrite, err := json.Marshal(userModelMap2)
		require.NoError(t, err)

		// execute
		actualData, err := sut.ReadForWrite(ctx)
		require.NoError(t, err)

		err = sut.WriteLocked(ctx, dataToWrite)
		require.NoError(t, err)

		err = sut.Unlock(ctx)
		require.Error(t, err)

		// assert
		assert.Equal(t, dataToRead, actualData)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLock is called without lock", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut := setup(t, nil)

		dataToWrite, err := json.Marshal(nil)
		require.NoError(t, err)

		// execute
		err = sut.WriteLocked(ctx, dataToWrite)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrLockDoesNotExist)
	})
}

func TestInMemory_Marshal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) *store.InMemory {
		t.Helper()

		sut := store.NewInMemory(util.NewSpy())

		return sut
	}

	t.Run("fail when the data can not be JSON marshaled", func(t *testing.T) {
		t.Parallel()

		// setup
		sut := setup(t)

		// execute
		err := sut.Marshal(ctx, make(chan int))
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "error marshaling data")
	})
}
