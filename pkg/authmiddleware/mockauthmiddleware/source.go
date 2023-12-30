package mockauthmiddleware

//go:generate mockgen -destination=./auth.go -package=mockauthmiddleware bitmex-api/pkg/authmiddleware AuthMiddleware
