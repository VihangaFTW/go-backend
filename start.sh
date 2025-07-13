#!/bin/sh

# stop script execution when it encounters an error
set -e

# Load environment variables from app.env file
#! NOTE: When we source a file in bash/sh, it only sets variables that don't already exist. Since Docker compose file pre-sets DB_SOURCE before the api container starts,that env variable is NOT overwriiten.
#* DB_SOURCE still points to the internal docker postgres service instead of localhost. Otherwise,
#* the DB_SOURCE would point to localhost as defined in env file and migrations would fail
if [ -f /app/app.env ]; then
    . /app/app.env
fi

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

