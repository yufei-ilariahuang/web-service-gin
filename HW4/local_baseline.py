# ===============================================
# FILE: local_baseline.py
# DESCRIPTION: Performs word count locally on the large file and records the time.
# USAGE: python local_baseline.py
# ===============================================

import time
from collections import Counter
import re
import os

# NOTE: Set this to the name of your large input file
LOCAL_FILE_PATH = 'big.txt' 

def count_words_locally(filepath):
    """Performs word count and times the core operation."""
    if not os.path.exists(filepath):
        print(f"Error: File not found at {filepath}. Please download the large file first.")
        return None
        
    print(f"Starting local word count on {filepath}...")
    
    # Read the file content
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            text = f.read()
    except Exception as e:
        print(f"Error reading file: {e}")
        return None
    
    # --- CORE WORD COUNT LOGIC (MUST MATCH MAPPER) ---
    start_time = time.time()
    
    # Tokenize: Find all word boundaries and convert to lowercase
    # This regex matches sequences of word characters
    words = re.findall(r'\b\w+\b', text.lower())
    
    # Count the words
    counts = Counter(words)
    
    end_time = time.time()
    # --------------------------------------------------
    
    time_taken = end_time - start_time
    
    print(f"Local, single-threaded time: {time_taken:.4f} seconds")
    print(f"Total unique words found: {len(counts)}")
    
    return time_taken

# Execute the baseline
local_time = count_words_locally(LOCAL_FILE_PATH)

# To save the time for the Jupyter Notebook
if local_time is not None:
    try:
        with open('local_time2.txt', 'w') as f:
            f.write(str(local_time))
        print("Time saved to local_time2.txt")
    except Exception as e:
        print(f"Error saving time to file: {e}")