@echo off
REM Build script for Liberty Unleashed Server Directory (Windows)
REM This script builds the application for multiple platforms

set APP_NAME=lusd
if "%VERSION%"=="" set VERSION=dev
set BUILD_TIME=%date:~-4,4%-%date:~-10,2%-%date:~-7,2%_%time:~0,2%:%time:~3,2%:%time:~6,2%
set BUILD_TIME=%BUILD_TIME: =0%

REM Get commit hash
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set COMMIT_HASH=%%i
if "%COMMIT_HASH%"=="" set COMMIT_HASH=unknown

echo Building %APP_NAME% version %VERSION%
echo Build time: %BUILD_TIME%
echo Commit: %COMMIT_HASH%

REM Create build directory
if not exist build mkdir build

REM Build flags
set LDFLAGS=-s -w -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.CommitHash=%COMMIT_HASH%

REM Build for different platforms
echo Building for Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o build/%APP_NAME%-linux-amd64 cmd/lusd/main.go

echo Building for Windows amd64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o build/%APP_NAME%-windows-amd64.exe cmd/lusd/main.go

echo Building for macOS amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o build/%APP_NAME%-darwin-amd64 cmd/lusd/main.go

echo Build completed successfully!
echo Binaries available in build/ directory:
dir build\
