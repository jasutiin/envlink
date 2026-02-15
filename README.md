
# envlink

I work on two different computers and routinely need to transfer secrets between them. I used to paste those secrets in another application - whether that be through Discord chat or any other means - and delete them after copying and pasting. Other times, I regenerate credentials when I forgot them. This is error-prone, and sharing these secrets in other applications is not safe. envlink removes that friction by providing a centralized, secure place to store and sync your projects' .env files so you can access the right secrets on any machine without sharing them elsewhere.

### Tools used
- Go (Golang)
- Cobra: Go library for building CLI applications
- Gin: Go HTTP web framework for building APIs
- Supabase (PostgreSQL): Database for storing information
- Docker: To containerize the server code for deployment
- Google Cloud Run (tentative): To host the server

### Project structure

- `cmd/cli/main.go`: the CLI entrypoint that calls `cli.Execute()`
- `internal/cli/`: cobra-based command definitions
	- `root.go`: root command with flags and viper config
	- all other files in this directory are the commands for envlink
- `server/`: HTTP server code with API endpoints, controllers, and routers

### Available commands

- `login`: log in to envlink
- `register`: register an account
- `push`: push .env to storage
- `pull`: pull latest .env changes
- `projects`: list all stored .env files
- `store`: store your secret key

### How to run the CLI

From the repo root:

```powershell
go run ./cmd/cli {command}
```

From the CLI directory:

```powershell
cd cmd/cli
go run . {command}
```