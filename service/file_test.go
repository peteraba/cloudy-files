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
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/filesystem"
	"github.com/peteraba/cloudy-files/repo"
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

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

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

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy, fileStoreData repo.FileModelMap) *service.File {
		t.Helper()

		fsStore := filesystem.NewInMemory(fsStoreSpy)

		fileStore := store.NewInMemory(fileStoreSpy)

		if fileStoreData != nil {
			err := fileStore.Marshal(ctx, fileStoreData)
			require.NoError(t, err)
		}

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

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

	t.Run("fail if retrieving the file fails", func(t *testing.T) {
		t.Parallel()

		// setup
		stubFileName := "foo.txt"
		stubAccess := []string{"foobar"}

		// mock
		sut := setup(t, unusedSpy, unusedSpy, nil)

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
		sut := setup(t, unusedSpy, fsStoreSpy, repo.FileModelMap{"foo.txt": {Name: "foo.txt", Access: []string{"foobar"}}})

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

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy, fileStoreData repo.FileModelMap) *service.File {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		{
			fsStore := filesystem.NewInMemory(fsStoreSpy)
			factory.SetFileSystem(fsStore)
		}

		{
			fileStore := store.NewInMemory(fileStoreSpy)
			err := fileStore.Marshal(ctx, fileStoreData)
			require.NoError(t, err)
			factory.SetStore(fileStore, compose.FileStore)
		}

		return factory.CreateFileService()
	}

	t.Run("can upload a file and get model", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.FileModelMap{})

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

	t.Run("fail if repo fails getting file", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		fileStoreSpy := util.NewSpy()
		fileStoreSpy.Register("Read", 0, assert.AnError)

		sut := setup(t, fileStoreSpy, unusedSpy, repo.FileModelMap{})

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		fileModel, err = sut.Get(ctx, stubFileName)
		require.Error(t, err)
		require.Empty(t, fileModel.Name)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestFile_Upload_and_List(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy()
	ctx := context.Background()

	setup := func(t *testing.T, fileStoreSpy, fsStoreSpy *util.Spy, fileStoreData repo.FileModelMap) *service.File {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		{
			fsStore := filesystem.NewInMemory(fsStoreSpy)
			factory.SetFileSystem(fsStore)
		}

		{
			fileStore := store.NewInMemory(fileStoreSpy)
			err := fileStore.Marshal(ctx, fileStoreData)
			require.NoError(t, err)
			factory.SetStore(fileStore, compose.FileStore)
		}

		return factory.CreateFileService()
	}

	t.Run("can upload a file and list models as admin", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.FileModelMap{})

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		fileModels, err := sut.List(ctx, nil, true)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		// assert
		assert.Len(t, fileModels, 1)
		assert.Equal(t, stubFileName, fileModels[0].Name)
		assert.Equal(t, stubAccess, fileModels[0].Access)
	})

	t.Run("can upload a file and list models as non-admin", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.FileModelMap{})

		// execute
		fileModel, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		fileModels, err := sut.List(ctx, []string{stubAccess[0]}, false)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel.Name)

		// assert
		assert.Len(t, fileModels, 1)
		assert.Equal(t, stubFileName, fileModels[0].Name)
		assert.Equal(t, stubAccess, fileModels[0].Access)
	})

	t.Run("as non-admin only files with matching access are returned", func(t *testing.T) {
		t.Parallel()

		// data
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// setup
		sut := setup(t, unusedSpy, unusedSpy, repo.FileModelMap{})

		// execute
		fileModel1, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess[1:])
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel1.Name)

		fileModel2, err := sut.Upload(ctx, stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel2.Name)

		fileModels, err := sut.List(ctx, []string{stubAccess[0]}, false)
		require.NoError(t, err)
		require.Equal(t, stubFileName, fileModel2.Name)

		// assert
		assert.Len(t, fileModels, 1)
		assert.Equal(t, stubFileName, fileModels[0].Name)
		assert.Equal(t, stubAccess, fileModels[0].Access)
	})

	t.Run("fail if listing fails", func(t *testing.T) {
		t.Parallel()

		// data
		fileStoreSpy := util.NewSpy()
		fileStoreSpy.Register("Read", 0, assert.AnError)

		// setup
		sut := setup(t, fileStoreSpy, unusedSpy, repo.FileModelMap{})

		// execute
		fileModels, err := sut.List(ctx, []string{}, false)
		require.Error(t, err)

		// assert
		assert.Empty(t, fileModels)
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

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

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
