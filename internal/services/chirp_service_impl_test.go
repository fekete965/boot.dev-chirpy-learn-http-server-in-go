package services

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testdb"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChirpService_CreateChirp(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		chirpCountBefore, err := chirpService.Db.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(0), chirpCountBefore)

		newChirp, err := chirpService.CreateChirp(testHelper.Ctx, CreateChirpInput{
			UserID: user.ID,
			Body:   "test chirp",
		})
		require.NoError(t, err)

		require.NotEqual(t, uuid.Nil, newChirp.ID)
		require.Equal(t, user.ID, newChirp.UserID)
		require.Equal(t, "test chirp", newChirp.Body)

		chirpCountAfter, err := chirpService.Db.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(chirpCountBefore+1), chirpCountAfter)

		return nil
	})
}

func TestChirpService_CreateChirp_CleansProfanity(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	type testCase struct {
		input    string
		expected string
	}
	testCases := make([]testCase, 0)
	for _, word := range constants.PROFANE_WORDS {
		input1 := fmt.Sprintf("Testing the word %s", word)
		expected1 := fmt.Sprintf("Testing the word %s", strings.Repeat("*", 4))

		testCases = append(testCases, testCase{
			input:    input1,
			expected: expected1,
		})

		input2 := fmt.Sprintf("Testing the word %s!", word)
		expected2 := fmt.Sprintf("Testing the word %s!", word)
		testCases = append(testCases, testCase{
			input:    input2,
			expected: expected2,
		})
	}

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		for _, testCase := range testCases {
			newChirp, err := chirpService.CreateChirp(testHelper.Ctx, CreateChirpInput{
				UserID: user.ID,
				Body:   testCase.input,
			})
			require.NoError(t, err)
			require.Equal(t, newChirp.Body, testCase.expected)
		}

		return nil
	})
}

func TestChirpService_CreateChirp_Returns_BadRequestError_When_InvalidChirpLength(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		_, err := chirpService.CreateChirp(testHelper.Ctx, CreateChirpInput{
			UserID: user.ID,
			Body:   strings.Repeat("a", constants.MAX_CHIRP_LENGTH+10),
		})
		require.Error(t, err)
		require.Equal(t, "invalid chirp length", err.Error())

		return nil
	})
}

func TestChirpService_CreateChirp_Returns_NotFoundError_When_UserDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		_, err := chirpService.CreateChirp(testHelper.Ctx, CreateChirpInput{
			UserID: uuid.New(),
			Body:   "this user does not exist",
		})
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestChirpService_DeleteChirp(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		chirpCountBefore, err := chirpService.Db.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(5), chirpCountBefore)

		err = chirpService.DeleteChirp(testHelper.Ctx, DeleteChirpInput{
			ChirpID: chirps[0].ID,
			UserID:  user.ID,
		})
		require.NoError(t, err)

		chirpCountAfter, err := chirpService.Db.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(chirpCountBefore-1), chirpCountAfter)

		return nil
	})
}

func TestChirpService_DeleteChirp_Returns_NotFoundError_When_UserDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		err := chirpService.DeleteChirp(testHelper.Ctx, DeleteChirpInput{
			ChirpID: chirps[0].ID,
			UserID:  uuid.New(),
		})
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestChirpService_DeleteChirp_Returns_NotFoundError_When_ChirpDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		err := chirpService.DeleteChirp(testHelper.Ctx, DeleteChirpInput{
			ChirpID: uuid.New(),
			UserID:  user.ID,
		})
		require.Error(t, err)
		require.Equal(t, "chirp not found", err.Error())

		return nil
	})
}

func TestChirpService_DeleteChirp_Returns_ForbiddenError_When_InvalidUserTryingToDeleteChirp(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		users := generator.GenerateUsers(testHelper.Ctx, 2)
		user1 := users[0]
		user2 := users[1]

		chirps := generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		chirp1 := chirps[0]

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		err := chirpService.DeleteChirp(testHelper.Ctx, DeleteChirpInput{
			ChirpID: chirp1.ID,
			UserID:  user2.ID,
		})
		require.Error(t, err)
		require.Equal(t, "invalid user permission", err.Error())

		return nil
	})
}

