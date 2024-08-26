package service_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/store"
)

func TestFile_Upload_and_Retrieve(t *testing.T) {
	t.Parallel()

	factory := compose.NewFactory()

	factory.SetFileSystem(store.NewInMemoryFileSystem())
	factory.SetFileStore(store.NewInMemoryFile())

	// Create a new file service with mock stores
	sut := factory.CreateFileService()

	t.Run("can upload file once and retrieve it multiple times", func(t *testing.T) {
		t.Parallel()

		// setup
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)

		// execute
		err := sut.Upload(stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)

		data, err := sut.Retrieve(stubFileName, stubAccess)
		require.NoError(t, err)

		data2, err := sut.Retrieve(stubFileName, stubAccess)
		require.NoError(t, err)

		// assert
		assert.Equal(t, stubData, string(data))
		assert.Equal(t, stubData, string(data2))
	})

	t.Run("can upload and overwrite a file", func(t *testing.T) {
		t.Parallel()

		// setup
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}
		stubFileName := gofakeit.Adjective() + "." + gofakeit.FileExtension()
		stubData := gofakeit.HipsterSentence(10)
		stubData2 := gofakeit.HipsterSentence(10)

		// execute
		err := sut.Upload(stubFileName, []byte(stubData), stubAccess)
		require.NoError(t, err)

		err = sut.Upload(stubFileName, []byte(stubData2), stubAccess)
		require.NoError(t, err)

		data2, err := sut.Retrieve(stubFileName, stubAccess)
		require.NoError(t, err)

		// assert
		assert.Equal(t, stubData2, string(data2))
	})
}
