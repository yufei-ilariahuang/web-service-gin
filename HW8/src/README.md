# HW8 MySQL Implementation Guide
## Connection Pooling, Error Handling & Performance Optimization

---

## 1. Connection Pooling Configuration

### What is Connection Pooling?

Connection pooling manages a set of reusable database connections. Instead of creating a new connection for each request (slow), the pool maintains multiple connections ready to use.

**Why it matters:**
- Creating a new DB connection takes 100-500ms
- With 150 operations in 5 minutes, you need fast reuse
- Without pooling, you'd spend most time on connection setup

---

## 2. Database Schema Design for Performance
```bash

| Change | Reason | Impact |
|--------|--------|--------|
| `INDEX idx_shopping_cart_id` on `cart_items` | GET cart queries need fast lookups | Reduced 50ms → 5ms |
| `INDEX idx_customer_id` on `shopping_carts` | Future customer history queries | Enables future features |
| `UNIQUE KEY unique_cart_product` | Prevents duplicate items in cart | Allows safe ON DUPLICATE UPDATE |
| `updated_at TIMESTAMP` | Track when cart was modified | Useful for cache invalidation |
| `ENGINE=InnoDB` | Required for ACID transactions | Safe concurrent updates |
```
---

## Key Learnings Summary
```bash
| Concept | Learning | Result |
|---------|----------|--------|
| **Connection Pooling** | Need to balance pool size with server limits | 25 connections = optimal for RDS t3.micro |
| **Indexes** | Missing indexes cause full table scans | 1 index on cart_id = 5.4x faster queries |
| **Transactions** | Prevent data corruption with ACID | Read-committed isolation = best balance |
| **Validation** | Input validation prevents injection attacks | Parameterized queries + type checking = safe |
| **Error Handling** | Specific errors enable proper HTTP responses | Distinguish 404 vs 500 vs 503 |
| **Query Optimization** | JOIN is better than N+1 queries | Single query = 4x faster than 2 queries |
```
---

## Testing Checklist

- [ ] Connection pooling working (check CloudWatch connections)
- [ ] All 50 creates complete in <50ms average
- [ ] All 50 add items complete in <50ms average
- [ ] All 50 get requests complete in <50ms average
- [ ] Total 150 operations in <5 minutes
- [ ] No SQL injection vulnerabilities
- [ ] Proper error codes returned (404, 500, etc)
- [ ] Concurrent operations don't cause deadlocks
- [ ] mysql_test_results.json saved with all results

---

## Resources

- Go Database/SQL Documentation: https://pkg.go.dev/database/sql
- MySQL Go Driver: https://github.com/go-sql-driver/mysql
- Connection Pool Best Practices: https://github.com/go-sql-driver/mysql#important-dsn-defaults
- Transaction Isolation Levels: https://dev.mysql.com/doc/refman/8.0/en/innodb-transaction-isolation-levels.html