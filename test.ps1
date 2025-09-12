$base = 'http://localhost:8080'
$user = 'alice'
$pass = 's3cret'

Write-Host "Using base: $base"

function PostJson($url, $body) {
    try {
        $r = Invoke-WebRequest -Uri $url -Method POST -ContentType 'application/json' -Body ($body | ConvertTo-Json -Compress) -UseBasicParsing -ErrorAction Stop
        return $r.Content | ConvertFrom-Json
    } catch {
        Write-Host "POST $url failed: $($_.Exception.Message)"
        return $null
    }
}

# call /profile without token
Write-Host "Calling /profile without token (should fail)..."
try {
    $prof = Invoke-WebRequest -Uri "$base/profile" -Method GET -UseBasicParsing -ErrorAction Stop
    Write-Host "Profile request without token unexpectedly succeeded. Response:`n$($prof.Content)"
    exit 1
} catch {
    Write-Host "Profile request without token failed as expected: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        try { $txt = $_.Exception.Response.GetResponseStream(); $sr = New-Object System.IO.StreamReader($txt); $respText = $sr.ReadToEnd(); Write-Host "Response body: $respText" } catch {}
    }
}

# register
Write-Host "Registering user $user..."
$reg = PostJson "$base/register" @{ username = $user; password = $pass }
if ($null -eq $reg) {
    Write-Host "Register request failed or returned non-JSON. Continuing..."
} else {
    Write-Host "Register response:`n$($reg | ConvertTo-Json -Depth 3)"
}

# login
Write-Host "Logging in..."
$login = PostJson "$base/login" @{ username = $user; password = $pass }
if ($null -eq $login -or -not $login.token) {
    Write-Host "Login failed or no token returned. Response:`n$($login | ConvertTo-Json -Depth 3)"
    exit 1
}
$token = $login.token
Write-Host "Got token: $token"

# call /profile
Write-Host "Calling /profile with token..."
try {
    $prof = Invoke-WebRequest -Uri "$base/profile" -Method GET -Headers @{ Authorization = "Bearer $token" } -UseBasicParsing -ErrorAction Stop
    $profBody = $prof.Content | ConvertFrom-Json
    Write-Host "Profile response:`n$($profBody | ConvertTo-Json -Depth 3)"
} catch {
    Write-Host "Profile request failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        try { $txt = $_.Exception.Response.GetResponseStream(); $sr = New-Object System.IO.StreamReader($txt); $respText = $sr.ReadToEnd(); Write-Host "Response body: $respText" } catch {}
    }
    exit 1
}

Write-Host "Test completed successfully."
