package testdb

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type Generator interface {
	GenerateUser(ctx context.Context) *database.User
	GenerateUsers(ctx context.Context, count int) []*database.User
	GenerateChirp(ctx context.Context, userID uuid.UUID) *database.Chirp
	GenerateChirps(ctx context.Context, userID uuid.UUID, count int) []*database.Chirp
}

type generator struct {
	Db    *database.Queries
	Faker *gofakeit.Faker
	Seed  uint64
	T     *testing.T
}

type NewGeneratorInput struct {
	Db   *database.Queries
	Seed uint64
	T    *testing.T
}

func NewGenerator(input NewGeneratorInput) *generator {
	return &generator{
		Db:    input.Db,
		Faker: gofakeit.New(input.Seed),
		Seed:  input.Seed,
		T:     input.T,
	}
}

func (g *generator) GenerateUser(ctx context.Context) *database.User {
	userParam := GenerateOne[database.CreateUserParams](g.T, g.Faker)

	user, err := g.Db.CreateUser(ctx, userParam)
	require.NoError(g.T, err)

	return &user
}

func (g *generator) GenerateUsers(ctx context.Context, count int) []*database.User {
	userParams := GenerateMany[database.CreateUserParams](g.T, g.Faker, count)

	users := make([]*database.User, count)
	for i, param := range userParams {
		user, err := g.Db.CreateUser(ctx, param)
		require.NoError(g.T, err)
		users[i] = &user
	}

	return users
}

func (g *generator) GenerateChirp(ctx context.Context, userID uuid.UUID) *database.Chirp {
	chirpParam := GenerateOne[database.CreateChirpParams](g.T, g.Faker)
	chirpParam.UserID = userID

	chirp, err := g.Db.CreateChirp(ctx, chirpParam)
	require.NoError(g.T, err)

	return &chirp
}

func (g *generator) GenerateChirps(ctx context.Context, userID uuid.UUID, count int) []*database.Chirp {
	chirpParams := GenerateMany[database.CreateChirpParams](g.T, g.Faker, count)

	chirps := make([]*database.Chirp, count)
	for i, param := range chirpParams {
		param.UserID = userID
		chirp, err := g.Db.CreateChirp(ctx, param)
		require.NoError(g.T, err)
		chirps[i] = &chirp
	}

	return chirps
}

func GenerateOne[T any](t *testing.T, faker *gofakeit.Faker) T {
	// Create placeholder for the value
	var generatedValue T

	// Generate the struct using gofakeit
	err := faker.Struct(&generatedValue)
	// Assert that the generation was successful
	require.NoError(t, err)

	// Return the generated value
	return generatedValue
}

func GenerateMany[T any](t *testing.T, faker *gofakeit.Faker, count int) []T {
	// Create placeholder for the value slice
	generatedValues := make([]T, count)

	for i := range generatedValues {
		err := faker.Struct(&generatedValues[i])
		require.NoError(t, err)
	}

	// Return the generated values
	return generatedValues
}
