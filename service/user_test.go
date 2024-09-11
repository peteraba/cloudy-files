package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestUser_Create_and_Login(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy, userData repo.UserModelMap) *service.User {
		t.Helper()

		userStore := store.NewInMemory(userStoreSpy)
		err := userStore.Marshal(ctx, userData)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(userStore, compose.UserStore)

		return factory.CreateUserService()
	}

	t.Run("fail to create user when passing is not OK", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := strings.Repeat("foobar", 20)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, nil)

		// execute
		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.Error(t, err)
		require.Empty(t, userModel)

		// assert
		assert.ErrorContains(t, err, "password is not OK")
	})

	t.Run("fail to create user when user store fails to read", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		userStoreSpy := (util.NewSpy()).Register("ReadForWrite", 0, assert.AnError)

		sut := setup(t, userStoreSpy, nil)

		// execute
		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.Error(t, err)
		require.Empty(t, userModel)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("fail logging in with wrong password", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, repo.UserModelMap{})

		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// extra
		wrongPassword := stubPassword + " "
		err = sut.CheckPassword(ctx, stubName, wrongPassword)
		require.Error(t, err)

		// execute
		sessionHash, err := sut.Login(ctx, stubName, wrongPassword)
		require.Error(t, err)
		require.Empty(t, sessionHash)

		// assert
		assert.ErrorContains(t, err, "password does not match")
	})

	t.Run("login fails if user can not be found", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubPassword := ""

		// setup
		sut := setup(t, unusedSpy, nil)

		// execute
		hash, err := sut.Login(ctx, stubName, stubPassword)
		require.Error(t, err)
		require.Empty(t, hash)

		// assert
		assert.ErrorIs(t, err, apperr.ErrNotFound)
	})

	t.Run("non-admin user can log in", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)
		stubAccess := []string{gofakeit.Adverb(), gofakeit.Adverb()}

		// setup
		sut := setup(t, unusedSpy, repo.UserModelMap{})

		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, false, stubAccess)
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// extra
		err = sut.CheckPassword(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// execute
		sessionHash, err := sut.Login(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, sessionHash)
	})

	t.Run("can create an admin user and user can log in", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubEmail := gofakeit.Email()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		sut := setup(t, unusedSpy, repo.UserModelMap{})

		userModel, err := sut.Create(ctx, stubName, stubEmail, stubPassword, true, []string{})
		require.NoError(t, err)
		require.NotEmpty(t, userModel)

		// extra
		err = sut.CheckPassword(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// execute
		sessionHash, err := sut.Login(ctx, stubName, stubPassword)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, sessionHash)
	})

	t.Run("password can be checked", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		sut := setup(t, unusedSpy, repo.UserModelMap{})
		passwordHash, err := sut.HashPassword(ctx, stubPassword)
		require.NoError(t, err)

		// execute
		err = sut.CheckPasswordHash(ctx, stubPassword, passwordHash)
		require.NoError(t, err)
	})

	t.Run("fail if user already exists", func(t *testing.T) {
		t.Parallel()

		// data
		user := repo.UserModel{
			Name:     "foo",
			Email:    "foo@example.com",
			Password: "foo123Bar321!$",
			IsAdmin:  true,
			Access:   []string{"foo", "bar"},
		}
		userData := repo.UserModelMap{
			"foo": user,
		}

		// setup
		sut := setup(t, unusedSpy, userData)

		// execute
		newUser, err := sut.Create(ctx, user.Name, user.Email, user.Password, user.IsAdmin, user.Access)
		require.Error(t, err)

		// assert
		assert.Empty(t, newUser)
		assert.ErrorContains(t, err, "user already exists")
	})
}

