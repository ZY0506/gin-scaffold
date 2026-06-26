@echo off
REM 项目快捷命令

if "%1"=="run" (
    go run ./cmd/server
    goto :eof
)
if "%1"=="build" (
    go build -o ./tmp/gin-scaffold.exe ./cmd/server
    goto :eof
)
if "%1"=="tidy" (
    go mod tidy
    goto :eof
)
if "%1"=="fmt" (
    go fmt ./...
    goto :eof
)
if "%1"=="test" (
    go test -v -cover ./...
    goto :eof
)
if "%1"=="swag" (
    swag init -g cmd/server/main.go --output docs/swagger
    goto :eof
)

REM 默认：运行服务
go run ./cmd/server
