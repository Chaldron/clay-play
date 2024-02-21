# jvbe 

## steps

### setup
1. Need at least go version 1.21.5
1. Initial download of go modules: `go mod download`
1. Create "config.yaml" file. Structure is as follows (contact me for what values to use for oauth):
    ```
    db_conn: ./test.db
    port: 8080

    oauth: 
      domain: oauth.domain.com
      client_id: oauthclientid
      client_secret: oauthclientsecret
      callback_url: http://localhost:8080/auth/callback
      logout_redirect_url: http://localhost:8080
    ```

### run 
1. Build css: `npm run build --prefix ./ui`
1. Run web app: `go run cmd/jvbe app`

### run with live reloading
If you don't want to run the above commands every time a change occurs, you can do the following for live reloading
1. Install [air](https://github.com/cosmtrek/air): `go install github.com/cosmtrek/air@latest`
1. `air -c .air.toml`

## db/migrations

Using SQlite as database. Migrations are handled through [goose](https://github.com/pressly/goose). We only use goose as a library rather than the CLI so no need to download the CLI.

- **Create a new migration:** `go run ./cmd/jvbe migration create <name>`
    - This will create a new migration file named something like `migrations/20240219151811_<name>.sql`, where you can put the migration details in
- On app startup, `goose.Up(...)` runs to always bring the DB schema up to date

## feature requests
If you want to request a new feature, please create a new issue on this repo
