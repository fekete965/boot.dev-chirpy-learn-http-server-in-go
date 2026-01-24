package services

import (
	"errors"
	"testing"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		newUserInput := CreateUserInput{
			Email: "test@example.co.uk",
			Password: "duckling", 
		}
		user, err := userService.CreateUser(testHelper.Ctx, newUserInput)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, user.ID)
		require.Equal(t, "test@example.co.uk", user.Email)
		require.NotEqual(t, "", user.HashedPassword)
		require.Equal(t, false, user.IsChirpyRed)
		require.False(t, user.CreatedAt.IsZero())
		require.False(t, user.UpdatedAt.IsZero())

		userCount, err := queries.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), userCount)

		return nil
	})
}

func TestUserService_CreateUser_Returns_ConflictError_When_EmailAlreadyExists(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		newUserInput := CreateUserInput{
			Email: "test@example.co.uk",
			Password: "duckling", 
		}
		user, err := userService.CreateUser(testHelper.Ctx, newUserInput)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, user.ID)
		require.Equal(t, "test@example.co.uk", user.Email)
		require.NotEqual(t, "", user.HashedPassword)
		require.Equal(t, false, user.IsChirpyRed)
		require.False(t, user.CreatedAt.IsZero())
		require.False(t, user.UpdatedAt.IsZero())

		userCount, err := queries.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), userCount)

		user, err = userService.CreateUser(testHelper.Ctx, newUserInput)
		require.Error(t, err)
		require.Equal(t, "email already exists", err.Error())

		return nil
	})
}

func TestUserService_DeleteAllUsers(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		users := []CreateUserInput{
			{
				Email: "test@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test2@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test3@example.co.uk",
				Password: "duckling", 
			},
		}
		for _, user := range users {
			_, err := userService.CreateUser(testHelper.Ctx, user)
			require.NoError(t, err)
		}

		userCount, err := queries.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(3), userCount)

		err = userService.DeleteAllUsers(testHelper.Ctx)
		require.NoError(t, err)

		userCount, err = queries.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(0), userCount)

		return nil
	})
}

func TestUserService_FindUserByEmail(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		users := []CreateUserInput{
			{
				Email: "test@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test2@example.co.uk",
				Password: "duckling", 
			},
		}
		newUserData := make([]User, len(users))

		for i, user := range users {
			newUser, err := userService.CreateUser(testHelper.Ctx, user)
			require.NoError(t, err)
			newUserData[i] = newUser			
		}

		for _, userData := range newUserData {
			foundUser, err := userService.FindUserByEmail(testHelper.Ctx, FindUserByEmailInput{Email: userData.Email})
			require.NoError(t, err)

			require.Equal(t, userData.ID, foundUser.ID)
			require.Equal(t, userData.HashedPassword, foundUser.HashedPassword)
			require.Equal(t, userData.Email, foundUser.Email)
			require.Equal(t, userData.IsChirpyRed, foundUser.IsChirpyRed)
			require.Equal(t, userData.CreatedAt, foundUser.CreatedAt)
			require.Equal(t, userData.UpdatedAt, foundUser.UpdatedAt)
		}
		
		return nil
	})
}

func TestUserService_FindUserByEmail_Returns_NotFound_WhenUserDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		users := []CreateUserInput{
			{
				Email: "test@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test2@example.co.uk",
				Password: "duckling", 
			},
		}
		newUserData := make([]User, len(users))

		for i, user := range users {
			newUser, err := userService.CreateUser(testHelper.Ctx, user)
			require.NoError(t, err)
			newUserData[i] = newUser			
		}

		missingEmailInput := FindUserByEmailInput{Email: "missing@email.co.uk"}
		_, err := userService.FindUserByEmail(testHelper.Ctx, missingEmailInput)
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestUserService_FindUserByID(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		users := []CreateUserInput{
			{
				Email: "test@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test2@example.co.uk",
				Password: "duckling", 
			},
		}
		newUserData := make([]User, len(users))

		for i, user := range users {
			newUser, err := userService.CreateUser(testHelper.Ctx, user)
			require.NoError(t, err)
			newUserData[i] = newUser			
		}

		for _, userData := range newUserData {
			foundUser, err := userService.FindUserByID(testHelper.Ctx, FindUserByIDInput{UserID: userData.ID})
			require.NoError(t, err)

			require.Equal(t, userData.ID, foundUser.ID)
			require.Equal(t, userData.HashedPassword, foundUser.HashedPassword)
			require.Equal(t, userData.Email, foundUser.Email)
			require.Equal(t, userData.IsChirpyRed, foundUser.IsChirpyRed)
			require.Equal(t, userData.CreatedAt, foundUser.CreatedAt)
			require.Equal(t, userData.UpdatedAt, foundUser.UpdatedAt)
		}
		
		return nil
	})
}

