# MySQL RDS CloudWatch Monitoring Setup Guide

## Step 1: Navigate to CloudWatch Console
1. Go to AWS CloudWatch console
2. Change region to Oregon (us-west-2)

## Step 2: Create RDS Dashboard
1. Click "Dashboards" in left sidebar
2. Click "Create dashboard"
3. Name it "MySQL-Performance-Test"

## Step 3: Add Key Metrics

### RDS Instance Metrics
Add the following metrics from "AWS/RDS" namespace:

1. **CPU and Memory**
```sql
SELECT AVG(CPUUtilization), AVG(FreeableMemory)
FROM "AWS/RDS"
WHERE DBInstanceIdentifier = '<your-rds-instance-id>'
GROUP BY DBInstanceIdentifier
```

2. **Connection Metrics**
```sql
SELECT AVG(DatabaseConnections)
FROM "AWS/RDS"
WHERE DBInstanceIdentifier = '<your-rds-instance-id>'
```

3. **I/O Metrics**
```sql
SELECT SUM(ReadIOPS), SUM(WriteIOPS), 
       SUM(ReadLatency), SUM(WriteLatency)
FROM "AWS/RDS"
WHERE DBInstanceIdentifier = '<your-rds-instance-id>'
```

### ECS Service Metrics
```sql
SELECT AVG(CPUUtilization), AVG(MemoryUtilization)
FROM "AWS/ECS"
WHERE ClusterName = 'CS6650L2-cluster'
AND ServiceName = 'CS6650L2'
```

## Step 4: Create Widgets Layout

1. **CPU & Memory Widget**
- Title: "RDS Resource Utilization"
- Metrics: CPUUtilization, FreeableMemory
- Graph type: Line
- Period: 1 minute

2. **Connections Widget**
- Title: "Database Connections"
- Metrics: DatabaseConnections
- Graph type: Line
- Period: 1 minute

3. **I/O Performance Widget**
- Title: "Database I/O"
- Metrics: ReadIOPS, WriteIOPS, ReadLatency, WriteLatency
- Graph type: Line
- Period: 1 minute

4. **ECS Performance Widget**
- Title: "ECS Service Performance"
- Metrics: CPUUtilization, MemoryUtilization
- Graph type: Line
- Period: 1 minute

## Step 5: Set Up Alarms (Optional)

1. **High CPU Alarm**
```sql
SELECT AVG(CPUUtilization)
FROM "AWS/RDS"
WHERE DBInstanceIdentifier = '<your-rds-instance-id>'
```
- Threshold: > 80% for 3 datapoints within 5 minutes

2. **Connection Count Alarm**
```sql
SELECT AVG(DatabaseConnections)
FROM "AWS/RDS"
WHERE DBInstanceIdentifier = '<your-rds-instance-id>'
```
- Threshold: > 80% of max connections for 3 datapoints within 5 minutes

## Step 6: Testing Process

1. Open the dashboard before starting tests
2. Set time range to "Last hour" or custom 5-minute window
3. Run test command:
```bash
cd /Users/liahuang/web-service-gin/HW8/tests && \
API_URL=http://CS6650L2-alb-397575304.us-west-2.elb.amazonaws.com \
go run runner.go mysql
```
4. Monitor metrics during test execution
5. After test completion:
   - Take screenshots of all metrics
   - Export dashboard if needed
   - Note any performance bottlenecks

## Key Metrics to Watch

1. **Performance Metrics**
- CPUUtilization
- FreeableMemory
- ReadLatency
- WriteLatency

2. **Connection Metrics**
- DatabaseConnections
- Connection timeouts
- Max connections

3. **I/O Metrics**
- ReadIOPS
- WriteIOPS
- DiskQueueDepth
- FreeStorageSpace

4. **ECS Metrics**
- Service CPU utilization
- Service memory utilization
- Task count

## Analysis Guidelines

1. **Performance Analysis**
- Monitor CPU usage patterns
- Check memory utilization
- Analyze I/O latency trends
- Review connection pool efficiency

2. **Resource Utilization**
- Track memory usage
- Monitor storage capacity
- Analyze connection patterns
- Check ECS resource usage

3. **Query Performance**
- Monitor slow query logs
- Check I/O patterns
- Analyze connection pooling
- Review transaction throughput

4. **Capacity Planning**
- Review peak CPU requirements
- Analyze memory needs
- Check storage growth
- Monitor connection limits