func TestChirpService_GetAllChirps(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		users := generator.GenerateUsers(testHelper.Ctx, 2)
		user1 := users[0]
		user2 := users[1]

		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{})
		require.NoError(t, err)
		require.Equal(t, 10, len(chirps))

		return nil
	})
}

func TestChirpService_GetAllChirps_Returns_EmptySlice_When_NoChirps(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})
		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{})
		require.NoError(t, err)
		require.Empty(t, chirps)

		return nil
	})
}

func TestChirpService_GetAllChirps_Related_To_A_SpecificUser(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		users := generator.GenerateUsers(testHelper.Ctx, 2)
		user1 := users[0]
		user2 := users[1]

		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{
			UserID: &user1.ID,
			Sort:   &constants.DEFAULT_SORT,
		})
		require.NoError(t, err)

		require.Equal(t, 5, len(chirps))

		chirpUserIds := make([]uuid.UUID, len(chirps))
		for i, chirp := range chirps {
			chirpUserIds[i] = chirp.UserID
		}
		require.NotContains(t, chirpUserIds, user2.ID)

		return nil
	})
}

func TestChirpService_GetAllChirps_Sorted_By_CreatedAt_Desc(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		sort := "desc"
		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{
			UserID: &user.ID,
			Sort:   &sort,
		})
		require.NoError(t, err)
		require.Equal(t, 5, len(chirps))

		for i := 0; i < len(chirps)-1; i++ {
			currentChirp := chirps[i]
			nextChirp := chirps[i+1]

			require.True(t, currentChirp.CreatedAt.After(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestChirpService_GetAllChirps_Sorted_By_CreatedAt_Asc(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		sort := "asc"
		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{
			Sort: &sort,
		})
		require.NoError(t, err)
		require.Equal(t, 5, len(chirps))

		for i := 0; i < len(chirps)-1; i++ {
			currentChirp := chirps[i]
			nextChirp := chirps[i+1]

			require.True(t, currentChirp.CreatedAt.Before(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestChirpService_GetAllChirps_Defaults_To_Asc_When_Sort_Is_Nil(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{})
		require.NoError(t, err)
		require.Equal(t, 5, len(chirps))

		for i := 0; i < len(chirps)-1; i++ {
			currentChirp := chirps[i]
			nextChirp := chirps[i+1]

			require.True(t, currentChirp.CreatedAt.Before(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestChirpService_GetAllChirps_Handles_InvalidSort(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		invalidSort := "invalid_sort"
		chirps, err := chirpService.GetAllChirps(testHelper.Ctx, GetAllChirpsInput{
			UserID: &user.ID,
			Sort:   &invalidSort,
		})
		require.NoError(t, err)
		require.Equal(t, 5, len(chirps))

		for i := 0; i < len(chirps)-1; i++ {
			currentChirp := chirps[i]
			nextChirp := chirps[i+1]

			require.True(t, currentChirp.CreatedAt.Before(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestChirpService_GetChirpByID(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: queries,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)
		targetChirp := chirps[0]

		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		chirp, err := chirpService.GetChirpByID(testHelper.Ctx, GetChirpByIDInput{ChirpID: targetChirp.ID})
		require.NoError(t, err)
		require.Equal(t, targetChirp.ID, chirp.ID)
		require.Equal(t, targetChirp.UserID, chirp.UserID)
		require.Equal(t, targetChirp.Body, chirp.Body)
		require.True(t, targetChirp.CreatedAt.Equal(chirp.CreatedAt))
		require.True(t, targetChirp.UpdatedAt.Equal(chirp.UpdatedAt))

		return nil
	})
}

func TestChirpService_GetChirpByID_Returns_NotFoundError_When_ChirpDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	testCfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		chirpService := NewChirpService(NewChirpServiceInput{
			Cfg:         testCfg,
			Db:          queries,
			UserService: userService,
		})

		_, err := chirpService.GetChirpByID(testHelper.Ctx, GetChirpByIDInput{ChirpID: uuid.New()})
		require.Error(t, err)
		require.Equal(t, "chirp not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}