func TestUser_CheckPassword(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy, userData repo.UserModelMap) *service.User {
		t.Helper()

		userStore := store.NewInMemory(userStoreSpy)
		err := userStore.Marshal(ctx, userData)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		factory.SetStore(userStore, compose.UserStore)

		return factory.CreateUserService()
	}

	t.Run("fail if user store fails to read", func(t *testing.T) {
		t.Parallel()

		// data
		stubName := gofakeit.Name()
		stubPassword := gofakeit.Password(true, true, true, true, false, 16)

		// setup
		userStoreSpy := (util.NewSpy()).Register("Read", 0, assert.AnError)

		sut := setup(t, userStoreSpy, repo.UserModelMap{})

		// execute
		err := sut.CheckPassword(ctx, stubName, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_HashPassword(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, hasherSpy *util.Spy) *service.User {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		hasher := password.NewDummyHasher(hasherSpy)
		factory.SetHasher(hasher)

		return factory.CreateUserService()
	}

	t.Run("fail if password is not OK", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := strings.Repeat("foobar", 20)

		// setup
		sut := setup(t, unusedSpy)

		// execute
		hash, err := sut.HashPassword(ctx, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorContains(t, err, "password is not OK")
		assert.Empty(t, hash)
	})

	t.Run("fail if password is not OK", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := "foobarFOOBAR123"

		// setup
		hasherSpy := util.NewSpy()
		hasherSpy.Register("Hash", 0, assert.AnError, stubPassword)
		sut := setup(t, hasherSpy)

		// execute
		hash, err := sut.HashPassword(ctx, stubPassword)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
		assert.Empty(t, hash)
	})
}

func TestUser_CheckPasswordHash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T, hasherSpy *util.Spy) *service.User {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		hasher := password.NewDummyHasher(hasherSpy)
		factory.SetHasher(hasher)

		return factory.CreateUserService()
	}

	t.Run("fail if hasher fails", func(t *testing.T) {
		t.Parallel()

		// data
		stubPassword := "foobarFOOBAR123"
		stubHash := "my-hash-foo-bar-123"

		// setup
		hasherSpy := util.NewSpy()
		hasherSpy.Register("Check", 0, assert.AnError, stubPassword, stubHash)
		sut := setup(t, hasherSpy)

		// execute
		err := sut.CheckPasswordHash(ctx, stubPassword, stubHash)
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_List(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy) (*service.User, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		userStoreStub := store.NewInMemory(userStoreSpy)

		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateUserService(), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "foo123Bar321!$",
				IsAdmin:  true,
				Access:   []string{"foo", "bar"},
			},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		list, err := sut.List(ctx)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, list)
	})

	t.Run("fail if repo fails to list users", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		userStoreSpy := util.NewSpy()
		userStoreSpy.Register("Read", 0, assert.AnError)
		sut, _ := setup(t, userStoreSpy)

		// execute
		list, err := sut.List(ctx)
		require.Error(t, err)

		// assert
		assert.Empty(t, list)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestUser_UpdatePassword(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy) (*service.User, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		userStoreStub := store.NewInMemory(userStoreSpy)

		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateUserService(), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		user := repo.UserModel{
			Name:     "foo",
			Email:    "foo@example.com",
			Password: "foo123Bar321!$",
			IsAdmin:  false,
			Access:   []string{"foo", "bar"},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, repo.UserModelMap{"foo": user})
		require.NoError(t, err)

		// execute
		newUser, err := sut.UpdatePassword(ctx, "foo", "bar723!@#Rab")
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, newUser)
	})

	t.Run("fail if hashing password fails", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, _ := setup(t, unusedSpy)

		// execute
		user, err := sut.UpdatePassword(ctx, "foo", "bar")
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
		assert.ErrorContains(t, err, "failed to hash password")
	})

	t.Run("fail if repo doesn't have the user to update", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, repo.UserModelMap{})
		require.NoError(t, err)

		// execute
		user, err := sut.UpdatePassword(ctx, "foo", "bar723!@#Rab?")
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
		assert.ErrorContains(t, err, "failed to update password")
	})

	t.Run("fail when repo fails to update", func(t *testing.T) {
		t.Parallel()

		// data
		user := repo.UserModel{
			Name:     "foo",
			Email:    "foo@example.com",
			Password: "foo123Bar321!$",
			IsAdmin:  false,
			Access:   []string{"foo", "bar"},
		}

		// setup
		userStoreSpy := util.NewSpy()
		userStoreSpy.Register("ReadForWrite", 0, apperr.ErrAccessDenied)
		sut, userStoreStub := setup(t, userStoreSpy)

		err := userStoreStub.Marshal(ctx, repo.UserModelMap{"foo": user})
		require.NoError(t, err)

		// execute
		newUser, err := sut.UpdatePassword(ctx, "foo", "bar723!@#Rab")
		require.Error(t, err)

		// assert
		assert.Empty(t, newUser)
		assert.ErrorIs(t, err, apperr.ErrAccessDenied)
	})
}

