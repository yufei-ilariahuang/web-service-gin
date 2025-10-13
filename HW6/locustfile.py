import time
import random
from locust import HttpUser, task, between, FastHttpUser

class ProductSearchUser(FastHttpUser):
    """
    User class that does basic product searches with minimal wait time.
    Uses FastHttpUser for potentially higher performance.
    """
    # Define common search terms based on your product data
    common_brands = ["Apple", "Samsung", "Google", "Microsoft", "Sony", "Logitech"]
    common_categories = ["Electronics", "Books", "Home", "Clothing", "Groceries", "Toys"]


    

    @task(2100)  # Brand-only searches
    def search_by_brand(self):
        """
        Search only by brand
        """
        brand_to_search = random.choice(self.common_brands)
        self.client.get(f"/products?brands={brand_to_search}", name="/products?brands=<brand>")
    
    @task(2100)  # Category-only searches
    def search_by_category(self):
        """
        Search only by category
        """
        category_to_search = random.choice(self.common_categories)
        self.client.get(f"/products?categories={category_to_search}", name="/products?categories=<category>")

    