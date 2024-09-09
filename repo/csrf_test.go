package repo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestCSRF_Create_and_Exist(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) (*repo.CSRF, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		csrfStoreStub := store.NewInMemory(util.NewSpy())
		factory.SetStore(csrfStoreStub, compose.CSRFStore)

		return factory.CreateCSRFRepo(csrfStoreStub), csrfStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		ipAddressStub := "127.0.0.1"
		tokenStub, err := util.RandomHex(8)
		require.NoError(t, err)

		// setup
		sut, _ := setup(t)

		// execute
		err = sut.Create(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)

		exists, err := sut.Exists(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)

		// assert
		assert.True(t, exists)
	})

	t.Run("fail if token is not created for ip address", func(t *testing.T) {
		t.Parallel()

		// data
		ipAddressStub := "127.0.0.1"
		tokenStub, err := util.RandomHex(8)
		require.NoError(t, err)

		// setup
		sut, _ := setup(t)

		// execute
		exists, err := sut.Exists(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)

		// assert
		assert.False(t, exists)
	})

	t.Run("fail if wrong token is provided", func(t *testing.T) {
		t.Parallel()

		// data
		ipAddressStub := "127.0.0.1"
		tokenStub, err := util.RandomHex(8)
		require.NoError(t, err)

		// setup
		sut, _ := setup(t)

		// execute
		err = sut.Create(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)

		exists, err := sut.Exists(ctx, ipAddressStub, "wrong-token")
		require.NoError(t, err)

		// assert
		assert.False(t, exists)
	})

	t.Run("fail if the token provided has expired", func(t *testing.T) {
		t.Parallel()

		// data
		ipAddressStub := "127.0.0.1"
		tokenStub, err := util.RandomHex(8)
		require.NoError(t, err)

		data := repo.CSRFModelMap{
			ipAddressStub: repo.CSRFModels{
				{
					Token:   tokenStub,
					Expires: 0,
				},
			},
		}

		// setup
		sut, csrfStoreStub := setup(t)

		// execute
		err = csrfStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		exists, err := sut.Exists(ctx, ipAddressStub, "wrong-token")
		require.NoError(t, err)

		// assert
		assert.False(t, exists)
	})
}
