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

func setupSessionStore(t *testing.T) (*repo.Session, *store.InMemory) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	sessionStoreStub := store.NewInMemory(util.NewSpy())
	factory.SetStore(sessionStoreStub, compose.SessionStore)

	sut := factory.CreateSessionRepo(sessionStoreStub)

	return sut, sessionStoreStub
}

func TestSessionModelMap_Slice(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		sessionModelMap := repo.SessionModelMap{
			"session1": {Hash: "session1"},
			"session2": {Hash: "session2"},
		}

		// execute
		sessions := sessionModelMap.Slice()

		// assert
		assert.Contains(t, sessions, sessionModelMap["session1"])
		assert.Contains(t, sessions, sessionModelMap["session2"])
	})
}

func TestSession_Start_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "session1"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, _ := setupSessionStore(t)

		// execute
		session, err := sut.Start(ctx, nameStub, isAdminStub, accessStub)
		require.NoError(t, err)

		session2, err := sut.Get(ctx, nameStub)
		require.NoError(t, err)

		// assert
		assert.Equal(t, session, session2)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "session1"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		spy := sessionStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		session, err := sut.Start(ctx, nameStub, isAdminStub, accessStub)
		require.Error(t, err)
		require.Empty(t, session)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "session1"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		sessions, err := sut.Start(ctx, nameStub, isAdminStub, accessStub)
		require.Error(t, err)
		require.Empty(t, sessions)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "session1"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		spy := sessionStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		// execute
		session, err := sut.Start(ctx, nameStub, isAdminStub, accessStub)
		require.Error(t, err)
		require.Empty(t, session)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestSession_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.SessionModelMap{
			"session1": {Hash: "session1"},
			"session2": {Hash: "session2"},
		}

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		session, err := sut.Get(ctx, "session1")
		require.NoError(t, err)

		// assert
		assert.Equal(t, data["session1"], session)
	})

	t.Run("fail if session is missing", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.SessionModelMap{}

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		session, err := sut.Get(ctx, "session1")
		require.Error(t, err)
		require.Empty(t, session)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("fail if read fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		spy := sessionStoreStub.GetSpy()
		spy.Register("Read", 0, assert.AnError)

		// execute
		sessions, err := sut.Get(ctx, "session1")
		require.Error(t, err)
		require.Empty(t, sessions)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		sessions, err := sut.Get(ctx, "session1")
		require.Error(t, err)
		require.Empty(t, sessions)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})
}

func TestSession_CleanUp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// data
	data := repo.SessionModelMap{
		"session1": {Hash: "session1", Expires: 0},
		"session2": {Hash: "session2", Expires: time.Now().Add(-1 * time.Hour).Unix()},
		"session3": {Hash: "session3", Expires: time.Date(3000, 1, 1, 12, 0, 0, 0, time.UTC).Unix()},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		err = sut.CleanUp(ctx)
		require.NoError(t, err)

		actualData, err := sessionStoreStub.Read(ctx)
		require.NoError(t, err)

		// assert
		assert.JSONEq(t, `{"session3":{"hash":"session3","expires":32503723200}}`, string(actualData))
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := sessionStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		err = sut.CleanUp(ctx)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		err := sessionStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		err = sut.CleanUp(ctx)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, sessionStoreStub := setupSessionStore(t)

		spy := sessionStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		// execute
		err := sut.CleanUp(ctx)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
