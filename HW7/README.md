
# Phase 1: Build Your Current System

Implement synchronous order processing:

```bash
POST /orders/sync → Verify Payment (3s delay) → Return 200 OK
```
#### Test Configuration (use Locust):

Spawn rate: 1 user/second (normal), 10 users/second (flash)
User wait time: random 100-500ms between requests
-  Test endpoint: POST /orders/sync
1. Test Normal Operations: 5 concurrent users, 30 seconds
Expected: 100% success rate

2. Test Flash Sale: 20 concurrent users, 60 seconds

```bash
Sync:   Customer → API → Payment (3s) → Response
Async:  Customer → API → Queue → Response (<100ms)
                           ↓
                   Background Workers → Payment (3s)
```