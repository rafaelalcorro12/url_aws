Write-Host "Building module3-stats..." -ForegroundColor Green
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -o bootstrap main.go
Compress-Archive -Path bootstrap -DestinationPath function.zip -Force
Remove-Item bootstrap
Write-Host "Build complete!" -ForegroundColor Green
