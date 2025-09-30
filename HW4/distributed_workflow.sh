#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

# --- 1. Configuration: UPDATE THESE IPs ---
SPLITTER_IP="44.244.30.179"
REDUCER_IP="44.250.214.213"
INPUT_KEY="example.txt" # Name of your large file

# Define the IPs for all three individual Mapper tasks
MAPPER_IPS=("34.222.13.11" "35.161.239.104" "44.251.158.158")
# ----------------------------------------

echo "--- STARTING DISTRIBUTED WORD COUNT ---"
START_TIME=$(date +%s.%N)

# 1. SPLIT (Creates chunks, response contains list of chunk S3 keys)
echo "1. Calling Splitter..."
SPLIT_RESPONSE=$(curl -s "http://${SPLITTER_IP}:8080/split?s3_key=${INPUT_KEY}")
CHUNK_KEYS=$(echo ${SPLIT_RESPONSE} | jq -r '.chunk_s3_keys[]') 

if [ -z "$CHUNK_KEYS" ]; then
    echo "ERROR: Splitter failed or returned no chunks."
    exit 1
fi
echo "   Chunks created: $(echo ${SPLIT_RESPONSE} | jq -r '.chunk_s3_keys | length') chunks."

# Prepare a combined list of 'IP:S3_KEY' pairs
# Assuming: Chunk 1 goes to IP 1, Chunk 2 to IP 2, etc.
# This logic requires the Splitter to return chunks in a predictable order (e.g., chunk_1, chunk_2, chunk_3).
CHUNK_KEY_ARRAY=($CHUNK_KEYS)
REQUEST_LIST=()
for i in "${!CHUNK_KEY_ARRAY[@]}"; do
    # Format: IP,S3_KEY
    REQUEST_LIST+=("${MAPPER_IPS[i]},${CHUNK_KEY_ARRAY[i]}")
done

# 2. MAP (Run in parallel, targeting the specific IP for each chunk)
echo "2. Calling Mappers in parallel (P=3) to different IPs..."

# The inner function processes a single "IP,S3_KEY" pair
MAPPER_OUTPUT_KEYS=$(printf '%s\n' "${REQUEST_LIST[@]}" | xargs -I {} -P 3 bash -c \
    'IFS=, read -r MAPPER_IP S3_KEY <<< "$0"; \
     curl -s "http://${MAPPER_IP}:8080/map?s3_key=${S3_KEY}"' \
    {})

# Consolidate all mapper result keys into a comma-separated string for the Reducer
REDUCER_KEYS=$(echo "${MAPPER_OUTPUT_KEYS}" | jq -r '.result_s3_key' | tr '\n' ',' | sed 's/,$//')

if [ -z "$REDUCER_KEYS" ]; then
    echo "ERROR: Mappers failed or Reducer keys string is empty."
    exit 1
fi
echo "   Mapper results key string: ${REDUCER_KEYS}"

# 3. REDUCE
echo "3. Calling Reducer..."
REDUCE_RESPONSE=$(curl -s "http://${REDUCER_IP}:8080/reduce?s3_keys=${REDUCER_KEYS}")

END_TIME=$(date +%s.%N)

# --- Performance Analysis ---
DISTRIBUTED_TIME=$(echo "$END_TIME - $START_TIME" | bc)

echo "--- FINISHED ---"
echo "Reducer Response: ${REDUCE_RESPONSE}"
echo "Total Distributed Time: ${DISTRIBUTED_TIME} seconds"

# Save time for plotting
echo "${DISTRIBUTED_TIME}" > distributed_time.txt
echo "Distributed time saved to distributed_time.txt"