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
1. `go run cmd/jvbe db create` - this will create a new Sqlite file (the value specified in `db_conn`) with the necessary tables

### run 
1. Build css: `npm run build --prefix ./ui`
1. Run web app: `go run cmd/jvbe app`

### run with live reloading
If you don't want to run the above commands every time a change occurs, you can do the following for live reloading
1. Install [air](https://github.com/cosmtrek/air): `go install github.com/cosmtrek/air@latest`
1. `air -c air.toml`