func TestUserService_FindUserByID_Returns_NotFound_WhenUserDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		users := []CreateUserInput{
			{
				Email: "test@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test2@example.co.uk",
				Password: "duckling", 
			},
		}
		newUserData := make([]User, len(users))

		for i, user := range users {
			newUser, err := userService.CreateUser(testHelper.Ctx, user)
			require.NoError(t, err)
			newUserData[i] = newUser			
		}

		missingUserId := uuid.New()
		missingUserIdInput := FindUserByIDInput{UserID: missingUserId}
		_, err := userService.FindUserByID(testHelper.Ctx, missingUserIdInput)
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))
		
		return nil
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		newUser, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: "test@example.co.uk",
			Password: "duckling", 
		})
		
		require.NoError(t, err)
		require.Equal(t, "test@example.co.uk", newUser.Email)
		require.Equal(t, false, newUser.IsChirpyRed)

		updatedUserInput := UpdateUserInput{
			UserID: newUser.ID,
			Email: "updated@example.co.uk",
			Password: "a completely new password", 
		}
		updatedUser, err := userService.UpdateUser(testHelper.Ctx, updatedUserInput)
		
		require.NoError(t, err)
		require.Equal(t, "updated@example.co.uk", updatedUser.Email)
		require.Equal(t, false, updatedUser.IsChirpyRed)
		require.True(t, newUser.CreatedAt.Equal(updatedUser.CreatedAt))
		
		require.NotEqual(t, newUser.HashedPassword, updatedUser.HashedPassword)

		return nil
	})
}

func TestUserService_UpdateUser_Returns_ConflictError_When_EmailAlreadyExists(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		users := []CreateUserInput{
			{
				Email: "test@example.co.uk",
				Password: "duckling", 
			},
			{
				Email: "test2@example.co.uk",
				Password: "duckling", 
			},
		}
		newUserData := make([]User, len(users))

		for i, user := range users {
			newUser, err := userService.CreateUser(testHelper.Ctx, user)
			require.NoError(t, err)
			newUserData[i] = newUser			
		}

		updatedUserInput := UpdateUserInput{
			UserID: newUserData[0].ID,
			Email: newUserData[1].Email,
			Password: "a completely new password", 
		}
		_, err := userService.UpdateUser(testHelper.Ctx, updatedUserInput)
		require.Error(t, err)
		require.Equal(t, "email already exists", err.Error())

		return nil
	})
}

func TestUserService_UpdateUser_Returns_NotFoundError_When_UserDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		updatedUserInput := UpdateUserInput{
			UserID: uuid.New(),
			Email: "updated@example.co.uk",
			Password: "a completely new password", 
		}
		_, err := userService.UpdateUser(testHelper.Ctx, updatedUserInput)
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestUserService_UpdateUserIsChirpyRed(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		newUser, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: "test@example.co.uk",
			Password: "duckling", 
		})
		
		require.NoError(t, err)
		require.Equal(t, "test@example.co.uk", newUser.Email)
		require.Equal(t, false, newUser.IsChirpyRed)

		now := time.Now()
		updatedUserIsChirpyRedInput := UpdateUserIsChirpyRedInput{
			UserID: newUser.ID,
			IsChirpyRed: true,
			UpdatedAt: now,
		}
		updatedUser, err := userService.UpdateUserIsChirpyRed(testHelper.Ctx, updatedUserIsChirpyRedInput)
		
		require.NoError(t, err)
		require.Equal(t, true, updatedUser.IsChirpyRed)
		require.True(t, now.Equal(updatedUser.UpdatedAt))

		return nil
	})
}

func TestUserService_UpdateUserIsChirpyRed_Returns_NotFoundError_When_UserDoesNotExist(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		
		updatedUserIsChirpyRedInput := UpdateUserIsChirpyRedInput{
			UserID: uuid.New(),
			IsChirpyRed: true,
			UpdatedAt: time.Now(),
		}
		_, err := userService.UpdateUserIsChirpyRed(testHelper.Ctx, updatedUserIsChirpyRedInput)
		
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}
