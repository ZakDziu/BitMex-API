package postgresstore

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bitmex-api/pkg/config"
	crmLog "bitmex-api/pkg/logger"
)

type PostgresStore struct {
	DB *gorm.DB

	UserRepository *UserRepository
	AuthRepository *AuthRepository
}

//nolint:nosprintfhostport
func dbURL(dbConfig *config.DBPostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)
}

func NewPostgresStore(cfg *config.DBPostgresConfig) (*PostgresStore, error) {
	db, err := gorm.Open(postgres.Open(dbURL(cfg)),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

	if err != nil {
		crmLog.Fatalf("NewPostgresStore: %s", err)
	}

	store := &PostgresStore{DB: db}

	return store, nil
}

func (s *PostgresStore) User() *UserRepository {
	if s.UserRepository == nil {
		s.UserRepository = NewUserRepository(s)
	}

	return s.UserRepository
}

func (s *PostgresStore) Auth() *AuthRepository {
	if s.AuthRepository == nil {
		s.AuthRepository = NewAuthRepository(s)
	}

	return s.AuthRepository
}
