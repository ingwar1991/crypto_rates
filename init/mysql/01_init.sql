CREATE DATABASE IF NOT EXISTS crypto_rates;

CREATE USER IF NOT EXISTS 'user'@'%' IDENTIFIED BY 'pass';
GRANT ALL PRIVILEGES ON crypto_rates.* TO 'user'@'%';
FLUSH PRIVILEGES;

USE crypto_rates;

CREATE TABLE IF NOT EXISTS latest_rates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pair VARCHAR(10) NOT NULL,
    rate DECIMAL(18,8) NOT NULL,
    time DATETIME NOT NULL,
    UNIQUE INDEX idx_pair_time (pair, time)
);

CREATE TABLE IF NOT EXISTS avg_rates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pair VARCHAR(10) NOT NULL,
    rate DECIMAL(18,8) NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    UNIQUE INDEX idx_pair_starttime (pair, start_time)
);
