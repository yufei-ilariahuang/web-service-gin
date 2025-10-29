from locust import HttpUser, task, between
import random

class OrderUser(HttpUser):
    wait_time = between(0.1, 0.5)  # Wait 100-500ms between requests
    
    def on_start(self):
        self.customer_id = random.randint(1, 10000)
    
    @task
    def create_order(self):
        """Place an order"""
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