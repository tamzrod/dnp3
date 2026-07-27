# DNP3 Engineering Workbench - Build Script
# PowerShell script for Windows (run with PowerShell 5.1+)
# Usage: .\build-workbench.ps1 [-Action {build|run|clean}]

param(
    [ValidateSet("build", "run", "clean", "all")]
    [string]$Action = "build"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$WorkbenchDir = Join-Path $ProjectRoot "cmd/workbench"
$OutputDir = Join-Path $WorkbenchDir "bin"

function Write-Step {
    param([string]$Message)
    Write-Host "[*] $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "[+] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "[-] $Message" -ForegroundColor Red
}

function Check-Go {
    Write-Step "Checking for Go installation..."
    try {
        $goVersion = go version
        Write-Success "Found: $goVersion"
        return $true
    }
    catch {
        Write-Error "Go is not installed or not in PATH"
        Write-Host "Please install Go 1.22+ from https://go.dev/dl/" -ForegroundColor Yellow
        return $false
    }
}

function Restore-Dependencies {
    Write-Step "Restoring dependencies..."
    Push-Location $WorkbenchDir
    try {
        go mod tidy
        if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }
        Write-Success "Dependencies restored"
    }
    finally {
        Pop-Location
    }
}

function Build-Application {
    Write-Step "Building DNP3 Workbench..."
    Push-Location $WorkbenchDir
    try {
        # Create output directory
        if (-not (Test-Path $OutputDir)) {
            New-Item -ItemType Directory -Path $OutputDir | Out-Null
        }

        # Build for Windows
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        $OutputPath = Join-Path $OutputDir "workbench.exe"
        
        go build -ldflags="-s -w" -o $OutputPath .
        
        if ($LASTEXITCODE -ne 0) { throw "Build failed" }
        
        Write-Success "Built: $OutputPath"
        return $OutputPath
    }
    finally {
        Pop-Location
        $env:GOOS = $null
        $env:GOARCH = $null
    }
}

function Run-Application {
    param([string]$BinaryPath)
    
    Write-Step "Launching DNP3 Workbench..."
    Write-Host "Press Ctrl+C to stop the application" -ForegroundColor Gray
    
    try {
        & $BinaryPath
    }
    finally {
        Write-Step "Application closed"
    }
}

function Clean-Build {
    Write-Step "Cleaning build artifacts..."
    Push-Location $WorkbenchDir
    try {
        if (Test-Path $OutputDir) {
            Remove-Item -Recurse -Force $OutputDir
            Write-Success "Cleaned output directory"
        }
        go clean
        Write-Success "Build artifacts removed"
    }
    finally {
        Pop-Location
    }
}

# Main execution
Write-Host ""
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "  DNP3 Engineering Workbench Builder" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta
Write-Host ""

# Check prerequisites
if (-not (Check-Go)) {
    exit 1
}

switch ($Action) {
    "build" {
        Restore-Dependencies
        $binary = Build-Application
        Write-Host ""
        Write-Success "Build complete!"
        Write-Host "Executable: $binary" -ForegroundColor Gray
    }
    
    "run" {
        Restore-Dependencies
        $binary = Build-Application
        Write-Host ""
        Run-Application -BinaryPath $binary
    }
    
    "clean" {
        Clean-Build
    }
    
    "all" {
        Clean-Build
        Restore-Dependencies
        $binary = Build-Application
        Write-Host ""
        Write-Success "All done!"
    }
}

Write-Host ""
