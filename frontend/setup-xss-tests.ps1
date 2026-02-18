# XSS Testing Setup Script

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "XSS Prevention Testing Setup" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the frontend directory
if (-not (Test-Path "package.json")) {
    Write-Host "Error: package.json not found!" -ForegroundColor Red
    Write-Host "Please run this script from the frontend directory." -ForegroundColor Red
    exit 1
}

Write-Host "Step 1: Installing Playwright browsers..." -ForegroundColor Green
npx playwright install chromium firefox webkit

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Playwright browsers installed successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to install Playwright browsers" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 2: Verifying test files..." -ForegroundColor Green

$testFiles = @(
    "e2e\xss-stored.spec.ts",
    "e2e\xss-stored-simple.spec.ts",
    "e2e\helpers.ts",
    "playwright.config.ts"
)

$allFilesExist = $true
foreach ($file in $testFiles) {
    if (Test-Path $file) {
        Write-Host "  ✓ $file" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $file (missing)" -ForegroundColor Red
        $allFilesExist = $false
    }
}

if (-not $allFilesExist) {
    Write-Host ""
    Write-Host "Some test files are missing!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 3: Checking backend services..." -ForegroundColor Green

# Check if docker compose is running
try {
    $dockerOutput = docker compose ps 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Docker Compose is available" -ForegroundColor Green
        
        # Check if services are running
        $runningServices = docker compose ps --filter "status=running" --quiet
        if ($runningServices) {
            Write-Host "  ✓ Backend services are running" -ForegroundColor Green
        } else {
            Write-Host "  ⚠ Backend services are not running" -ForegroundColor Yellow
            Write-Host "    Run 'docker compose up -d' from the root directory" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "  ⚠ Could not check Docker Compose status" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Ensure backend services are running:" -ForegroundColor White
Write-Host "   cd .. (to root directory)" -ForegroundColor Gray
Write-Host "   docker compose up -d" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Run the XSS tests:" -ForegroundColor White
Write-Host "   npm run test:e2e:xss" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Or run in interactive UI mode:" -ForegroundColor White
Write-Host "   npm run test:e2e:ui" -ForegroundColor Gray
Write-Host ""
Write-Host "For more information, see:" -ForegroundColor White
Write-Host "  - e2e\QUICKSTART.md" -ForegroundColor Gray
Write-Host "  - e2e\README.md" -ForegroundColor Gray
Write-Host "  - XSS_TESTING.md" -ForegroundColor Gray
Write-Host ""
