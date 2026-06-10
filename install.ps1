#Requires -Version 5.1
[CmdletBinding()]
param(
    [string]$Version = "v0.1.0",
    [string]$InstallDir,
    [string]$FileName = "frontend-mcp.exe",
    [switch]$Force,
    [switch]$SkipHashCheck,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

$Repo = "defineS6/frontend-solution-mcp"
$AssetName = "frontend-mcp-windows-amd64.exe"
$KnownHashes = @{
    "v0.1.0" = "8244EE88D87C0BBBAA2D92AFAF8DFDF3C6DF7F5656EBB73C2FE66BAA84D9B724"
}

function Write-Step {
    param([string]$Message)
    Write-Host "[frontend-mcp] $Message" -ForegroundColor Cyan
}

function Resolve-InstallDir {
    param([string]$InputPath)

    if (-not [string]::IsNullOrWhiteSpace($InputPath)) {
        return $InputPath
    }

    $localAppData = [Environment]::GetFolderPath("LocalApplicationData")
    if ([string]::IsNullOrWhiteSpace($localAppData)) {
        $localAppData = Join-Path $HOME "AppData\Local"
    }

    return Join-Path $localAppData "Programs\frontend-solution-mcp"
}

function Write-McpConfigExample {
    param([string]$CommandPath)

    $config = [ordered]@{
        mcpServers = [ordered]@{
            "frontend-mcp" = [ordered]@{
                command = $CommandPath
                env = [ordered]@{
                    FRONTEND_MCP_BASE_URL = "https://api.example.com/v1"
                    FRONTEND_MCP_API_KEY = "your api key"
                    FRONTEND_MCP_MODEL = "your model"
                    FRONTEND_MCP_TIMEOUT = "120s"
                }
                type = "stdio"
            }
        }
    }

    Write-Host ""
    Write-Step "MCP config example:"
    $config | ConvertTo-Json -Depth 8
}

if ([Environment]::OSVersion.Platform -ne "Win32NT") {
    throw "This install script only supports Windows."
}

$resolvedInstallDir = Resolve-InstallDir -InputPath $InstallDir
$targetPath = Join-Path $resolvedInstallDir $FileName
$downloadURL = "https://github.com/$Repo/releases/download/$Version/$AssetName"

Write-Step "Version: $Version"
Write-Step "Download URL: $downloadURL"
Write-Step "Install path: $targetPath"

if ($DryRun) {
    Write-Step "DryRun enabled. No files will be downloaded or written."
    Write-McpConfigExample -CommandPath $targetPath
    exit 0
}

if ((Test-Path -LiteralPath $targetPath) -and -not $Force) {
    Write-Step "Target file already exists. Use -Force to overwrite it."
    Write-McpConfigExample -CommandPath $targetPath
    exit 0
}

New-Item -ItemType Directory -Force -Path $resolvedInstallDir | Out-Null

$tempPath = Join-Path ([IO.Path]::GetTempPath()) ("frontend-mcp-" + [Guid]::NewGuid().ToString("N") + ".exe")

try {
    Write-Step "Downloading binary..."
    Invoke-WebRequest -Uri $downloadURL -OutFile $tempPath -UseBasicParsing

    if (-not $SkipHashCheck) {
        if ($KnownHashes.ContainsKey($Version)) {
            $actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $tempPath).Hash.ToUpperInvariant()
            $expectedHash = $KnownHashes[$Version].ToUpperInvariant()
            if ($actualHash -ne $expectedHash) {
                throw "SHA256 verification failed. Expected: $expectedHash, actual: $actualHash"
            }
            Write-Step "SHA256 verification passed."
        } else {
            Write-Warning "No built-in SHA256 for this version. Skipping verification."
        }
    }

    Move-Item -LiteralPath $tempPath -Destination $targetPath -Force
    Write-Step "Installed: $targetPath"
    Write-McpConfigExample -CommandPath $targetPath
}
finally {
    if (Test-Path -LiteralPath $tempPath) {
        Remove-Item -LiteralPath $tempPath -Force
    }
}
