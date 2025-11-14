# Backend API Specification

This document describes the backend API that the Cloudflared Desktop Tunnel application connects to for token management and remote commands.

## Base URL

Configured in the application settings. Default: `https://api.example.com`

## Authentication

Currently, no authentication is implemented in the demo. In production, you should implement:
- API key authentication
- JWT tokens
- OAuth 2.0
- mTLS (mutual TLS)

## REST API Endpoints

### 1. Get Tunnel Token

**Endpoint**: `GET /api/token`

**Description**: Fetches a Cloudflare tunnel token for the client to use.

**Request**:
```http
GET /api/token HTTP/1.1
Host: api.example.com
Content-Type: application/json
```

**Response**:
```json
{
  "token": "eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ==",
  "expiresAt": "2025-11-15T12:00:00Z"
}
```

**Status Codes**:
- `200 OK`: Token fetched successfully
- `401 Unauthorized`: Authentication failed
- `500 Internal Server Error`: Server error

---

### 2. Report Tunnel Status

**Endpoint**: `POST /api/status`

**Description**: Reports the current tunnel status to the backend.

**Request**:
```http
POST /api/status HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "running": true,
  "tunnelName": "my-tunnel",
  "connections": 4,
  "uptime": 3600,
  "version": "1.0.0",
  "timestamp": "2025-11-14T10:00:00Z"
}
```

**Response**:
```json
{
  "status": "ok",
  "message": "Status received"
}
```

**Status Codes**:
- `200 OK`: Status received successfully
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Authentication failed
- `500 Internal Server Error`: Server error

---

## WebSocket API

### Commands WebSocket

**Endpoint**: `WS /api/commands`

**Description**: WebSocket connection for receiving real-time commands from the backend.

**Connection**:
```
ws://api.example.com/api/commands
```

or for secure connections:
```
wss://api.example.com/api/commands
```

### Command Format

All commands are sent as JSON messages:

```json
{
  "type": "command_type",
  "payload": {
    // Command-specific data
  }
}
```

### Supported Commands

#### 1. Update Command

**Description**: Instructs the client to update the application.

**Message**:
```json
{
  "type": "update",
  "payload": {
    "version": "1.1.0",
    "url": "https://releases.example.com/v1.1.0/app.exe",
    "checksum": "sha256:abc123...",
    "force": false
  }
}
```

**Payload Fields**:
- `version`: New version number
- `url`: Download URL for the update
- `checksum`: File checksum for verification
- `force`: Whether to force update immediately

---

#### 2. Restart Command

**Description**: Instructs the client to restart the tunnel.

**Message**:
```json
{
  "type": "restart",
  "payload": {
    "reason": "Configuration updated",
    "delay": 5
  }
}
```

**Payload Fields**:
- `reason`: Reason for restart (optional)
- `delay`: Delay in seconds before restart (optional, default: 0)

---

#### 3. Patch Command

**Description**: Applies a configuration patch without full restart.

**Message**:
```json
{
  "type": "patch",
  "payload": {
    "config": {
      "refreshInterval": 600,
      "backendURL": "https://new-api.example.com"
    }
  }
}
```

**Payload Fields**:
- `config`: Object containing configuration fields to update

---

#### 4. Stop Command

**Description**: Instructs the client to stop the tunnel.

**Message**:
```json
{
  "type": "stop",
  "payload": {
    "reason": "Maintenance window"
  }
}
```

**Payload Fields**:
- `reason`: Reason for stopping (optional)

---

#### 5. Fetch Logs Command

**Description**: Requests the client to send current logs.

**Message**:
```json
{
  "type": "fetch_logs",
  "payload": {
    "lines": 100
  }
}
```

**Payload Fields**:
- `lines`: Number of log lines to send

**Client Response** (sent back via WebSocket):
```json
{
  "type": "logs_response",
  "payload": {
    "logs": [
      "[2025-11-14 10:00:00] Tunnel started",
      "[2025-11-14 10:00:01] Connected to edge"
    ]
  }
}
```

---

## Example Backend Implementation

### Node.js + Express Example

```javascript
const express = require('express');
const expressWs = require('express-ws');

const app = express();
expressWs(app);

app.use(express.json());

// Store active WebSocket connections
const clients = new Set();

// GET /api/token
app.get('/api/token', (req, res) => {
  // Generate or fetch tunnel token from Cloudflare API
  const token = generateTunnelToken();
  
  res.json({
    token: token,
    expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000)
  });
});

// POST /api/status
app.post('/api/status', (req, res) => {
  const status = req.body;
  console.log('Received status:', status);
  
  // Store status in database
  // Send notifications if needed
  
  res.json({ status: 'ok', message: 'Status received' });
});

// WebSocket /api/commands
app.ws('/api/commands', (ws, req) => {
  console.log('Client connected');
  clients.add(ws);
  
  ws.on('message', (msg) => {
    const message = JSON.parse(msg);
    console.log('Received:', message);
    
    // Handle client responses (e.g., logs_response)
  });
  
  ws.on('close', () => {
    console.log('Client disconnected');
    clients.delete(ws);
  });
});

// Function to send command to all clients
function broadcastCommand(command) {
  clients.forEach(client => {
    client.send(JSON.stringify(command));
  });
}

// Example: Send restart command to all clients
setTimeout(() => {
  broadcastCommand({
    type: 'restart',
    payload: { reason: 'Scheduled maintenance', delay: 60 }
  });
}, 10000);

app.listen(3000, () => {
  console.log('Backend API running on port 3000');
});
```

## Security Recommendations

### Production Deployment

1. **Use HTTPS/WSS**: Always use encrypted connections
2. **Implement Authentication**: 
   - API keys in headers
   - JWT tokens with expiration
   - OAuth 2.0 for user authentication
3. **Rate Limiting**: Prevent abuse with rate limits
4. **Input Validation**: Validate all incoming data
5. **Token Rotation**: Rotate tunnel tokens regularly
6. **Audit Logging**: Log all API calls and commands
7. **CORS**: Configure CORS properly for web access

### Example with JWT Authentication

```javascript
const jwt = require('jsonwebtoken');

// Middleware for JWT authentication
function authenticateToken(req, res, next) {
  const token = req.header('Authorization')?.split(' ')[1];
  
  if (!token) {
    return res.status(401).json({ error: 'No token provided' });
  }
  
  jwt.verify(token, process.env.JWT_SECRET, (err, user) => {
    if (err) {
      return res.status(403).json({ error: 'Invalid token' });
    }
    req.user = user;
    next();
  });
}

// Apply to routes
app.get('/api/token', authenticateToken, (req, res) => {
  // ... endpoint logic
});
```

## Error Handling

### Standard Error Response

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      // Additional error details
    }
  }
}
```

### Common Error Codes

- `INVALID_TOKEN`: Provided token is invalid or expired
- `TUNNEL_NOT_FOUND`: Specified tunnel doesn't exist
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server encountered an error
- `UNAUTHORIZED`: Authentication failed
- `FORBIDDEN`: Insufficient permissions

## Testing

Use tools like:
- **curl** for REST API testing
- **wscat** for WebSocket testing
- **Postman** for API development
- **Artillery** for load testing

### Example curl Commands

```bash
# Get token
curl -X GET https://api.example.com/api/token

# Report status
curl -X POST https://api.example.com/api/status \
  -H "Content-Type: application/json" \
  -d '{"running":true,"tunnelName":"my-tunnel"}'
```

### Example wscat Command

```bash
# Connect to WebSocket
wscat -c ws://api.example.com/api/commands

# Server will send commands like:
# {"type":"restart","payload":{"reason":"Test"}}
```
