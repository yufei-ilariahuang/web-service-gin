from flask import Flask, request, jsonify
import boto3
import os
import math

app = Flask(__name__)
s3 = boto3.client('s3')

# Get the S3 bucket name from an environment variable for security and flexibility
BUCKET_NAME = os.environ.get('S3_BUCKET_NAME')
if not BUCKET_NAME:
    raise ValueError("S3_BUCKET_NAME environment variable is not set.")

@app.route('/split', methods=['GET'])
def split_file():
    """
    Splits a text file from S3 into three chunks and saves them back to S3.
    Expects a GET request with a query parameter: ?s3_key=path/to/your/file.txt
    """
    s3_key = request.args.get('s3_key')
    if not s3_key:
        return jsonify({"error": "s3_key parameter is required"}), 400

    print(f"Processing key: {s3_key} from bucket: {BUCKET_NAME}")

    # 1. Download the original file from S3
    try:
        response = s3.get_object(Bucket=BUCKET_NAME, Key=s3_key)
        # keepends=True preserves newlines, making the chunks valid text files
        lines = response['Body'].read().decode('utf-8').splitlines(keepends=True)
    except s3.exceptions.NoSuchKey:
        return jsonify({"error": f"File not found in S3: {s3_key}"}), 404
    except Exception as e:
        return jsonify({"error": f"Failed to download from S3: {str(e)}"}), 500

    # 2. Split the content into 3 chunks
    total_lines = len(lines)
    # Using math.ceil ensures all lines are included, even if not perfectly divisible
    lines_per_chunk = math.ceil(total_lines / 3)
    
    chunks = []
    for i in range(0, total_lines, lines_per_chunk):
        chunk_lines = lines[i:i + lines_per_chunk]
        chunks.append("".join(chunk_lines))

    # 3. Upload the chunks back to S3
    chunk_keys = []
    base_filename = os.path.splitext(os.path.basename(s3_key))[0]

    for i, chunk_content in enumerate(chunks):
        chunk_key = f"chunks/{base_filename}_chunk_{i+1}.txt"
        try:
            s3.put_object(
                Bucket=BUCKET_NAME,
                Key=chunk_key,
                Body=chunk_content.encode('utf-8')
            )
            chunk_keys.append(chunk_key)
            print(f"Successfully uploaded {chunk_key}")
        except Exception as e:
            return jsonify({"error": f"Failed to upload chunk {i+1} to S3: {str(e)}"}), 500

    # 4. Return the S3 keys of the new chunks
    return jsonify({"chunk_s3_keys": chunk_keys})

if __name__ == '__main__':
    # Listens on all available network interfaces, required for Docker
    app.run(host='0.0.0.0', port=8080)