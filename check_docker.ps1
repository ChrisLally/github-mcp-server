docker ps -a
Write-Host "`nRunning Docker container output:"
docker ps | Select-String github-mcp-server 