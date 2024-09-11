package repo_test

import (
	"context"
	"testing"
	"time"

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

func TestCSRF_Create_and_Use(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const (
		ipAddressStub = "158.121.70.25"
		tokenStub     = "f8e414b2"
	)

	setup := func(t *testing.T) *repo.CSRF {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		csrfStoreStub := store.NewInMemory(util.NewSpy())
		factory.SetStore(csrfStoreStub, compose.CSRFStore)

		return factory.CreateCSRFRepo(csrfStoreStub)
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		sut := setup(t)

		// execute
		err := sut.Create(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)

		err = sut.Use(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)
	})

	t.Run("fail if wrong token is provided", func(t *testing.T) {
		t.Parallel()

		// setup
		sut := setup(t)

		// execute
		err := sut.Create(ctx, ipAddressStub, tokenStub)
		require.NoError(t, err)

		err = sut.Use(ctx, ipAddressStub, "invalid-token")
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})
}

func TestCSRF_Use(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const (
		ipAddressStub = "158.121.70.25"
		tokenStub     = "f8e414b2"
	)

	setup := func(t *testing.T) (*repo.CSRF, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		csrfStoreStub := store.NewInMemory(util.NewSpy())
		factory.SetStore(csrfStoreStub, compose.CSRFStore)

		return factory.CreateCSRFRepo(csrfStoreStub), csrfStoreStub
	}

	t.Run("fail if token is not created for ip address", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, _ := setup(t)

		// execute
		err := sut.Use(ctx, ipAddressStub, tokenStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})

	t.Run("fail if the token provided has expired", func(t *testing.T) {
		t.Parallel()

		// data
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

		err := csrfStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		err = sut.Use(ctx, ipAddressStub, "invalid-token")
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})

	t.Run("fail if ReadForWrite fails in Use", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, csrfStoreStub := setup(t)

		spy := csrfStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		err := sut.Use(ctx, ipAddressStub, tokenStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails in Use", func(t *testing.T) {
		t.Parallel()

		// setup
		data := repo.CSRFModelMap{
			ipAddressStub: repo.CSRFModels{
				{
					Token:   tokenStub,
					Expires: time.Now().Add(time.Hour).Unix(),
				},
			},
		}

		// setup
		sut, csrfStoreStub := setup(t)

		// execute
		err := csrfStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := csrfStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		// execute
		err = sut.Use(ctx, ipAddressStub, tokenStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestCSRF_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const (
		ipAddressStub = "158.121.70.25"
		tokenStub     = "f8e414b2"
	)

	setup := func(t *testing.T) (*repo.CSRF, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		csrfStoreStub := store.NewInMemory(util.NewSpy())
		factory.SetStore(csrfStoreStub, compose.CSRFStore)

		return factory.CreateCSRFRepo(csrfStoreStub), csrfStoreStub
	}

	t.Run("fail if entries can not be built", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, csrfStoreStub := setup(t)

		err := csrfStoreStub.Write(ctx, []byte("invalid-data"))
		require.NoError(t, err)

		// execute
		err = sut.Create(ctx, ipAddressStub, tokenStub)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("fail if ReadForWrite fails in Create", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, csrfStoreStub := setup(t)

		spy := csrfStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		err := sut.Create(ctx, ipAddressStub, tokenStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails in Create", func(t *testing.T) {
		t.Parallel()

		// data
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

		err := csrfStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := csrfStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		// execute
		err = sut.Create(ctx, ipAddressStub, tokenStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
