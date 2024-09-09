package repo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func setupFileStore(t *testing.T) (*repo.File, *store.InMemory) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	filesStoreStub := store.NewInMemory(util.NewSpy())
	factory.SetStore(filesStoreStub, compose.FileStore)

	sut := factory.CreateFileRepo(filesStoreStub)

	return sut, filesStoreStub
}

func TestFileModelMap_Slice(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		fileModelMap := repo.FileModelMap{
			"file1": {Name: "file1"},
			"file2": {Name: "file2"},
		}

		// execute
		files := fileModelMap.Slice()

		// assert
		assert.Contains(t, files, fileModelMap["file1"])
		assert.Contains(t, files, fileModelMap["file2"])
	})
}

func TestFile_Create_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "file1"
		accessStub := []string{"user1", "user2"}

		// setup
		sut, _ := setupFileStore(t)

		// execute
		file, err := sut.Create(ctx, nameStub, accessStub)
		require.NoError(t, err)

		file2, err := sut.Get(ctx, nameStub)
		require.NoError(t, err)

		// assert
		assert.Equal(t, file, file2)
		assert.Equal(t, nameStub, file.Name)
		assert.Equal(t, accessStub, file.Access)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "file1"
		accessStub := []string{"user1", "user2"}

		// setup
		sut, fileStoreStub := setupFileStore(t)

		spy := fileStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		file, err := sut.Create(ctx, nameStub, accessStub)
		require.Error(t, err)
		require.Empty(t, file)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "file1"
		accessStub := []string{"user1", "user2"}

		// setup
		sut, fileStoreStub := setupFileStore(t)

		err := fileStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		files, err := sut.Create(ctx, nameStub, accessStub)
		require.Error(t, err)
		require.Empty(t, files)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "file1"
		accessStub := []string{"user1", "user2"}

		// setup
		sut, fileStoreStub := setupFileStore(t)

		spy := fileStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		// execute
		file, err := sut.Create(ctx, nameStub, accessStub)
		require.Error(t, err)
		require.Empty(t, file)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestFile_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.FileModelMap{
			"file1": {Name: "file1"},
			"file2": {Name: "file2"},
		}

		// setup
		sut, fileStoreStub := setupFileStore(t)

		err := fileStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		file, err := sut.Get(ctx, "file1")
		require.NoError(t, err)

		// assert
		assert.Equal(t, data["file1"], file)
	})

	t.Run("fail if file is missing", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.FileModelMap{}

		// setup
		sut, fileStoreStub := setupFileStore(t)

		err := fileStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		file, err := sut.Get(ctx, "file1")
		require.Error(t, err)
		require.Empty(t, file)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("fail if read fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, fileStoreStub := setupFileStore(t)

		spy := fileStoreStub.GetSpy()
		spy.Register("Read", 0, assert.AnError)

		// execute
		files, err := sut.Get(ctx, "file1")
		require.Error(t, err)
		require.Empty(t, files)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, fileStoreStub := setupFileStore(t)

		err := fileStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		files, err := sut.Get(ctx, "file1")
		require.Error(t, err)
		require.Empty(t, files)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})
}

func TestFile_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.FileModelMap{
			"file1": {Name: "file1"},
			"file2": {Name: "file2"},
		}

		// setup
		sut, fileStoreStub := setupFileStore(t)

		err := fileStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		files, err := sut.List(ctx)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, files)
		assert.Contains(t, files, data["file1"])
		assert.Contains(t, files, data["file2"])
	})

	t.Run("fail if read fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, fileStoreStub := setupFileStore(t)

		spy := fileStoreStub.GetSpy()
		spy.Register("Read", 0, assert.AnError)

		// execute
		files, err := sut.List(ctx)
		require.Error(t, err)
		require.Empty(t, files)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, fileStoreStub := setupFileStore(t)

		err := fileStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		files, err := sut.List(ctx)
		require.Error(t, err)
		require.Empty(t, files)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})
}
