# Запуск: .\test.ps1
# Сервер должен быть запущен: go run ./cmd/server

$BASE = "http://localhost:8010/api/v1/subscriptions"
$SEP  = "-----------------------------------------------------"
$USER_ID = [guid]::NewGuid().ToString()

Write-Host "Тестовый user_id: $USER_ID"

function Print-JSON($obj) {
    $obj | ConvertTo-Json -Depth 10
}

function Invoke-API($url, $method, $body = $null) {
    try {
        $params = @{ Uri = $url; Method = $method; ErrorAction = "Stop" }
        if ($body) {
            $params.ContentType = "application/json"
            $params.Body = ($body | ConvertTo-Json)
        }
        $resp = Invoke-WebRequest @params
        return [PSCustomObject]@{ Status = $resp.StatusCode; Data = ($resp.Content | ConvertFrom-Json) }
    } catch {
        $code = if ($_.Exception.Response) { $_.Exception.Response.StatusCode.value__ } else { 0 }
        $content = $_.ErrorDetails.Message
        $data = if ($content) { try { $content | ConvertFrom-Json } catch { @{error = $content} } } else { @{error = $_.Exception.Message} }
        return [PSCustomObject]@{ Status = $code; Data = $data }
    }
}

function Post-JSON($url, $body) { return Invoke-API $url "POST" $body }
function Get-URL($url)          { return Invoke-API $url "GET" }
function Put-JSON($url, $body)  { return Invoke-API $url "PUT" $body }

function Delete-URL($url) {
    $r = Invoke-API $url "DELETE"
    return $r.Status
}

function Get-Status($url) {
    $r = Invoke-API $url "GET"
    return $r.Status
}

function Delete-URL($url) {
    try {
        $resp = Invoke-WebRequest -Uri $url -Method DELETE -ErrorAction Stop
        return $resp.StatusCode
    } catch {
        return $_.Exception.Response.StatusCode.value__
    }
}

function Get-Status($url) {
    try {
        $resp = Invoke-WebRequest -Uri $url -Method GET -ErrorAction Stop
        return $resp.StatusCode
    } catch {
        return $_.Exception.Response.StatusCode.value__
    }
}

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 1. CREATE subscription ======================="
$r = Post-JSON $BASE @{
    service_name = "Netflix"
    price        = 990
    user_id      = $USER_ID
    start_date   = "01-2025"
    end_date     = "06-2025"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data
$ID = $r.Data.id
Write-Host ">>> Полученный ID: $ID"

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 2. CREATE ещё одну (без end_date) ==========="
$r = Post-JSON $BASE @{
    service_name = "Spotify"
    price        = 299
    user_id      = $USER_ID
    start_date   = "03-2025"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 2.1 DUPLICATE PERIOD (ожидаем 409) ==========="
$r = Post-JSON $BASE @{
    service_name = "Spotify"
    price        = 399
    user_id      = $USER_ID
    start_date   = "04-2025"
    end_date     = "05-2025"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 3. LIST (все) ================================"
$r = Get-URL $BASE
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 4. LIST (фильтр по user_id + пагинация) ======"
$r = Get-URL "$BASE`?user_id=$USER_ID&limit=10&offset=0"
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 5. GET BY ID ================================="
$r = Get-URL "$BASE/$ID"
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 6. UPDATE (меняем цену и end_date) =========="
$r = Put-JSON "$BASE/$ID" @{
    price     = 1190
    end_date  = "12-2025"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 6.1 UPDATE (сбрасываем end_date в null) ======"
$r = Put-JSON "$BASE/$ID" @{
    end_date = $null
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 7. TOTAL (янв-июнь 2025) ===================="
$r = Get-URL "$BASE/total?user_id=$USER_ID&from=01-2025&to=06-2025"
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 8. DELETE ===================================="
$code = Delete-URL "$BASE/$ID"
Write-Host "HTTP статус: $code"

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 9. GET BY ID после удаления (ожидаем 404) ===="
$r = Get-URL "$BASE/$ID"
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

# ────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "===== 10. ОШИБКИ - невалидные данные ==============="

Write-Host $SEP
Write-Host "bad price (0):"
$r = Post-JSON $BASE @{
    service_name = "Test"; price = 0
    user_id = $USER_ID; start_date = "01-2025"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

Write-Host $SEP
Write-Host "bad date format (2025-01 вместо 01-2025):"
$r = Post-JSON $BASE @{
    service_name = "Test"; price = 500
    user_id = $USER_ID; start_date = "2025-01"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

Write-Host $SEP
Write-Host "end_date раньше start_date:"
$r = Post-JSON $BASE @{
    service_name = "Test"; price = 500
    user_id = $USER_ID
    start_date = "06-2025"; end_date = "01-2025"
}
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

Write-Host $SEP
Write-Host "total без from/to:"
$r = Get-URL "$BASE/total"
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

Write-Host $SEP
Write-Host "total: to раньше from:"
$r = Get-URL "$BASE/total?from=06-2025&to=01-2025"
Write-Host "HTTP $($r.Status)"; Print-JSON $r.Data

Write-Host ""
Write-Host "===== ГОТОВО ======================================="
