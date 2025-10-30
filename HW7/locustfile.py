from locust import HttpUser, task, between, tag
import random

# locust -f locustfile.py --host=http://CS6650L2-alb-767794308.us-west-2.elb.amazonaws.com

class SyncOrderUser(HttpUser):
    """Test synchronous order endpoint - will be slow and timeout under load"""
    wait_time = between(0.1, 0.5)
    
    def on_start(self):
        self.customer_id = random.randint(1, 10000)
    
    @task
    @tag('sync')
    def create_order(self):
        """Place an order via sync endpoint"""
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": str(random.randint(1, 100)),
                    "quantity": random.randint(1, 3)
                }
            ]
        }
        
        self.client.post("/orders/sync", json=order_data)


class AsyncOrderUser(HttpUser):
    """Test asynchronous order endpoint - should handle 100% of load"""
    wait_time = between(0.1, 0.5)
    
    def on_start(self):
        self.customer_id = random.randint(1, 10000)
    
    @task
    @tag('async')
    def create_async_order(self):
        """Place an order via async endpoint - should return 202 immediately"""
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": str(random.randint(1, 100)),
                    "quantity": random.randint(1, 3)
                }
            ]
        }
        
        with self.client.post("/orders/async", json=order_data, catch_response=True) as response:
            if response.status_code == 202:
                response.success()
            else:
                response.failure(f"Expected 202, got {response.status_code}")
    
    @task(1)
    @tag('async')
    def check_stats(self):
        """Check system statistics"""
        self.client.get("/stats")