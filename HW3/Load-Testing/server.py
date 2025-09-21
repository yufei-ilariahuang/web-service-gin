# server.py
from flask import Flask, request, jsonify
import time
import random

app = Flask(__name__)

# Using a dictionary (hash map) to store our data in memory.
data_store = {}

@app.route('/data/<key>', methods=['GET'])
def get_data(key):
    """Handles GET requests to retrieve data."""
    if key in data_store:
        return jsonify({key: data_store[key]})
    else:
        return jsonify({"error": "Key not found"}), 404

@app.route('/data', methods=['POST'])
def post_data():
    """Handles POST requests to add new data."""
    if not request.json or 'key' not in request.json or 'value' not in request.json:
        return jsonify({"error": "Missing key or value in request"}), 400

    key = request.json['key']
    value = request.json['value']

    # Simulate some processing or database write delay
    time.sleep(random.uniform(0.01, 0.05))

    data_store[key] = value
    return jsonify({"message": f"Successfully added {key}"}), 201

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)