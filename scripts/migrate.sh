# Run migrations using docker compose
set -e

docker compose up -d postgres
sleep 2
PGPASSWORD=postgres psql -h localhost -U postgres -d xo -f ./migrations/001_init.sql
