package cli_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/cli"
	cliTest "github.com/peteraba/cloudy-files/cli/test"
	"github.com/peteraba/cloudy-files/compose"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/filesystem"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/store"
	"github.com/peteraba/cloudy-files/util"
)

func TestApp_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) (*cli.App, *cliTest.FakeDisplay) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay)
	}

	t.Run("unknown subcommand", func(t *testing.T) {
		t.Parallel()

		// setup
		sut, fakeDisplay := setup(t)

		// execute
		sut.Route(ctx, "foo")

		actual := fakeDisplay.String()

		// assert
		assert.Contains(t, actual, cli.Help)
		assert.Contains(t, actual, "Unknown subcommand: foo")
	})
}

func TestApp_CleanUp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T, sessionData repo.SessionModelMap) (*cli.App, *cliTest.FakeDisplay, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		// setup store
		fakeSessionStore := store.NewInMemory(util.NewSpy())
		err := fakeSessionStore.Marshal(ctx, sessionData)
		require.NoError(t, err)

		factory.SetStore(fakeSessionStore, compose.SessionStore)

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay), fakeSessionStore
	}

	t.Run("fail if session returns an error", func(t *testing.T) {
		t.Parallel()

		// setup
		app, fakeDisplay, fakeSessionStore := setup(t, repo.SessionModelMap{})

		sessionSpy := fakeSessionStore.GetSpy()
		sessionSpy.Register("ReadForWrite", 0, assert.AnError)

		// assert
		fakeDisplay.QueueContainsAssertion("Session cleanup failed.")
		fakeDisplay.QueueContainsAssertion(assert.AnError.Error())

		// execute
		app.Route(ctx, "cleanUp")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		sessionDataStub := repo.SessionModelMap{
			"foo": {
				Hash:    "bar",
				IsAdmin: false,
				Expires: 1725623414,
				Access:  []string{"foo", "bar"},
			},
		}

		app, fakeDisplay, fakeSessionStore := setup(t, sessionDataStub)

		// execute
		app.Route(ctx, "cleanUp")

		// collect data
		data, err := fakeSessionStore.Read(ctx)
		require.NoError(t, err)

		// assert
		assert.Contains(t, fakeDisplay.String(), "Session data is cleaned up.")
		assert.Equal(t, []byte(`{}`), data)
	})
}

func TestApp_HashPassword_CheckPasswordHash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) (*cli.App, *cliTest.FakeDisplay) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay)
	}

	t.Run("fail password too weak", func(t *testing.T) {
		t.Parallel()

		// setup
		passwordStub := "foo"

		sut, fakeDisplay := setup(t)

		// execute
		sut.Route(ctx, "hashPassword", passwordStub)

		// assert
		assert.Contains(t, fakeDisplay.String(), "password is not strong enough")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		passwordStub := "fooFoo123Barbar"

		passwordRegex := regexp.MustCompile(`Hashed password: (\S{60})\n`)

		sut, fakeDisplay := setup(t)

		// execute
		sut.Route(ctx, "hashPassword", passwordStub)
		require.Regexp(t, passwordRegex, fakeDisplay.String())

		found := passwordRegex.FindStringSubmatch(fakeDisplay.String())
		require.Len(t, found, 2)

		sut.Route(ctx, "checkPasswordHash", passwordStub, found[1])

		// assert
		assert.Contains(t, fakeDisplay.String(), "Hashed password:")
	})

	t.Run("fail if password does not match", func(t *testing.T) {
		t.Parallel()

		// setup
		password := "foo"
		hash := "bar"

		sut, fakeDisplay := setup(t)

		// assert
		fakeDisplay.QueueContainsAssertion("Password does not match the hash received")

		// execute
		sut.Route(ctx, "checkPasswordHash", password, hash)
	})
}

