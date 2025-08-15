#!/bin/bash

# Run the Python script and capture output in generated_types.ts
python3 retab/generate_types.py > ../node/src/generated_types.ts
