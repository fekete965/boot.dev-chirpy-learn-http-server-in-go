package services

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		newUser, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		require.NotEqual(t, loginOutput.UserID, uuid.Nil)
		require.Equal(t, loginOutput.Email, newUser.Email)
		require.Equal(t, loginOutput.IsChirpyRed, newUser.IsChirpyRed)
		require.NotEqual(t, loginOutput.RefreshToken, "")
		require.NotEqual(t, loginOutput.Token, "")
		require.False(t, loginOutput.CreatedAt.IsZero())
		require.False(t, loginOutput.UpdatedAt.IsZero())
		
		return nil
	})
}

func TestAuthService_Login_Returns_NotFoundError_When_User_Not_Found(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		_, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: "test@email.co.uk",
			Password: "duckling",
		})
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestAuthService_Login_Returns_UnauthorizedError_When_Password_Is_Incorrect(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		_, err = authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: "incorrect password",
		})
		require.Error(t, err)
		require.Equal(t, "invalid credentials", err.Error())
		
		return nil
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: cfg,
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)
		
		newAccessToken, err := authService.RefreshToken(testHelper.Ctx, RefreshTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.NoError(t, err)

		require.NotEqual(t, newAccessToken, "")
		require.NotEqual(t, newAccessToken, loginOutput.Token)

		decodedToken, err := auth.ValidateJWT(newAccessToken, cfg.JWTSecret)
		require.NoError(t, err)
		require.Equal(t, decodedToken, loginOutput.UserID)

		return nil
	})
}

func TestAuthService_RefreshToken_Return_NotFoundError_When_RefreshToken_Is_Missing(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		_, err = authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)
		
		_, err = authService.RefreshToken(testHelper.Ctx, RefreshTokenInput{
			RefreshToken: "some random token",
		})
		require.Error(t, err)
		require.Equal(t, "refresh token not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestAuthService_RefreshToken_Return_UnauthorizedError_When_RefreshToken_Is_Expired(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		now := time.Now()
		queries.ExpireRefreshToken(testHelper.Ctx, database.ExpireRefreshTokenParams{
			Token: loginOutput.RefreshToken,
			ExpiresAt: now.Add(-time.Hour),
			UpdatedAt: now,
		})

		_, err = authService.RefreshToken(testHelper.Ctx, RefreshTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.Error(t, err)
		require.Equal(t, "refresh token expired", err.Error())

		return nil
	})
}

func TestAuthService_RefreshToken_Return_UnauthorizedError_When_RefreshToken_Is_Revoked(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		now := time.Now()
		queries.RevokeRefreshToken(testHelper.Ctx, database.RevokeRefreshTokenParams{
			Token: loginOutput.RefreshToken,
			RevokedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: now,
		})

		_, err = authService.RefreshToken(testHelper.Ctx, RefreshTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.Error(t, err)
		require.Equal(t, "refresh token revoked", err.Error())

		return nil
	})
}

func TestAuthService_RevokeToken(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		err = authService.RevokeToken(testHelper.Ctx, RevokeTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.NoError(t, err)

		revokedRefreshToken, err := queries.FindRefreshToken(testHelper.Ctx, loginOutput.RefreshToken)
		require.NoError(t, err)

		now := time.Now()
		require.True(t, revokedRefreshToken.RevokedAt.Valid)
		require.True(t, revokedRefreshToken.RevokedAt.Time.Before(now))

		return nil
	})
}

func TestAuthService_RevokeToken_Returns_NotFoundError_When_Token_Not_Found(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		err := authService.RevokeToken(testHelper.Ctx, RevokeTokenInput{
			RefreshToken: "some random token",
		})
		require.Error(t, err)
		require.Equal(t, "refresh token not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}

func TestAuthService_RevokeToken_Returns_UnauthorizedError_When_Token_Already_Revoked(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		err = authService.RevokeToken(testHelper.Ctx, RevokeTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.NoError(t, err)

		err = authService.RevokeToken(testHelper.Ctx, RevokeTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.Error(t, err)
		require.Equal(t, "refresh token already revoked", err.Error())

		return nil
	})
}

func TestAuthService_RevokeToken_Returns_UnauthorizedError_When_Token_Already_Expired(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		_, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		loginOutput, err := authService.Login(testHelper.Ctx, LoginInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		now := time.Now()
		queries.ExpireRefreshToken(testHelper.Ctx, database.ExpireRefreshTokenParams{
			Token: loginOutput.RefreshToken,
			ExpiresAt: now.Add(-time.Hour),
			UpdatedAt: now,
		})

		err = authService.RevokeToken(testHelper.Ctx, RevokeTokenInput{
			RefreshToken: loginOutput.RefreshToken,
		})
		require.Error(t, err)
		require.Equal(t, "refresh token expired", err.Error())

		return nil
	})
}

func TestAuthService_UpgradeUser(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		newUserEmail := "test@email.co.uk"
		newUserPassword := "duckling"
	
		newUser, err := userService.CreateUser(testHelper.Ctx, CreateUserInput{
			Email: newUserEmail,
			Password: newUserPassword,
		})
		require.NoError(t, err)

		err = authService.UpgradeUser(testHelper.Ctx, UpgradeUserInput{
			UserID: newUser.ID,
		})
		require.NoError(t, err)

		updatedUser, err := userService.FindUserByID(testHelper.Ctx, FindUserByIDInput{
			UserID: newUser.ID,
		})
		require.NoError(t, err)
		require.True(t, updatedUser.IsChirpyRed)

		return nil
	})
}

func TestAuthService_UpgradeUser_Returns_NotFoundError_When_User_Not_Found(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	testHelper.WithTx(func(queries *database.Queries) error {
		userService := NewUserService(queries)
		authService := NewAuthService(NewAuthServiceInput{
			Cfg: testutils.GetTestApiConfig(),
			Db: queries,
			UserService: userService,
		})

		err := authService.UpgradeUser(testHelper.Ctx, UpgradeUserInput{
			UserID: uuid.New(),
		})
		require.Error(t, err)
		require.Equal(t, "user not found", err.Error())
		var notFoundErr *service_errors.NotFoundError
		require.True(t, errors.As(err, &notFoundErr))

		return nil
	})
}