func TestUser_UpdateAccess(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy) (*service.User, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		userStoreStub := store.NewInMemory(userStoreSpy)

		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateUserService(), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		user := repo.UserModel{
			Name:     "foo",
			Email:    "foo@example.com",
			Password: "foo123Bar321!$",
			IsAdmin:  false,
			Access:   []string{"foo", "bar"},
		}
		data := repo.UserModelMap{
			"foo": user,
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		newUser, err := sut.UpdateAccess(ctx, user.Name, user.Access)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, newUser)
	})

	t.Run("fail if repo doesn't have the user to update", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, repo.UserModelMap{})
		require.NoError(t, err)

		// execute
		user, err := sut.UpdateAccess(ctx, "foo", nil)
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
	})
}

func TestUser_Promote(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy) (*service.User, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		userStoreStub := store.NewInMemory(userStoreSpy)

		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateUserService(), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "foo123Bar321!$",
				IsAdmin:  false,
				Access:   []string{"foo", "bar"},
			},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Promote(ctx, data["foo"].Name)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user)
	})

	t.Run("fail to promote admin user", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "foo123Bar321!$",
				IsAdmin:  true,
				Access:   []string{"foo", "bar"},
			},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Promote(ctx, data["foo"].Name)
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
		assert.ErrorContains(t, err, "user is already an admin")
	})

	t.Run("fail if repo fails to promote user", func(t *testing.T) {
		t.Parallel()

		// setup
		userStoreSpy := util.NewSpy()
		userStoreSpy.Register("ReadForWrite", 0, assert.AnError)
		sut, _ := setup(t, userStoreSpy)

		// execute
		user, err := sut.Promote(ctx, "foo")
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
	})

	t.Run("fail if repo doesn't have the user to promote", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, repo.UserModelMap{})
		require.NoError(t, err)

		// execute
		user, err := sut.Promote(ctx, "foo")
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
	})
}

func TestUser_Demote(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy) (*service.User, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		userStoreStub := store.NewInMemory(userStoreSpy)

		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateUserService(), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "foo123Bar321!$",
				IsAdmin:  true,
				Access:   []string{"foo", "bar"},
			},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Demote(ctx, data["foo"].Name)
		require.NoError(t, err)

		// assert
		assert.NotEmpty(t, user)
	})

	t.Run("fail to demote normal user", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "foo123Bar321!$",
				IsAdmin:  false,
				Access:   []string{"foo", "bar"},
			},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		user, err := sut.Demote(ctx, data["foo"].Name)
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
		assert.ErrorContains(t, err, "user is not an admin")
	})

	t.Run("fail if repo fails to demote user", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		userStoreSpy := util.NewSpy()
		userStoreSpy.Register("ReadForWrite", 0, assert.AnError)
		sut, _ := setup(t, userStoreSpy)

		// execute
		user, err := sut.Demote(ctx, "foo")
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
	})

	t.Run("fail if repo doesn't have the user to promote", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, repo.UserModelMap{})
		require.NoError(t, err)

		// execute
		user, err := sut.Demote(ctx, "foo")
		require.Error(t, err)

		// assert
		assert.Empty(t, user)
	})
}

func TestUser_Delete(t *testing.T) {
	t.Parallel()

	unusedSpy := util.NewSpy() // DO NOT USE !!!
	ctx := context.Background()

	setup := func(t *testing.T, userStoreSpy *util.Spy) (*service.User, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())
		userStoreStub := store.NewInMemory(userStoreSpy)

		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateUserService(), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		data := repo.UserModelMap{
			"foo": {
				Name:     "foo",
				Email:    "foo@example.com",
				Password: "foo123Bar321!$",
				IsAdmin:  true,
				Access:   []string{"foo", "bar"},
			},
		}

		// setup
		sut, userStoreStub := setup(t, unusedSpy)

		err := userStoreStub.Marshal(ctx, data)
		require.NoError(t, err)

		// execute
		err = sut.Delete(ctx, data["foo"].Name)
		require.NoError(t, err)

		// assert
		actualList, err := sut.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, actualList)
	})

	t.Run("fail if service fails to delete user", func(t *testing.T) {
		t.Parallel()

		// data

		// setup
		userStoreSpy := util.NewSpy()
		userStoreSpy.Register("ReadForWrite", 0, assert.AnError)
		sut, _ := setup(t, userStoreSpy)

		// execute
		err := sut.Delete(ctx, "foo")
		require.Error(t, err)

		// assert
		assert.ErrorIs(t, err, assert.AnError)
	})
}
