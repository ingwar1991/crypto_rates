FROM php:8.2-fpm AS crypto_rates_rest_api_image

# Install dependencies
RUN apt-get update && apt-get install -y \
    git unzip libpng-dev libjpeg-dev libfreetype6-dev \
    libicu-dev libonig-dev libzip-dev libxml2-dev zlib1g-dev \
    && docker-php-ext-install intl mbstring pdo pdo_mysql zip gd

# Install Composer
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html

# Copy application
COPY . .

# Install PHP dependencies
RUN composer install --no-interaction --optimize-autoloader

## tests will also run on startup, that may be more convenient
## int this case, because without `composer install` you won't be able to debug anyway

# Run tests and fail build if they fail
RUN php ./vendor/bin/phpunit

# Set permissions
RUN chown -R www-data:www-data /var/www/html/var /var/www/html/vendor

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
