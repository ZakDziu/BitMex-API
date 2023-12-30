package mockpostgresstore

//nolint:lll
//go:generate mockgen -destination=./store.go -package=mockpostgresstore bitmex-api/pkg/store UserRepository,AuthRepository