func TestApp_CreateUser_CheckPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) (*cli.App, *cliTest.FakeDisplay, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		// setup user store
		userStoreStub := store.NewInMemory(util.NewSpy())
		factory.SetStore(userStoreStub, compose.UserStore)

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay), userStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := gofakeit.Name()
		emailStub := gofakeit.Email()
		passwordStub := gofakeit.Password(true, true, true, true, false, 20)
		isAdminStub := "Y"

		app, fakeDisplay, _ := setup(t)

		// execute
		app.Route(ctx, "createUser", nameStub, emailStub, passwordStub, isAdminStub)

		app.Route(ctx, "checkPassword", nameStub, passwordStub)

		// assert
		assert.Contains(t, fakeDisplay.String(), "User created: "+nameStub)
		assert.Contains(t, fakeDisplay.String(), "Password matches")
	})

	t.Run("fail if password does not match", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		passwordStub := "bar"

		app, fakeDisplay, _ := setup(t)

		// assert
		fakeDisplay.QueueContainsAssertion("Password received does not match the user password.")

		// execute
		app.Route(ctx, "checkPassword", nameStub, passwordStub)
	})

	t.Run("fail if storing user fails", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		emailStub := "foo@example.com"
		passwordStub := "bar"
		isAdminStub := "yes"

		app, fakeDisplay, userStoreStub := setup(t)

		spy := userStoreStub.GetSpy()
		spy.Register("ReadForWrite", 0, assert.AnError)

		// assert
		fakeDisplay.QueueContainsAssertion("User creation failed.")

		// execute
		app.Route(ctx, "createUser", nameStub, emailStub, passwordStub, isAdminStub)
	})
}

func TestApp_CreateUser_Login_CheckSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) (*cli.App, *cliTest.FakeDisplay, *store.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		// setup user store
		{
			userStoreStub := store.NewInMemory(util.NewSpy())

			factory.SetStore(userStoreStub, compose.UserStore)
		}

		// setup session store
		sessionStoreStub := store.NewInMemory(util.NewSpy())

		factory.SetStore(sessionStoreStub, compose.SessionStore)

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay), sessionStoreStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := gofakeit.Name()
		emailStub := gofakeit.Email()
		passwordStub := gofakeit.Password(true, true, true, true, false, 20)
		isAdminStub := "Y"

		hashRegex := regexp.MustCompile(`Session started: ([0-9a-f]{32})\n`)

		app, fakeDisplay, _ := setup(t)

		// execute
		app.Route(ctx, "createUser", nameStub, emailStub, passwordStub, isAdminStub)

		app.Route(ctx, "login", nameStub, passwordStub)
		require.Regexp(t, hashRegex, fakeDisplay.String())

		found := hashRegex.FindStringSubmatch(fakeDisplay.String())
		require.Len(t, found, 2)

		app.Route(ctx, "checkSession", nameStub, found[1])

		// assert
		assert.Contains(t, fakeDisplay.String(), "User created: "+nameStub)
		assert.Contains(t, fakeDisplay.String(), "Session started: ")
	})

	t.Run("fail if login fails", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		passwordStub := "bar"

		app, fakeDisplay, _ := setup(t)

		// assert
		fakeDisplay.QueueContainsAssertion("Login failed.")

		// execute
		app.Route(ctx, "login", nameStub, passwordStub)
	})

	t.Run("fail if session does not exist", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		hashStub := "bar"

		app, fakeDisplay, _ := setup(t)

		// assert
		fakeDisplay.QueueContainsAssertion("Session does not exist.")
		fakeDisplay.QueueContainsAssertion("not found")

		// execute
		app.Route(ctx, "checkSession", nameStub, hashStub)
	})

	t.Run("fail if session does not match the user session", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		hashStub := "bar"

		app, fakeDisplay, sessionStore := setup(t)

		err := sessionStore.Marshal(ctx, repo.SessionModelMap{
			nameStub: {
				Expires: time.Now().Add(time.Hour).Unix(),
				Hash:    "baz",
				IsAdmin: false,
				Access:  []string{"foo", "bar"},
			},
		})
		require.NoError(t, err)

		// assert
		fakeDisplay.QueueContainsAssertion("Session does not exist.")

		// execute
		app.Route(ctx, "checkSession", nameStub, hashStub)
	})

	t.Run("fail if session is expired", func(t *testing.T) {
		t.Parallel()

		// setup
		nameStub := "foo"
		hashStub := "bar"

		app, fakeDisplay, sessionStore := setup(t)

		err := sessionStore.Marshal(ctx, repo.SessionModelMap{
			nameStub: {
				Expires: time.Now().Add(-1 * time.Hour).Unix(),
				Hash:    hashStub,
				IsAdmin: false,
				Access:  []string{"foo", "bar"},
			},
		})
		require.NoError(t, err)

		// assert
		fakeDisplay.QueueContainsAssertion("Session does not exist.")

		// execute
		app.Route(ctx, "checkSession", nameStub, hashStub)
	})
}

