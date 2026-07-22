param(
    [string]$EnvFile = "..\manager\.env"
)

$ErrorActionPreference = "Stop"
$resolvedEnvFile = (Resolve-Path -LiteralPath $EnvFile).Path
$databaseLine = Get-Content -LiteralPath $resolvedEnvFile |
    Where-Object { $_ -match '^\s*DATABASE_URL\s*=' } |
    Select-Object -First 1

if (-not $databaseLine) {
    throw "DATABASE_URL was not found in $resolvedEnvFile"
}

$databaseUrl = ($databaseLine -replace '^\s*DATABASE_URL\s*=\s*', '').Trim()
if (($databaseUrl.StartsWith('"') -and $databaseUrl.EndsWith('"')) -or
    ($databaseUrl.StartsWith("'") -and $databaseUrl.EndsWith("'"))) {
    $databaseUrl = $databaseUrl.Substring(1, $databaseUrl.Length - 2)
}
if (-not $databaseUrl) {
    throw "DATABASE_URL is empty in $resolvedEnvFile"
}

$previousValue = $env:QUOTA_TEST_DATABASE_URL
try {
    $env:QUOTA_TEST_DATABASE_URL = $databaseUrl
    go test -tags=integration ./internal/authz -run '^TestQuotaStrategyIntegration$' -count=1 -v
    if ($LASTEXITCODE -ne 0) {
        throw "Quota strategy integration test failed with exit code $LASTEXITCODE"
    }
} finally {
    if ($null -eq $previousValue) {
        Remove-Item Env:QUOTA_TEST_DATABASE_URL -ErrorAction SilentlyContinue
    } else {
        $env:QUOTA_TEST_DATABASE_URL = $previousValue
    }
}
