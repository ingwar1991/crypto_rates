#!/bin/sh

# exit if any fails
set -e

# Run tests before starting services
echo "Running tests..."
php ./vendor/bin/phpunit

echo "Starting PHP-FPM and Messenger consumer..."
php-fpm &

exec bin/console messenger:consume scheduler_default
