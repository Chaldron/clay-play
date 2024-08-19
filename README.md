# Clay play

## Setup
1. Install go ≥ 1.21.5
2. Download go modules: `go mod download`
3. Create "config.yaml" file as follows:
    ```
    db_conn: ./test.db
    port: 8080
    ```

### Run 
1. Build CSS: `npm run build --prefix ./ui`
2. Run app: `go run ./cmd/jvbe`

### Run with live reloading
If you don't want to run the above commands every time a change occurs, you can use [air](https://github.com/air-verse/air) for live reloading.

1. Install `air` with `go install github.com/air-verse/air`
2. Run `air -c .air.toml`

## db/migrations

Using SQlite as database. Migrations are handled through [goose](https://github.com/pressly/goose). We only use goose as a library rather than the CLI so no need to download the CLI.

- **Create a new migration:** `go run ./cmd/jvbe migration create <name>`
    - This will create a new migration file named something like `migrations/20240219151811_<name>.sql`, where you can put the migration details in
- On app startup, `goose.Up(...)` runs to always bring the DB schema up to date
