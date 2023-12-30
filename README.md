# Welcome to CRM System!

## To start the project you need to do the following steps:
1. Create access and refresh keys
    ```
    openssl ecparam -name prime256v1 -genkey -noout -out ec-prime256v1-acc-priv-key.pem

    openssl ecparam -name prime256v1 -genkey -noout -out ec-prime256v1-ref-priv-key.pem
   ```
2. Set up .env, for example
    ```dotenv
   POSTGRES_HOST=postgres
   POSTGRES_PORT=5432
   POSTGRES_USER=postgres
   POSTGRES_DB=postgres
   POSTGRES_PASSWORD=postgres
   SERVER_PORT=:8000
   API_URL=localhost:8000
   HASH_KEY_ACCESS=ec-prime256v1-acc-priv-key.pem
   HASH_KEY_REFRESH=ec-prime256v1-ref-priv-key.pem
   ```
3. Run ``docker-compose up`` to start the project

## After server start on 8000 port and postgres on 5432 port
1. Check out Swagger API documentation at the link ``http://localhost:8000/docs/index.html``
2. To register new users - use Tech Admin credentials
   ```json
   {
    "username": "techAdmin",
   "password": "techAdminPassword"
   }
   ```
3. If you want to use token - in authorization add *Bearer* to the start of token


## For DB tests put *.env* file in pkg/store/posgresstore
```dotenv
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_DBNAME=postgres
POSTGRES_PASSWORD=postgres

```
To check test coverage run
``go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html && open coverage.html``