# 🪙 Crypto Rates Platform

A modular, containerized platform for collecting, aggregating, and serving cryptocurrency price data with authentication and logging. 
Built with Go, PHP/Symfony, MongoDB, Redis and MySQL


## 🚀 Overview

This project consists of four main services:

✅ Auth Service (Go + MongoDB)<br/>
&nbsp;&nbsp;🔐 Handles API key registration, JWT issuance, and logging <br/>
&nbsp;&nbsp;🗄️ MongoDB for storing credentials and logs <br/>
&nbsp;&nbsp;↔️ Communicates with REST API for authentication and logging<br/>

<br/>

✅ Worker Service (Go + Redis)<br/>
&nbsp;&nbsp;📡 Streams live trade ticks from Binance WebSocket API <br/>
&nbsp;&nbsp;🧠 Stores ticks in Redis<br/>
&nbsp;&nbsp;➡️ Feeds data into Aggregator Service<br/>

<br/>

✅ Aggregator Service (Go + Redis)<br/>
&nbsp;&nbsp;📊 Aggregates ticks into OHLCV candles every 10 seconds <br/>
&nbsp;&nbsp;🔁 Uses Redis for input/output <br/>
&nbsp;&nbsp;➡️ Supplies data to REST API Service<br/>

<br/>

✅ REST API Service (PHP/Symfony + MySQL)<br/>
&nbsp;&nbsp;🌐 Provides endpoints for latest/historical price data <br/>
&nbsp;&nbsp;🧮 Calculates VWAP, fetches prices from Redis or Binance REST API <br/>
&nbsp;&nbsp;🗄️ Stores results in MySQL <br/>
&nbsp;&nbsp;🔐 Requires JWT from Auth Service <br/>
&nbsp;&nbsp;📝 Logs API calls back to Auth Service<br/>

<br/>

🧩 All services are containerized and orchestrated via Docker Compose.<br/>

#### Ports
All ports are set at .env in root dir
Default values are
```.env
AUTH_APP_PORT=8081
REST_API_APP_PORT=8082
```

## 🧩 Services Breakdown

### 🔐 Auth Service

- **Tech Stack**: Go, MongoDB
- **Features**:
  - API key registration via simple HTML page
  - OTP via SMTP (or displayed directly for testing)
  - JWT issuance and refresh
  - REST endpoint for logging API calls

#### Example Usage

```bash
# Get JWT token
curl -X POST http://localhost:8081/token \
     -d "api_key=68b6d4763d69baec1d2a4970" \
     -H "Content-Type: application/x-www-form-urlencoded"

# Refresh JWT token
curl -X POST http://localhost:8081/token/refresh \
     -H "Authorization: Bearer <your_token>"

# Log an API call
curl -X POST http://localhost:8081/log/rest \
     -H "Authorization: Bearer <your_token>" \
     -H "Content-Type: application/json" \
     -d '{
           "endpoint": "test_event",
           "params": {"key": "value"},
           "response_status": 200,
           "timestamp": 1693400000
         }'
```

---

### 📡 Worker Service

- **Tech Stack**: Go, Redis
- **Function**: Connects to Binance WebSocket API and stores trade ticks in Redis.
- **Symbols Tracked**: Can be specified in config.yml. 
Default: `BTCEUR`, `ETHEUR`, `LTCEUR`

---

### 📊 Aggregator Service

- **Tech Stack**: Go, Redis
- **Function**: Aggregates trade ticks into OHLCV candles every 10 seconds.

---

### 🌐 REST API Service

- **Tech Stack**: PHP/Symfony, MySQL, Redis
- **Functionality**:
  - Scheduled tasks (every 5 minutes using symfony/scheduler):
    - Fetch latest prices from Redis or Binance REST API
    - Calculate VWAP from Redis or Binance historical data
  - Stores results in MySQL
  - Provides REST endpoints for accessing price data
  - Requires JWT authentication via Auth Service
  - Logs each API call to Auth Service

#### API Routes

| Route                 | Description                        |
|-----------------------|------------------------------------|
| `/api/rates/avg`      | Get average rate for the day       |
| `/api/rates/latest`   | Get latest rate                    |
| `/api/rates/day`      | Alias for `/api/rates/latest`      |
| `/api/rates/last_24h` | Alias for `/api/rates/day`         |

---

#### Example Usage

**Note**: 
1. REST api requires authentication and session, so `-c/-d cookies.txt` should be used
2. [`/last-24h`, `/day`] routes are just aliases, so request will be redirected, `-L` should be used 

```bash
# Authorize 
curl -X POST http://localhost:8082/auth \
     -d "api_key=68b6d4763d69baec1d2a4970" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -c cookies.txt

# Get latest rates
curl -X GET http://localhost:8082/api/rates/latest?pair[]=BTCEUR \
     -H "Authorization: Bearer " \
     -b cookies.txt

# Get avg rates
curl -X GET http://localhost:8082/api/rates/last-24h?pair[]=BTCEUR \
     -H "Authorization: Bearer " \
     -L \
     -b cookies.txt

# /last-24h alias for latest rates
curl -X GET "http://localhost:8082/api/rates/last-24h?pair[]=BTCEUR" \
     -H "Authorization: Bearer " \
     -L \
     -b cookies.txt

# /day alias for latest rates
curl -X GET "http://localhost:8082/api/rates/day?pair[]=BTCEUR&date=2025-09-01" \
     -H "Authorization: Bearer " \
     -L \
     -b cookies.txt
```

---

## 🧪 Testing & Development

- Predefined API key: `68b6d4763d69baec1d2a4970`
- Preloaded MySQL data for dates: `2025-09-01`, `2025-09-02`, `2025-09-03`
- Simple HTML+JavaScript homepage for testing API calls

---

## 📁 Configuration Files

Each service uses its own config file (`.config.yaml` or `.env`) to define environment variables, credentials, and service URLs. See individual service folders for details.

---

## 🐳 Running the Project

**Note**: auth, collector & rest_api services have built-in tests. They will be triggered during build and build will be canceled
if they fail

To spin up all services:

```bash
docker-compose up
```

Ensure:
1. config & env files are created from .example version:
* .env 
* service/rest_api/.env
* service/auth/config.yaml
* service/collector/config.yaml
2. Docker and Docker Compose are installed on your system.

After docker is running you can access services by links
* Auth Service: [localhost:AUTH_APP_PORT](http://localhost:8081/)
* REST Api Service: [localhost:REST_API_APP_PORT](http://localhost:8082/)


## 🚀 Future Improvements

Below are planned enhancements to improve scalability, security, and usability:

### 1. WebSocket Integration
- Implement a WebSocket API that authenticates via the Auth Service and streams data from Redis.
- Migrate Redis operations from the REST API to the WebSocket API.
- Remove Redis network dependency from the REST API service.

### 2. Web Application
Develop a lightweight frontend to visualize and interact with backend data:
- 📈 Real-time graphs powered by WebSocket streaming data.
- 📋 Daily rate lists fetched via REST API.
- 🛡️ Authentication logs display for monitoring and debugging.

### 3. Security Enhancements
- Transition all services to secure communication protocols (e.g., HTTPS, WSS).

### 4. Configuration Management
- Consolidate all config and environment variables into a single source file.
- Generate service-specific config files during build time from the unified config.

### 5. Integration Testing with Mocks
- Implement full mock-based integration tests using tools like [Testcontainers](https://www.testcontainers.org/) to simulate remote services.
- Validate service interactions (e.g., Auth, Redis, MySQL) in isolated environments with reproducible test setups.
