# DynamoDB CloudWatch Monitoring Setup Guide

## Step 1: Navigate to CloudWatch Console
1. Go to AWS CloudWatch console
2. Change region to Oregon (us-west-2)

## Step 2: Create DynamoDB Dashboard
1. Click "Dashboards" in left sidebar
2. Click "Create dashboard"
3. Name it "DynamoDB-Performance-Test"

## Step 3: Add Key Metrics

### Table Operation Metrics
Add the following metrics from "DynamoDB" namespace:

1. **Throughput Metrics**
```sql
SELECT SUM(ConsumedReadCapacityUnits), SUM(ConsumedWriteCapacityUnits)
FROM "AWS/DynamoDB"
WHERE TableName = 'shopping_carts'
GROUP BY TableName
```

2. **Latency Metrics**
```sql
SELECT AVG(SuccessfulRequestLatency)
FROM "AWS/DynamoDB"
WHERE TableName = 'shopping_carts'
AND Operation IN ('GetItem', 'PutItem', 'UpdateItem')
GROUP BY Operation
```

3. **Error Metrics**
```sql
SELECT SUM(ThrottledRequests), SUM(SystemErrors), SUM(UserErrors)
FROM "AWS/DynamoDB"
WHERE TableName = 'shopping_carts'
GROUP BY Operation
```

### GSI Metrics (if using)
```sql
SELECT SUM(ConsumedReadCapacityUnits), SUM(ConsumedWriteCapacityUnits)
FROM "AWS/DynamoDB"
WHERE TableName = 'shopping_carts'
AND GlobalSecondaryIndexName = '<your-gsi-name>'
GROUP BY GlobalSecondaryIndexName
```

## Step 4: Create Widgets Layout

1. **Throughput Widget**
- Title: "DynamoDB Throughput"
- Metrics: ConsumedReadCapacityUnits, ConsumedWriteCapacityUnits
- Graph type: Line
- Period: 1 minute

2. **Latency Widget**
- Title: "Operation Latency"
- Metrics: SuccessfulRequestLatency by Operation
- Graph type: Line
- Period: 1 minute

3. **Errors Widget**
- Title: "DynamoDB Errors"
- Metrics: ThrottledRequests, SystemErrors, UserErrors
- Graph type: Line
- Period: 1 minute

4. **Table Metrics Widget**
- Title: "Table Metrics"
- Include:
  - ItemCount
  - TableSizeBytes
  - ProvisionedReadCapacityUnits (if provisioned)
  - ProvisionedWriteCapacityUnits (if provisioned)

## Step 5: Set Up Alarms (Optional)

1. **Throttling Alarm**
```sql
SELECT SUM(ThrottledRequests)
FROM "AWS/DynamoDB"
WHERE TableName = 'shopping_carts'
```
- Threshold: > 0 for 1 datapoint within 1 minute

2. **High Latency Alarm**
```sql
SELECT AVG(SuccessfulRequestLatency)
FROM "AWS/DynamoDB"
WHERE TableName = 'shopping_carts'
```
- Threshold: > 100ms for 3 datapoints within 5 minutes

## Step 6: Testing Process

1. Open the dashboard before starting tests
2. Set time range to "Last hour" or custom 5-minute window
3. Run test command:
```bash
cd /Users/liahuang/web-service-gin/HW8/tests && \
API_URL=http://CS6650L2-alb-397575304.us-west-2.elb.amazonaws.com \
go run runner.go dynamodb
```
4. Monitor metrics during test execution
5. After test completion:
   - Take screenshots of all metrics
   - Export dashboard if needed
   - Note any throttling or latency spikes

## Key Metrics to Watch

1. **Performance Metrics**
- SuccessfulRequestLatency
- ConsumedReadCapacityUnits
- ConsumedWriteCapacityUnits

2. **Error Metrics**
- ThrottledRequests
- SystemErrors
- UserErrors

3. **Capacity Metrics**
- ProvisionedThroughput (if using provisioned capacity)
- ConsumedThroughput
- OnlineIndexConsumedThroughput (for GSIs)

4. **Storage Metrics**
- TableSizeBytes
- ItemCount

## Analysis Guidelines

1. **Latency Analysis**
- Check average latency per operation type
- Note any spikes during bulk operations
- Compare GET vs PUT latencies

2. **Throughput Analysis**
- Monitor consumed capacity vs available
- Check for throttling events
- Analyze read/write patterns

3. **Error Analysis**
- Track throttling events
- Monitor system errors
- Check for hot partitions

4. **Capacity Planning**
- Review peak throughput requirements
- Analyze capacity utilization
- Identify optimization opportunities