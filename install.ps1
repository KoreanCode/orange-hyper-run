$ErrorActionPreference = "Stop"

$repo = if ($env:HYPER_RUN_UPDATE_SOURCE) { $env:HYPER_RUN_UPDATE_SOURCE } else { "KoreanCode/orange-hyper-run" }
$installDir = if ($env:HYPER_INSTALL_DIR) { $env:HYPER_INSTALL_DIR } else { Join-Path $env:USERPROFILE ".local\bin" }
$asset = "hyper-windows-amd64.exe"

if ($repo -match "^https?://") {
    $url = $repo
    $checksumUrl = if ($env:HYPER_RUN_CHECKSUM_URL) { $env:HYPER_RUN_CHECKSUM_URL } else { "" }
    $asset = Split-Path -Leaf ([Uri]$url).AbsolutePath
}
else {
    if ($repo.StartsWith("github:")) {
        $repo = $repo.Substring("github:".Length)
    }
    $url = "https://github.com/$repo/releases/latest/download/$asset"
    $checksumUrl = "https://github.com/$repo/releases/latest/download/checksums.txt"
}

New-Item -ItemType Directory -Force $installDir | Out-Null

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("hyper-install-{0}.exe" -f $PID)
$checksumsTmp = Join-Path ([System.IO.Path]::GetTempPath()) ("hyper-install-{0}.checksums.txt" -f $PID)
$target = Join-Path $installDir "hyper.exe"

try {
    Write-Host "Installing Hyper Run from $url"
    Invoke-WebRequest -Uri $url -OutFile $tmp

    if ($checksumUrl) {
        Write-Host "Verifying checksum from $checksumUrl"
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumsTmp
        $expected = $null
        foreach ($line in Get-Content $checksumsTmp) {
            $parts = $line -split "\s+"
            if ($parts.Length -ge 2 -and $parts[1] -eq $asset) {
                $expected = $parts[0]
                break
            }
        }
        if (-not $expected) {
            throw "Checksum not found for $asset in checksums.txt"
        }

        $actual = (Get-FileHash -Algorithm SHA256 $tmp).Hash.ToLowerInvariant()
        if ($actual -ne $expected.ToLowerInvariant()) {
            throw "Checksum mismatch for $asset. Expected $expected, got $actual"
        }
    }

    Move-Item -Force $tmp $target
    Write-Host "Installed: $target"

    $pathParts = $env:PATH -split ";"
    if ($pathParts -notcontains $installDir) {
        Write-Warning "$installDir is not on PATH"
        Write-Host "Add it to the user PATH with:"
        Write-Host "[Environment]::SetEnvironmentVariable(""Path"", `$env:Path + "";$installDir"", ""User"")"
    }

    & $target version
}
finally {
    if (Test-Path $tmp) {
        Remove-Item -Force $tmp
    }
    if (Test-Path $checksumsTmp) {
        Remove-Item -Force $checksumsTmp
    }
}
