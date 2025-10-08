from flask import Flask, jsonify, request
import uuid

app = Flask(__name__)

# In-memory data store (a dictionary acting as a hashmap)
products = {}

@app.route('/products', methods=['GET'])
def get_products():
    """
    Retrieves a list of all products.
    """
    return jsonify(list(products.values()))

@app.route('/products', methods=['POST'])
def create_product():
    """
    Creates a new product.
    """
    product_data = request.get_json()

    # Input validation
    if not product_data or 'name' not in product_data or 'price' not in product_data:
        return jsonify({'error': 'Invalid input'}), 400

    product_id = str(uuid.uuid4())
    new_product = {
        'id': product_id,
        'name': product_data['name'],
        'description': product_data.get('description', ''),
        'price': product_data['price']
    }
    products[product_id] = new_product

    return jsonify(new_product), 201

@app.route('/products/<string:product_id>', methods=['DELETE'])
def delete_product(product_id):
    """
    Deletes a product by its ID.
    """
    if product_id in products:
        del products[product_id]
        return '', 204  # No Content
    else:
        return jsonify({'error': 'Product not found'}), 404 # Not Found

@app.route('/products/<string:product_id>', methods=['GET'])
def get_product(product_id):
    
    if product_id not in products:
        return jsonify({'error': 'Product not found'}), 404
    else:
        product = products[product_id]
        return jsonify(product)
    
if __name__ == '__main__':
    # Adding some initial data for easier testing of DELETE
    initial_id = str(uuid.uuid4())
    products[initial_id] = {
        'id': initial_id,
        'name': 'Sample Product',
        'description': 'This is a sample product to test deletion.',
        'price': 99.99
    }
    app.run(host='0.0.0.0', port=5002, debug=True)