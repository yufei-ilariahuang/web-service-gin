"""
Locust Load Testing: All 4 Scenarios
Matches YOUR exact API calls from runner.go:
  1. POST /shopping-carts (create_cart)
  2. POST /shopping-carts/{id}/items (add_items)  
  3. GET /shopping-carts/{id} (get_cart)

IMPORTANT: Switch backend via environment variable or redeploy:
  - DB_BACKEND=mysql (default)
  - DB_BACKEND=dynamodb
"""

from locust import HttpUser, task, between, events
from locust.contrib.fasthttp import FastHttpUser
import json
import time
import statistics
from datetime import datetime

# Metrics collection (same as runner.go structure)
response_times = {
    'create_cart': [],
    'add_items': [],
    'get_cart': [],
}

total_operations = {'count': 0}


@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """Print test configuration"""
    print("\n" + "="*80)
    print("SHOPPING CART API LOAD TEST")
    print("="*80)
    print(f"Start Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"Target: {environment.host}")
    print("Testing the EXACT same API calls as runner.go")
    print("="*80 + "\n")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Print comprehensive test results (same metrics as runner.go)"""
    stats = environment.stats
    
    print("\n" + "="*80)
    print("TEST RESULTS SUMMARY")
    print("="*80)
    
    # Overall stats
    total_requests = stats.total.num_requests
    total_failures = stats.total.num_failures
    success_rate = ((total_requests - total_failures) / max(total_requests, 1)) * 100
    
    print(f"\nOverall Performance:")
    print(f"  Total Requests: {total_requests}")
    print(f"  Total Failures: {total_failures}")
    print(f"  Success Rate: {success_rate:.2f}%")
    
    print(f"\nResponse Time Statistics (milliseconds):")
    print(f"  Average: {stats.total.avg_response_time:.2f}ms")
    print(f"  Min: {stats.total.min_response_time:.2f}ms")
    print(f"  Max: {stats.total.max_response_time:.2f}ms")
    print(f"  P50 (Median): {stats.total.get_response_time_percentile(0.5):.2f}ms")
    print(f"  P95: {stats.total.get_response_time_percentile(0.95):.2f}ms")
    print(f"  P99: {stats.total.get_response_time_percentile(0.99):.2f}ms")
    
    # Per-operation breakdown (same format as runner.go output)
    print(f"\nPer-Operation Breakdown:")
    for operation in sorted(response_times.keys()):
        times = response_times[operation]
        if times:
            print(f"\n  {operation.upper()}:")
            print(f"    Count: {len(times)}")
            print(f"    Avg: {statistics.mean(times):.2f}ms")
            if len(times) > 1:
                print(f"    Median: {statistics.median(times):.2f}ms")
                sorted_times = sorted(times)
                p95_idx = min(int(len(times)*0.95), len(times)-1)
                p99_idx = min(int(len(times)*0.99), len(times)-1)
                print(f"    P95: {sorted_times[p95_idx]:.2f}ms")
                print(f"    P99: {sorted_times[p99_idx]:.2f}ms")
    
    print("\n" + "="*80 + "\n")


class CartUser(FastHttpUser):
    """
    Base user class - uses EXACT API calls from runner.go
    
    API Endpoints (matching runner.go):
    1. POST /shopping-carts           (create_cart)
    2. POST /shopping-carts/{id}/items (add_items)
    3. GET /shopping-carts/{id}        (get_cart)
    """
    
    wait_time = between(0.5, 2)
    cart_id = None
    user_id = None
    
    def on_start(self):
        """Initialize user with first cart (matches runner.go on_start)"""
        self.user_id = hash(id(self)) % 100000
        self._create_cart()
    
    def _create_cart(self):
        """
        Exact match to runner.go create_cart call:
        POST /shopping-carts
        Body: {"customer_id": int32}
        Response: {"shopping_cart_id": int32}
        """
        start = time.time()
        try:
            response = self.client.post(
                "/shopping-carts",
                json={"customer_id": self.user_id},
                timeout=5
            )
            elapsed = (time.time() - start) * 1000
            response_times['create_cart'].append(elapsed)
            total_operations['count'] += 1
            
            # Parse response matching runner.go structure
            if response.status_code in [200, 201]:
                data = response.json()
                self.cart_id = data.get('shopping_cart_id')
                print(f"[CREATE] Cart ID: {self.cart_id}, Time: {elapsed:.2f}ms")
        except Exception as e:
            elapsed = (time.time() - start) * 1000
            response_times['create_cart'].append(elapsed)
            total_operations['count'] += 1
            print(f"[CREATE ERROR] {str(e)}, Time: {elapsed:.2f}ms")
    
    def _add_item(self):
        """
        Exact match to runner.go add_items call:
        POST /shopping-carts/{cartID}/items
        Body: {"product_id": int32, "quantity": int32}
        Response: 204 or 200 on success
        """
        if not self.cart_id:
            self._create_cart()
            if not self.cart_id:
                return
        
        product_id = (hash(self.user_id) % 100) + 1
        quantity = (hash(self.user_id + product_id) % 5) + 1
        
        start = time.time()
        try:
            response = self.client.post(
                f"/shopping-carts/{self.cart_id}/items",
                json={
                    "product_id": product_id,
                    "quantity": quantity
                },
                timeout=5
            )
            elapsed = (time.time() - start) * 1000
            response_times['add_items'].append(elapsed)
            total_operations['count'] += 1
            
            if response.status_code in [200, 204]:
                print(f"[ADD_ITEMS] Cart: {self.cart_id}, Product: {product_id}, Time: {elapsed:.2f}ms")
            else:
                print(f"[ADD_ITEMS ERROR] Status: {response.status_code}, Time: {elapsed:.2f}ms")
        except Exception as e:
            elapsed = (time.time() - start) * 1000
            response_times['add_items'].append(elapsed)
            total_operations['count'] += 1
            print(f"[ADD_ITEMS ERROR] {str(e)}, Time: {elapsed:.2f}ms")
    
    def _get_cart(self):
        """
        Exact match to runner.go get_cart call:
        GET /shopping-carts/{cartID}
        Response: {"id": int32, "customer_id": int32, "items": [...]}
        """
        if not self.cart_id:
            return
        
        start = time.time()
        try:
            response = self.client.get(
                f"/shopping-carts/{self.cart_id}",
                timeout=5
            )
            elapsed = (time.time() - start) * 1000
            response_times['get_cart'].append(elapsed)
            total_operations['count'] += 1
            
            if response.status_code == 200:
                data = response.json()
                items_count = len(data.get('items', []))
                print(f"[GET_CART] Cart: {self.cart_id}, Items: {items_count}, Time: {elapsed:.2f}ms")
            else:
                print(f"[GET_CART ERROR] Status: {response.status_code}, Time: {elapsed:.2f}ms")
        except Exception as e:
            elapsed = (time.time() - start) * 1000
            response_times['get_cart'].append(elapsed)
            total_operations['count'] += 1
            print(f"[GET_CART ERROR] {str(e)}, Time: {elapsed:.2f}ms")
    
    @task
    def user_workflow(self):
        """
        Realistic workflow matching runner.go test sequence:
        1. Create cart
        2. Add items
        3. Get cart to verify
        """
        self._create_cart()
        time.sleep(0.1)  # Small delay
        self._add_item()
        time.sleep(0.1)  # Small delay
        self._get_cart()


###############################################################################
# SCENARIO A: Startup MVP (100 users/day)
###############################################################################

class ScenarioAUser(CartUser):
    """
    Scenario A: Startup MVP
    - 100 users/day
    - 1 developer
    - Limited budget
    - Quick launch
    
    Peak Load: ~10 concurrent users (100 users/day ÷ 15min session = ~0.11 concurrent)
    Simulated: 10 concurrent to test system
    
    COMMAND:
    locust -f locust_4scenarios.py --host=http://localhost:8080 \
            -u 10 -r 2 -t 5m --headless ScenarioAUser
    
    Expected Results:
    - MySQL: Avg ~55ms, P99 ~120ms ✓
    - DynamoDB: Avg ~63ms, P99 ~150ms ✓
    - Both handle easily at this scale
    """
    wait_time = between(1, 3)  # Casual browsing


###############################################################################
# SCENARIO B: Growing Business (10K users/day)
###############################################################################

class ScenarioBUser(CartUser):
    """
    Scenario B: Growing Business
    - 10K users/day
    - 5 developers
    - Moderate budget
    - Feature expansion
    
    Peak Load: ~50 concurrent users (10K users/day ÷ 15min session ≈ 1.1k/hour = 18 concurrent)
    Simulated: 50 concurrent for sustained load test
    
    COMMAND:
    locust -f locust_4scenarios.py --host=http://localhost:8080 \
            -u 50 -r 5 -t 15m --headless ScenarioBUser
    
    CRITICAL TEST: P99 latency determines which DB is better
    
    Expected Results (FROM YOUR TEST DATA):
    - MySQL: Avg ~60ms, P95 ~93ms, P99 ~135ms ✓
    - DynamoDB: Avg ~75ms, P95 ~144ms, P99 ~322ms ⚠️ (2.38x worse!)
    
    At 10K users/day with 10K daily requests:
      MySQL P99: 100 requests × 135ms = 13.5 seconds total
      DynamoDB P99: 100 requests × 322ms = 32.2 seconds total
    """
    wait_time = between(0.5, 2)  # More active shopping


###############################################################################
# SCENARIO C: Normal Load (50K users/day baseline)
###############################################################################

class ScenarioCNormalUser(CartUser):
    """
    Scenario C Normal: Normal Load Baseline
    - 50K users/day
    - Peak: ~100 concurrent users
    - Revenue critical but not spiking
    
    COMMAND:
    locust -f locust_4scenarios.py --host=http://localhost:8080 \
            -u 100 -r 10 -t 10m --headless ScenarioCNormalUser
    
    Expected Results:
    - MySQL: Avg ~60ms, P99 ~150ms ✓
    - DynamoDB: Avg ~80ms, P99 ~300ms (starting to show strain)
    """
    wait_time = between(0.3, 1.5)


###############################################################################
# SCENARIO C: Flash Sale Spike (1M spike, 20x increase)
###############################################################################

class ScenarioCSpike(CartUser):
    """
    Scenario C Spike: Flash Sale / Peak Traffic Event
    - Normal: 50K users/day
    - Spike: 1M concurrent users (20x increase!)
    - Revenue critical - cannot fail
    
    Peak Load: ~500 concurrent users (simulating spike)
    
    COMMAND (STRESS TEST - WARNING):
    locust -f locust_4scenarios.py --host=http://localhost:8080 \
            -u 500 -r 50 -t 10m --headless ScenarioCSpike
    
    ⚠️  WARNING: This heavily stresses infrastructure
    
    Expected Results:
    - MySQL: Avg ~150-200ms, P99 ~400-600ms (struggling)
    - DynamoDB: Avg ~120-150ms, P99 ~300-400ms (handles better)
    
    Recommendation: Use hybrid (MySQL + cache layer)
    """
    wait_time = between(0.1, 0.5)  # Rapid shopping during flash sale


###############################################################################
# SCENARIO D: Global Platform (Millions of users)
###############################################################################

class ScenarioDUser(CartUser):
    """
    Scenario D: Global Platform
    - Millions of users
    - Multi-region requirement
    - 24/7 availability
    - Enterprise requirements
    
    Peak Load: ~1000 concurrent users (simulating global scale)
    Duration: 30 minutes (test for resource exhaustion)
    
    COMMAND (VERY LONG TEST):
    locust -f locust_4scenarios.py --host=localhost:8080 \
            -u 1000 -r 100 -t 30m --headless ScenarioDUser
    
    ⚠️  WARNING: 30-minute test - watch for memory leaks
    
    Expected Results:
    - MySQL: Response time degrades over time, P99 → 600ms+, Success rate drops
    - DynamoDB: Maintains stable performance, P99 stays 300-500ms
    
    Conclusion: Only DynamoDB scales to millions
    """
    wait_time = between(0.2, 1)  # Sustained high activity


"""
================================================================================
HOW TO SWITCH BETWEEN MYSQL AND DYNAMODB
================================================================================

The SAME Locust test works for both databases!
You just need to switch the backend in your API.

METHOD 1: Environment Variables (Recommended)
==========================================

Run API with MySQL backend:
  DB_BACKEND=mysql \
  DB_ENDPOINT=terraform-xxx.cpkkycme8ops.us-west-2.rds.amazonaws.com \
  DB_USER=admin \
  DB_PASSWORD=your-password \
  DB_NAME=ecommerce \
  ./api_binary

Then run Locust test:
  locust -f locust_4scenarios.py --host=http://localhost:8080 \
          -u 50 -r 5 -t 15m --headless ScenarioBUser

Results: MySQL performance data

---

Switch to DynamoDB backend:
  DB_BACKEND=dynamodb \
  AWS_REGION=us-west-2 \
  DYNAMO_TABLE=shopping_carts \
  ./api_binary

Then run SAME Locust test:
  locust -f locust_4scenarios.py --host=http://localhost:8080 \
          -u 50 -r 5 -t 15m --headless ScenarioBUser

Results: DynamoDB performance data (for direct comparison!)

---

METHOD 2: Docker/Container (if using containers)
================================================

MySQL backend container:
  docker run -e DB_BACKEND=mysql \
             -e DB_ENDPOINT=rds-endpoint \
             -p 8080:8080 \
             your-api:latest

DynamoDB backend container:
  docker run -e DB_BACKEND=dynamodb \
             -e AWS_REGION=us-west-2 \
             -p 8080:8080 \
             your-api:latest

Then run same Locust tests against each.

---

METHOD 3: Check which backend is running
=========================================

The API logs on startup show which backend is active:

MySQL starts and says:
  "Initializing MySQL database at: terraform-xxx.cpkkycme8ops.us-west-2.rds.amazonaws.com"

DynamoDB starts and says:
  "Initializing DynamoDB in region: us-west-2"

================================================================================
QUICK COMPARISON TEST SEQUENCE
================================================================================

1. Start API with MySQL backend:

   DB_BACKEND=mysql \
   DB_ENDPOINT=your-rds-endpoint \
   DB_USER=admin \
   DB_PASSWORD=your-password \
   DB_NAME=ecommerce \
   ./api_binary &

   Verify it's running: curl http://localhost:8080/health

2. Run Scenario B test on MySQL:

   locust -f locust_4scenarios.py --host=http://CS6650L2-alb-397575304.us-west-2.elb.amazonaws.com:8080 \
           -u 50 -r 5 -t 15m --headless ScenarioBUser

   WRITE DOWN THE RESULTS:
   - MySQL Avg Response: _____ ms
   - MySQL P99 Response: _____ ms
   - MySQL Success Rate: _____ %

3. Stop API:
   kill %1 (or Ctrl+C)

4. Start API with DynamoDB backend:

   DB_BACKEND=dynamodb \
   AWS_REGION=us-west-2 \
   DYNAMO_TABLE=shopping_carts \
   ./api_binary &

   Verify it's running: curl http://localhost:8080/health

5. Run SAME Scenario B test on DynamoDB:

   locust -f locust_4scenarios.py --host=http://CS6650L2-alb-397575304.us-west-2.elb.amazonaws.com:8080 \
           -u 50 -r 5 -t 15m --headless ScenarioBUser

   WRITE DOWN THE RESULTS:
   - DynamoDB Avg Response: _____ ms
   - DynamoDB P99 Response: _____ ms
   - DynamoDB Success Rate: _____ %

6. COMPARE:

   Your test data (from runner.go) predicted:
   - MySQL Avg: 55.90ms, P99: 134.95ms
   - DynamoDB Avg: 63.98ms, P99: 322.13ms

   Your Locust test should show similar results!

================================================================================
KEY METRICS TO COMPARE (MySQL vs DynamoDB)
================================================================================

Run this command to extract key metrics from Locust output:

For each operation, record:
1. COUNT: How many times operation ran
2. AVG: Average response time
3. MEDIAN (P50): 50th percentile
4. P95: 95th percentile (when users start noticing lag)
5. P99: 99th percentile (SLA critical - tail latency)

SCENARIO B Results should show:

MySQL:
  create_cart - Avg: ~42ms, P99: ~100ms
  add_items - Avg: ~41ms, P99: ~100ms
  get_cart - Avg: ~85ms, P99: ~150ms

DynamoDB:
  create_cart - Avg: ~40ms, P99: ~100ms
  add_items - Avg: ~39ms, P99: ~100ms
  get_cart - Avg: ~113ms, P99: ~300ms ⚠️ (much worse!)

The get_cart difference is critical: 27.71ms slower on DynamoDB!

================================================================================
"""