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

func setupUserStore(t *testing.T) (*repo.User, *store.InMemory) {
	t.Helper()

	factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

	usersStoreStub := store.NewInMemory(util.NewSpy())
	factory.SetStore(usersStoreStub, compose.UserStore)

	sut := factory.CreateUserRepo(usersStoreStub)

	return sut, usersStoreStub
}

func TestUserModelMap_Slice(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		userModelMap := repo.UserModelMap{
			"user1": {Name: "user1"},
			"user2": {Name: "user2"},
		}

		// execute
		users := userModelMap.Slice()

		// assert
		assert.Contains(t, users, userModelMap["user1"])
		assert.Contains(t, users, userModelMap["user2"])
	})
}

func TestUser_Create_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "user1"
		emailStub := "user1@example.com"
		passwordStub := "password"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, _ := setupUserStore(t)

		// execute
		user, err := sut.Create(ctx, nameStub, emailStub, passwordStub, isAdminStub, accessStub)
		require.NoError(t, err)

		user2, err := sut.Get(ctx, nameStub)
		require.NoError(t, err)

		// assert
		assert.Equal(t, user, user2)
		assert.Equal(t, nameStub, user.Name)
		assert.Equal(t, accessStub, user.Access)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "user1"
		emailStub := "user1@example.com"
		passwordStub := "password"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		user, err := sut.Create(ctx, nameStub, emailStub, passwordStub, isAdminStub, accessStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "user1"
		emailStub := "user1@example.com"
		passwordStub := "password"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		user, err := sut.Create(ctx, nameStub, emailStub, passwordStub, isAdminStub, accessStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "user1"
		emailStub := "user1@example.com"
		passwordStub := "password"
		isAdminStub := true
		accessStub := []string{"user1", "user2"}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		// execute
		user, err := sut.Create(ctx, nameStub, emailStub, passwordStub, isAdminStub, accessStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"user1": {Name: "user1"},
			"user2": {Name: "user2"},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Get(ctx, "user1")
		require.NoError(t, err)

		// assert
		assert.Equal(t, data["user1"], user)
	})

	t.Run("fail if user is missing", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Get(ctx, "user1")
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("fail if read fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("Read", 0, assert.AnError)

		// execute
		users, err := sut.Get(ctx, "user1")
		require.Error(t, err)
		require.Empty(t, users)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		users, err := sut.Get(ctx, "user1")
		require.Error(t, err)
		require.Empty(t, users)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})
}

func TestUser_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"user1": {Name: "user1"},
			"user2": {Name: "user2"},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		users, err := sut.List(ctx)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, users)
		assert.Contains(t, users, data["user1"])
		assert.Contains(t, users, data["user2"])
	})

	t.Run("fail if read fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("Read", 0, assert.AnError)

		// execute
		users, err := sut.List(ctx)
		require.Error(t, err)
		require.Empty(t, users)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if unmarshaling fails", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Write(ctx, []byte("invalid"))
		require.NoError(t, err)

		// execute
		users, err := sut.List(ctx)
		require.Error(t, err)
		require.Empty(t, users)

		// assert
		assert.ErrorContains(t, err, "error unmarshaling data")
	})
}

func TestUser_UpdateAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		accessStub := []string{"user1", "user2"}

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.UpdateAccess(ctx, nameStub, accessStub)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user)
		assert.Equal(t, accessStub, user.Access)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		accessStub := []string{"user1", "user2"}

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		user, err := sut.UpdateAccess(ctx, nameStub, accessStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		accessStub := []string{"user1", "user2"}

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.UpdateAccess(ctx, nameStub, accessStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_UpdatePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		passwordStub := "password"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.UpdatePassword(ctx, nameStub, passwordStub)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user)
		assert.Equal(t, passwordStub, user.Password)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		passwordStub := "password"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		user, err := sut.UpdatePassword(ctx, nameStub, passwordStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		passwordStub := "password"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.UpdatePassword(ctx, nameStub, passwordStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_Promote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Promote(ctx, nameStub)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user)
		assert.True(t, user.IsAdmin)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		user, err := sut.Promote(ctx, nameStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Promote(ctx, nameStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_Demote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub, IsAdmin: true},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Demote(ctx, nameStub)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user)
		assert.False(t, user.IsAdmin)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub, IsAdmin: true},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		user, err := sut.Demote(ctx, nameStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub, IsAdmin: true},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Demote(ctx, nameStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_AddCSRF(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		tokenStub := "token"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.AddCSRF(ctx, nameStub, tokenStub)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user.CSRFList)
		assert.Equal(t, tokenStub, user.CSRFList[0].Token)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		tokenStub := "token"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		user, err := sut.AddCSRF(ctx, nameStub, tokenStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"
		tokenStub := "token"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.AddCSRF(ctx, nameStub, tokenStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub, IsAdmin: true},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		err = sut.Delete(ctx, nameStub)
		require.NoError(t, err)

		user, err := sut.Get(ctx, nameStub)
		require.Error(t, err)
		require.Empty(t, user)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("fail if ReadForWrite fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub, IsAdmin: true},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// execute
		err = sut.Delete(ctx, nameStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail if WriteLocked fails", func(t *testing.T) {
		t.Parallel()

		// data
		nameStub := "foo"

		data := repo.UserModelMap{
			nameStub: {Name: nameStub, IsAdmin: true},
		}

		// setup
		sut, userStoreStub := setupUserStore(t)

		spy := userStoreStub.GetSpy()
		spy.Register("WriteLocked", 0, assert.AnError, util.Any)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		err = sut.Delete(ctx, nameStub)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
