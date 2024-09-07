package filesystem_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"

	"github.com/peteraba/cloudy-files/appconfig"
	composeTest "github.com/peteraba/cloudy-files/compose/test"
	"github.com/peteraba/cloudy-files/filesystem"
)

func TestS3_Write_and_Read(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	setup := func(t *testing.T) *filesystem.S3 {
		t.Helper()

		const bucket = "cloudy-files-123-test"

		awsConfig, err := config.LoadDefaultConfig(ctx)
		require.NoError(t, err)

		factory := composeTest.NewTestFactory(t, appconfig.NewConfig()).SetAWS(awsConfig)

		return filesystem.NewS3(factory.GetS3Client(), factory.GetLogger(), bucket)
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// data
		stubFileName := gofakeit.UUID() + ".txt"
		stubData := []byte("bar")

		// setup
		sut := setup(t)

		// execute
		err := sut.Write(ctx, stubFileName, stubData)
		require.NoError(t, err)

		data, err := sut.Read(ctx, stubFileName)
		require.NoError(t, err)

		// assert
		require.Equal(t, stubData, data)
	})
}
