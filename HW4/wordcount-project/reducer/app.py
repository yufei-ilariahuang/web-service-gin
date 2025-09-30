from flask import Flask, request, jsonify
import boto3
import os
import json
from collections import Counter

app = Flask(__name__)
s3 = boto3.client('s3')

# Get the S3 bucket name from an environment variable
BUCKET_NAME = os.environ.get('S3_BUCKET_NAME')
if not BUCKET_NAME:
    raise ValueError("S3_BUCKET_NAME environment variable is not set.")

@app.route('/reduce', methods=['GET'])
def reduce_files():
    """
    Aggregates word counts from multiple S3 JSON files into a final result.
    Expects a GET request with a query parameter: 
    ?s3_keys=key1.json,key2.json,key3.json
    """
    s3_keys_str = request.args.get('s3_keys')
    if not s3_keys_str:
        return jsonify({"error": "s3_keys parameter is required"}), 400

    s3_keys = s3_keys_str.split(',')
    final_counts = Counter()

    print(f"Reducing keys: {s3_keys} from bucket: {BUCKET_NAME}")

    # 1. Download and process each mapper result file
    for key in s3_keys:
        try:
            response = s3.get_object(Bucket=BUCKET_NAME, Key=key)
            mapper_counts = json.loads(response['Body'].read().decode('utf-8'))
            
            # The Counter object makes aggregation simple
            final_counts.update(mapper_counts)
        except s3.exceptions.NoSuchKey:
             return jsonify({"error": f"Mapper result file not found in S3: {key}"}), 404
        except Exception as e:
            return jsonify({"error": f"Failed to process key {key}: {str(e)}"}), 500

    # 2. Upload the final aggregated result to S3
    result_key = "final_results/final_word_counts.json"
    try:
        # Sort by count descending for a nice, readable output file
        sorted_final_counts = dict(sorted(final_counts.items(), key=lambda item: item[1], reverse=True))
        
        s3.put_object(
            Bucket=BUCKET_NAME,
            Key=result_key,
            Body=json.dumps(sorted_final_counts, indent=2),
            ContentType='application/json'
        )
        print(f"Successfully uploaded final results to {result_key}")
    except Exception as e:
        return jsonify({"error": f"Failed to upload final result to S3: {str(e)}"}), 500

    # 3. Return the S3 key of the final result file
    return jsonify({
        "final_result_s3_key": result_key,
        "total_unique_words": len(final_counts)
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)