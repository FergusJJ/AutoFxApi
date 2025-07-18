# AutoFxApi

A Go-based API service that monitors cTrader positions and broadcasts trading signals to connected clients, designed for [AutoFxApp](https://github.com/FergusJJ/AutoFxApp).

## Overview

AutoFxApi serves as the backend infrastructure for automated trading signal distribution:
- Monitors cTrader accounts using sharing codes
- Decodes cTrader's proprietary WebSocket protocol
- Stores position data in Redis for real-time distribution
- Provides REST API and WebSocket endpoints for clients
- Manages user authentication with JWT tokens
- Originally integrated with WHOP for license management

## Architecture

### Core Services

**HTTP API Server (`cmd/http/main.go`)**
- Fiber framework for high-performance HTTP handling
- JWT authentication for protected endpoints
- WebSocket support for real-time position updates
- PostgreSQL for persistent storage
- Redis for position caching

**Monitor Service (`cmd/monitor/main.go`)**
- Connects to cTrader WebSocket API (h30.p.ctrader.com)
- Decodes binary protocol messages
- Tracks position changes (OPEN/CLOSE)
- Publishes updates to Redis

**Monitor Manager (`cmd/monitor-manager/main.go`)**
- Orchestrates multiple monitor instances
- Handles start/stop commands via Redis queue
- Manages process lifecycle and restarts

### Data Flow

1. **Position Monitoring**
   - Monitor connects to cTrader using sharing code
   - Receives position updates via WebSocket
   - Decodes binary messages using custom protocol implementation
   - Compares with cached positions to detect changes

2. **Signal Distribution**
   - Position changes pushed to Redis queue
   - API server polls Redis for updates
   - Broadcasts to connected WebSocket clients by pool
   - Clients receive OPEN/CLOSE signals with position details

### Protocol Implementation (`internal/monitor/`)

**Message Encoding/Decoding**
- Custom binary protocol parser
- Handles various data types (string, int, float, bytes)
- Recursive message structure decoding
- Type-safe payload marshaling

**Supported Message Types**
- 4314: Initial sharing code request
- 4315: Trader information response
- 4258: Position list request
- 4259: Position list response
- 50: Error responses

### API Endpoints

**Authentication**
- `GET /whop/validate` - Validate license key
- `GET /api/auth/new` - Generate JWT tokens
- `POST /api/auth/refresh` - Refresh access token

**Account Management**
- `POST /api/account/create` - Create account
- `POST /api/account/get` - Get account details
- `POST /api/account/update` - Update account
- `POST /api/account/delete` - Delete account

**Position Management**
- `GET /api/user/position/all` - Get all positions
- `POST /api/user/position/new` - Store new position
- `POST /api/user/position/delete` - Delete position

**WebSocket**
- `WS /ws/monitor` - Real-time position updates

**Internal**
- `POST /internal/monitor/configure-monitor` - Start/stop monitors

### Database Schema

**PostgreSQL Tables**
- `account` - User accounts with license keys
- `user_position` - Stored trading positions
- `user_data` - JWT refresh tokens

**Redis Keys**
- `positionUpdates` - Queue for position changes
- `monitorUpdates` - Queue for monitor commands
- `storage-positions-pool-{id}` - Cached positions per pool

### Key Features

**Real-time Monitoring**
- WebSocket connection to cTrader
- Binary protocol decoding
- Position change detection
- Automatic reconnection handling

**Authentication & Security**
- JWT-based authentication
- API key validation for internal endpoints
- License key verification (WHOP integration)
- Refresh token rotation

**Scalability**
- Redis for distributed caching
- Docker Compose deployment
- Separate monitor processes
- Message queue architecture

## Configuration

**Environment Variables**
- `PORT` - API server port
- `JWT_SIGNING_KEY` - JWT signing secret
- `JWT_REFRESH_SIGNING_KEY` - Refresh token secret
- `MONITOR_SECRET_KEY` - Internal API authentication
- `WHOP_VALIDATE_LICENSE_API_KEY` - WHOP API key
- `PG_USER`, `PG_PASSWORD`, `PG_DB` - PostgreSQL credentials

**Docker Services**
- `api` - Main HTTP/WebSocket server
- `monitor-manager` - Monitor orchestration
- `redis` - Position caching
- `postgres` - Persistent storage
- `redis_insight` - Redis debugging UI

## Technical Notes

- Complex binary protocol encoding/decoding implementation
- Project was abandoned before completion
- WHOP authentication needs removal for standalone use
- Monitor processes run as separate executables
- Position volume divided by 100 for proper display

## Message Format

Position updates are broadcast as:
```json
{
  "symbol": "EURUSD",
  "pool": "pool_id",
  "copyPID": "12345",
  "symbolID": 1,
  "price": 1.0856,
  "volume": 1000,
  "direction": "BUY",
  "type": "OPEN",
  "openedTimestamp": 1692295096822
}
