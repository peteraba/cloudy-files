package service_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/filesystem"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestFile_Upload(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy) *service.File {
		t.Helper()

		factory := compose.NewTestFactory(appconfig.NewConfig())

		factory.SetFileSystem(filesystem.NewInMemory(fsStoreSpy))
		factory.SetStore(store.NewInMemory(fileStoreSpy), compose.FileStore)

		return factory.CreateFileService()
	}

	t.Run("fail upload when writing fails", func(t *testing.T) {
		t.Parallel()

		// data
		stubFileName := "foo.txt"

		// setup
		fsStoreSpy := util.NewSpy()
		fsStoreSpy.Register("Write", 0, assert.AnError, stubFileName, util.Any)

		sut := setup(t, unusedSpy, fsStoreSpy)

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte{}, []string{})
		require.Error(t, err)
		require.Empty(t, fileModel)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail upload when readForWrite fails", func(t *testing.T) {
		t.Parallel()

		// setup
		fileStoreSpy := util.NewSpy()
		fileStoreSpy.Register("ReadForWrite", 0, assert.AnError)

		sut := setup(t, fileStoreSpy, unusedSpy)

		// execute
		fileModel, err := sut.Upload(ctx, "foo", []byte{}, []string{})
		require.Error(t, err)
		require.Empty(t, fileModel)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestFile_Retrieve(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy, fileStoreData []byte) *service.File {
		t.Helper()

		fsStore := filesystem.NewInMemory(fsStoreSpy)

		fileStore := store.NewInMemory(fileStoreSpy)

		if fileStoreData != nil {
			err := fileStore.Write(ctx, fileStoreData)
			require.NoError(t, err)
		}

		factory := compose.NewTestFactory(appconfig.NewConfig())

		factory.SetFileSystem(fsStore)
		factory.SetStore(fileStore, compose.FileStore)

		return factory.CreateFileService()
	}

	t.Run("fail when store read fails", func(t *testing.T) {
		t.Parallel()

		// setup
		fileStoreSpy := util.NewSpy()
		fileStoreSpy.Register("Read", 0, assert.AnError)

		sut := setup(t, fileStoreSpy, unusedSpy, nil)

		// execute
		data, err := sut.Retrieve(ctx, "foo", []string{})
		require.Error(t, err)
		require.Nil(t, data)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail when store is invalid", func(t *testing.T) {
		t.Parallel()

		// data
		stubFileName := "foo.txt"

		// setup
		sut := setup(t, unusedSpy, unusedSpy, []byte("invalid json"))

		// execute
		data, err := sut.Retrieve(ctx, stubFileName, []string{})
		require.Error(t, err)
		require.Nil(t, data)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("fail if retrieving the file fails", func(t *testing.T) {
		t.Parallel()

		// setup
		stubFileName := "foo.txt"
		stubAccess := []string{"foobar"}

		// mock
		sut := setup(t, unusedSpy, unusedSpy, []byte("{}"))

		// execute
		data, err := sut.Retrieve(ctx, stubFileName, stubAccess)
		require.Error(t, err)
		require.Nil(t, data)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("fail when retrieving the data fails", func(t *testing.T) {
		t.Parallel()

		// setup
		stubFileName := "foo.txt"
		stubAccess := []string{"foobar"}

		// mock
		fsStoreSpy := util.NewSpy()
		fsStoreSpy.Register("Read", 0, assert.AnError, stubFileName)
		sut := setup(t, unusedSpy, fsStoreSpy, []byte(`{"foo.txt":{"name":"foo.txt","access":["foobar"]}}`))

		// execute
		data, err := sut.Retrieve(ctx, stubFileName, stubAccess)
		require.Error(t, err)
		require.Nil(t, data)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestFile_Upload_and_Get(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy, fileStoreData []byte) *service.File {
		t.Helper()

		fsStore := filesystem.NewInMemory(fsStoreSpy)

		fileStore := store.NewInMemory(fileStoreSpy)
		err := fileStore.Write(ctx, fileStoreData)
		require.NoError(t, err)

		factory := compose.NewTestFactory(appconfig.NewConfig())

		factory.SetFileSystem(fsStore)
		factory.SetStore(fileStore, compose.FileStore)

		return factory.CreateFileService()
	}

	t.Run("can upload a file and get model", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil)

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		fileModel, err = sut.Get(ctx, stubFileName)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		// assert
		assert.Equal(t, stubFileName, fileModel.Name)
		assert.Equal(t, stubAccess, fileModel.Access)
	})
}

func TestFile_Upload_and_Retrieve(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy, fileStoreData []byte) *service.File {
		t.Helper()

		fsStore := filesystem.NewInMemory(fsStoreSpy)

		fileStore := store.NewInMemory(fileStoreSpy)
		err := fileStore.Write(ctx, fileStoreData)
		require.NoError(t, err)

		factory := compose.NewTestFactory(appconfig.NewConfig())

		factory.SetFileSystem(fsStore)
		factory.SetStore(fileStore, compose.FileStore)

		return factory.CreateFileService()
	}

	t.Run("fail retrieve when access is missing", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil)

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		data, err := sut.Retrieve(ctx, stubFileName, []string{})
		require.Error(t, err)

		// assert
		assert.Nil(t, data)
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})

	t.Run("can upload file once and retrieve it multiple times", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil)

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		data, err := sut.Retrieve(ctx, stubFileName, stubAccess)
		require.NoError(t, err)

		data2, err := sut.Retrieve(ctx, stubFileName, stubAccess)
		require.NoError(t, err)

		// assert
		assert.Equal(t, stubData, string(data))
		assert.Equal(t, stubData, string(data2))
	})

	t.Run("can upload and overwrite a file", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)
		stubData2 := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, nil)

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		fileModel, err = sut.Upload(ctx, stubFileName, []byte(stubData2), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		data2, err := sut.Retrieve(ctx, stubFileName, stubAccess)
		require.NoError(t, err)

		// assert
		assert.Equal(t, stubData2, string(data2))
	})
}
