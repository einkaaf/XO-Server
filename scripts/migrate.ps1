# Run migrations using docker compose

docker compose up -d postgres
Start-Sleep -Seconds 2
$env:PGPASSWORD="postgres"
psql -h localhost -U postgres -d xo -f .\migrations\001_init.sql
