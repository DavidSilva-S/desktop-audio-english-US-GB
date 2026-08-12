$ErrorActionPreference = "Stop"

$GITHUB_REPO = "DavidSilva-S/desktop-audio-english-US-GB"
$BINARY_NAME = "desktop-audio-english-US-GB"
$TAG         = "language"

$INSTALL_DIR = Join-Path $env:LOCALAPPDATA "Programs\$BINARY_NAME"
$DATA_DIR    = Join-Path $env:LOCALAPPDATA "$BINARY_NAME"
$PIPER_DIR   = Join-Path $DATA_DIR "piper"
$VOICES_DIR  = Join-Path $DATA_DIR "piper-voices"
$AUDIOS_DIR  = Join-Path $DATA_DIR "audios"

function Write-Info($msg) { Write-Host "[info] $msg" -ForegroundColor Cyan }
function Write-Ok($msg)   { Write-Host "[ok] $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[warning] $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "[error] $msg" -ForegroundColor Red; exit 1 }

$TMP_DIR = Join-Path $env:TEMP ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TMP_DIR -Force | Out-Null

try {
    $EXE_ASSET = "${BINARY_NAME}.exe"
    $DOWNLOAD_URL = "https://github.com/$GITHUB_REPO/releases/download/$TAG/$EXE_ASSET"
    $TEMP_EXE = Join-Path $TMP_DIR $EXE_ASSET

    Write-Info "Downloading $EXE_ASSET..."
    try {
        Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $TEMP_EXE -UseBasicParsing
    } catch {
        Write-Err "Failed to download $DOWNLOAD_URL. Check the release artifacts for tag $TAG."
    }

    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    Copy-Item -Path $TEMP_EXE -Destination (Join-Path $INSTALL_DIR $EXE_ASSET) -Force
    Write-Ok "Executable installed at: $INSTALL_DIR\$EXE_ASSET"

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -split ";" -notcontains $INSTALL_DIR) {
        Write-Info "Adding $INSTALL_DIR to user PATH..."
        $NewPath = "$UserPath;$INSTALL_DIR"
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        $env:Path = "$env:Path;$INSTALL_DIR"
        Write-Ok "PATH updated successfully."
    } else {
        Write-Ok "$INSTALL_DIR is already in PATH."
    }

    New-Item -ItemType Directory -Path $PIPER_DIR -Force | Out-Null
    New-Item -ItemType Directory -Path $VOICES_DIR -Force | Out-Null
    New-Item -ItemType Directory -Path $AUDIOS_DIR -Force | Out-Null

    function Download-And-Extract-Asset($ZipFileName, $DestDir, $Label) {
        $UrlZip = "https://github.com/$GITHUB_REPO/releases/download/$TAG/$ZipFileName"
        $OutZip = Join-Path $TMP_DIR $ZipFileName

        Write-Info "Downloading $Label ($ZipFileName)..."
        try {
            Invoke-WebRequest -Uri $UrlZip -OutFile $OutZip -UseBasicParsing
            Expand-Archive -Path $OutZip -DestinationPath $DestDir -Force
            Write-Ok "$Label installed to $DestDir"
        } catch {
            Write-Warn "Failed to download or extract $Label ($ZipFileName)."
        }
    }

    Download-And-Extract-Asset -ZipFileName "piper_windows_amd64.zip" -DestDir $PIPER_DIR -Label "Piper"
    Download-And-Extract-Asset -ZipFileName "piper_voices_amd64.zip" -DestDir $VOICES_DIR -Label "Voices"

    Write-Ok "Installation completed successfully!"
    Write-Info "Restart your terminal or PowerShell for '$BINARY_NAME' to become available."

} finally {
    if (Test-Path $TMP_DIR) {
        Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
    }
}
