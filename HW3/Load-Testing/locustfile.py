# 1. python
# import random
# from locust import HttpUser, task, between

# class MyUser(HttpUser):
#     # Users will wait between 1 and 2 seconds between tasks
#     wait_time = between(1, 2)
#     host = "http://server:5000"  # 'server' is the hostname inside the Docker network

#     def on_start(self):
#         """ on_start is called when a Locust start before any task is scheduled """
#         # We pre-populate some data to ensure GET requests have something to fetch
#         for i in range(10):
#             self.client.post("/data", json={"key": f"key-{i}", "value": f"value-{i}"})

#     @task(3) # Give GET requests twice the weight/probability
#     def get_data_task(self):
#         """ Defines the GET request task """
#         # Choose a random key that we know exists
#         random_key = random.randint(0, 9)
#         self.client.get(f"/data/key-{random_key}", name="/data/[key]") # Group stats under one name

#     @task(1) # Give POST requests a lower weight
#     def post_data_task(self):
#         """ Defines the POST request task """
#         # Generate a new random key to avoid conflicts
#         random_key_id = random.randint(10, 10000)
#         self.client.post("/data", json={"key": f"new-key-{random_key_id}", "value": "some-value"})

# 2.C
import random
from locust import FastHttpUser, task, between, constant

class MyUser(FastHttpUser): # Step 2: Change the base class
    # To really see the difference, let's use a very small wait time.
    # This will make the HTTP client itself the bottleneck.
    wait_time = constant(0.01) # 10ms wait
    
    # If you want to replicate the throttled test, use this line instead:
    # wait_time = between(1, 2)

    host = "http://server:5000"

    def on_start(self):
        """ on_start is called when a Locust user starts before any task is scheduled """
        for i in range(10):
            self.client.post("/data", json={"key": f"key-{i}", "value": f"value-{i}"})

    @task(3)
    def get_data_task(self):
        random_key = random.randint(0, 9)
        self.client.get(f"/data/key-{random_key}", name="/data/[key]")

    @task(1)
    def post_data_task(self):
        random_key_id = random.randint(10, 100000)
        self.client.post("/data", json={"key": f"new-key-{random_key_id}", "value": "some-value"})

