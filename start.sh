#!/bin/sh

# stop script execution when it encounters an error
set -e

# Load environment variables from app.env file
if [ -f /app/app.env ]; then
    set -a  # automatically export all variables
    . /app/app.env
    set +a  # turn off automatic export
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

