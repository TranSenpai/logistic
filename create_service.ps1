param (
    [Parameter(Mandatory=$true, HelpMessage="Service name")]
    [string]$ServiceName
)

Write-Host "Creating new service: $ServiceName..." -ForegroundColor Cyan

# 1. Create directories
$dirs = @(
    "$ServiceName/cmd",
    "$ServiceName/internal/delivery",
    "$ServiceName/internal/biz",
    "$ServiceName/internal/repo",
    "$ServiceName/internal/entity",
    "$ServiceName/internal/dto",
    "$ServiceName/internal/di"
)

foreach ($dir in $dirs) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    Write-Host "  Created directory: $dir" -ForegroundColor Green
}

# 2. Create cmd/main.go
$mainGoContent = @"
package main

import (
	"log"
)

func main() {
	log.Println("Starting ${ServiceName}...")
	app := NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run ${ServiceName}: %v", err)
	}
}
"@
Set-Content -Path "$ServiceName/cmd/main.go" -Value $mainGoContent
Write-Host "  Created file: $ServiceName/cmd/main.go" -ForegroundColor Green

# 3. Create cmd/app.go
$appGoContent = @"
package main

import "log"

type App struct {
	// Declare app dependencies here
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	log.Println("${ServiceName} is running!")
	// Initialize HTTP or gRPC server here
	return nil
}
"@
Set-Content -Path "$ServiceName/cmd/app.go" -Value $appGoContent
Write-Host "  Created file: $ServiceName/cmd/app.go" -ForegroundColor Green

# 4. Init go.mod
Write-Host "Initializing go.mod..." -ForegroundColor Cyan
Push-Location $ServiceName
go mod init $ServiceName
Pop-Location

# 5. Add to go.work
if (Test-Path "go.work") {
    Write-Host "Adding $ServiceName to go.work..." -ForegroundColor Cyan
    go work use "./$ServiceName"
}

Write-Host "Done! You can now start coding in $ServiceName." -ForegroundColor Yellow
