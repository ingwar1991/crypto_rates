// ===== DB =====
db = db.getSiblingDB('authDB');

// ===== User for accessing DB =====
db.createUser({
  user: "usr",
  pwd: "pwd",
  roles: [
    { role: "readWrite", db: "authDB" }
  ]
});

// ===== Users =====
db.createCollection("users");
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ api_key: 1 }, { unique: true });

// Adding a sample (admin) api_key 
db.users.insertOne({
    email: "admin@crypto_rates.com",
    api_key: "68b6d4763d69baec1d2a4970",
    created_at: ISODate('2025-09-02T11:26:46.475Z')
});

// ===== OTP Codes =====
db.createCollection("login_codes");
db.login_codes.createIndex({ email: 1 });
db.login_codes.createIndex({ code_hash: 1 });
db.login_codes.createIndex(
  { expires_at: 1 },
  { expireAfterSeconds: 0 } 
);

// ===== Active Connected Secrets =====
db.createCollection("active_secrets");
db.active_secrets.createIndex({ email: 1 });
db.active_secrets.createIndex({ secret: 1 });
db.active_secrets.createIndex(
  { expires_at: 1 },
  { expireAfterSeconds: 0 } // optional TTL for session expiry
);

// ===== Logs for Stream API =====
db.createCollection("logs_stream");
db.logs_stream.createIndex({ email: 1 });
db.logs_stream.createIndex({ secret: 1 });
db.logs_stream.createIndex({ endpoint: 1 });
db.logs_stream.createIndex({ response_status: 1 });
db.logs_stream.createIndex({ timestamp: -1 });

// ===== Logs for REST API =====
db.createCollection("logs_rest");
db.logs_rest.createIndex({ email: 1 });
db.logs_rest.createIndex({ secret: 1 });
db.logs_rest.createIndex({ endpoint: 1 });
db.logs_rest.createIndex({ response_status: 1 });
db.logs_rest.createIndex({ timestamp: -1 });