func TestApp_Upload_and_Size(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T, files repo.FileModelMap) (*cli.App, *cliTest.FakeDisplay, *filesystem.InMemory) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		// setup store
		{
			storeStub := store.NewInMemory(util.NewSpy())
			err := storeStub.Marshal(ctx, files)
			require.NoError(t, err)

			factory.SetStore(storeStub, compose.FileStore)
		}

		// setup file system
		fsStub := filesystem.NewInMemory(util.NewSpy())
		factory.SetFileSystem(fsStub)

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay), fsStub
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		app, fakeDisplay, _ := setup(t, repo.FileModelMap{})

		// execute
		app.Route(ctx, "upload", fileNameStub, accessStub[0], accessStub[1])

		app.Route(ctx, "size", fileNameStub, accessStub[0])

		// assert
		assert.Contains(t, fakeDisplay.String(), "File stored:")
		assert.Contains(t, fakeDisplay.String(), "File size: 5")
	})

	t.Run("fail if access is missing for reading", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		app, fakeDisplay, _ := setup(t, repo.FileModelMap{})

		// execute
		app.Route(ctx, "upload", fileNameStub, accessStub[0], accessStub[1])

		app.Route(ctx, "size", fileNameStub)

		// assert
		assert.Contains(t, fakeDisplay.String(), "File stored:")
		assert.NotContains(t, fakeDisplay.String(), "File size: 5")
		assert.Contains(t, fakeDisplay.String(), "access denied")
	})

	t.Run("fail if file is missing", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "bar.txt"
		accessStub := []string{"foo", "bar"}

		app, fakeDisplay, _ := setup(t, nil)

		// execute
		app.Route(ctx, "upload", fileNameStub, accessStub[0], accessStub[1])

		// assert
		assert.Contains(t, fakeDisplay.String(), "File could not be found")
	})

	t.Run("fail if upload fails", func(t *testing.T) {
		t.Parallel()

		// setup
		fileNameStub := "foo.txt"
		accessStub := []string{"foo", "bar"}

		app, fakeDisplay, fsStub := setup(t, nil)

		spy := fsStub.GetSpy()
		spy.Register("Write", 0, assert.AnError, fileNameStub, util.Any)

		// execute
		app.Route(ctx, "upload", fileNameStub, accessStub[0], accessStub[1])

		// assert
		assert.Contains(t, fakeDisplay.String(), "File could not be found")
	})
}

func TestApp_MissingArguments(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const missingArguments = "Please provide "

	setup := func(t *testing.T) (*cli.App, *cliTest.FakeDisplay) {
		t.Helper()

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig())

		return factory.CreateCliApp(), factory.GetDisplay().(*cliTest.FakeDisplay)
	}

	tests := []struct {
		name       string
		subcommand string
		args       []string
	}{
		{
			name:       "createUser",
			subcommand: "createUser",
			args:       nil,
		},
		{
			name:       "hashPassword",
			subcommand: "hashPassword",
			args:       nil,
		},
		{
			name:       "login",
			subcommand: "login",
			args:       nil,
		},
		{
			name:       "checkPassword",
			subcommand: "checkPassword",
			args:       nil,
		},
		{
			name:       "checkPasswordHash",
			subcommand: "checkPasswordHash",
			args:       nil,
		},
		{
			name:       "checkSession",
			subcommand: "checkSession",
			args:       nil,
		},
		{
			name:       "upload",
			subcommand: "upload",
			args:       nil,
		},
		{
			name:       "size",
			subcommand: "size",
			args:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			sut, fakeDisplay := setup(t)

			// execute
			sut.Route(ctx, tt.subcommand, tt.args...)

			actual := fakeDisplay.String()

			// assert
			assert.Contains(t, actual, missingArguments)
			assert.Contains(t, actual, cli.Help)
		})
	}
}
