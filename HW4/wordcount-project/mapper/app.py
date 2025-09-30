from flask import Flask, request, jsonify
import boto3
import os
import re
import json
from collections import Counter

app = Flask(__name__)
s3 = boto3.client('s3')

# Get bucket name from environment variable for flexibility
BUCKET_NAME = os.environ.get('S3_BUCKET_NAME')

def count_words(text):
    # Simple word count: lowercase and find all words
    words = re.findall(r'\b\w+\b', text.lower())
    return Counter(words)

@app.route('/map', methods=['GET'])
def map_chunk():
    # Get the S3 key (filename) of the chunk from the request
    s3_key = request.args.get('s3_key')
    if not s3_key:
        return jsonify({"error": "s3_key parameter is required"}), 400

    # Download the chunk from S3
    try:
        response = s3.get_object(Bucket=BUCKET_NAME, Key=s3_key)
        text_chunk = response['Body'].read().decode('utf-8')
    except Exception as e:
        return jsonify({"error": f"Failed to download from S3: {str(e)}"}), 500

    # Count the words
    word_counts = count_words(text_chunk)

    # Save the result as a JSON file in S3
    result_key = f"results/mapper-{s3_key}.json"
    try:
        s3.put_object(
            Bucket=BUCKET_NAME,
            Key=result_key,
            Body=json.dumps(word_counts, indent=2),
            ContentType='application/json'
        )
    except Exception as e:
        return jsonify({"error": f"Failed to upload result to S3: {str(e)}"}), 500

    # Return the S3 URL of the result
    return jsonify({"result_s3_key": result_key})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)