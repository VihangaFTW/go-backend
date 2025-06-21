#!/bin/sh

# stop script execution when it encounters an error
set -e

. app.env
# The $DB_SOURCE environment variable is available here because:
# 1. It's defined in docker-compose.yaml under the 'api' service environment section
#* 2. Docker automatically makes environment variables available to the container at runtime
# 3. When this script runs inside the Docker container, it inherits all environment variables set by Docker Compose

echo "running db migration..."
# run the migration files to populate the db with the tables
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up


echo "starting the app..."
# Execute the command passed as arguments to this script
# In the Dockerfile, CMD ["/app/main"] becomes the arguments ($@)
# so final command becomes exec exec /app/main
#! When exec is used, the shell running the script is replaced by the command, 
#! meaning the command runs directly in place of the shell, not as a child process.
exec "$@"

