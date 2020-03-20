# ezoauth

> Ez oauth login package (with gorm)

## Examples

### `examples/minimal`

Shows a minimal setup/config

This includes a dummy oauth server that can be started with `go run examples/minimal/oauth-server/server.go`

Then run the actual server with `go run examples/minimal/server.go`, the site will be available at http://localhost:8080, and you can log in with any username/password

### `examples/lego`

Shows a setup authenticating against [LEGO](https://github.com/webkom/lego/)

Requires you to first create an OAuth application with a redirect url to `http://localhost:8080/callback`, and set the id and secret as `CLIENT_ID` and `CLIENT_SECRET` environment variables.
