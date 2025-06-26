# ---------------------------------------------
## Quick example: Extract structured data using Retab's all-in-one `.parse()` method.
# ---------------------------------------------

import os
import json

from dotenv import load_dotenv

from retab import Retab

# Load environment variables
load_dotenv()

retab_api_key = os.getenv("RETAB_API_KEY")
assert retab_api_key, "Missing RETAB_API_KEY"

# Retab Setup
reclient = Retab(api_key=retab_api_key)

# Document Extraction via Retab API
response = reclient.documents.extract(
    document="../../assets/code/invoice.jpeg",
    model="gpt-4.1",
    json_schema="../../assets/code/invoice_schema.json",
    modality="native",
    image_resolution_dpi=96,
    browser_canvas="A4",
    temperature=0.0,
    n_consensus=4,  # Number of parallel extraction runs for consensus voting. Values > 1 enable consensus mode.
)

# Output
print("CONSENSUS EXTRACTION RESULTS")

for i in range(len(response.choices)):
    print(f"\nConsensus Result #{i + 1}:")
    print("-" * 40)
    try:
        content = response.choices[i].message.content
        if content and (content.strip().startswith("{") or content.strip().startswith("[")):
            parsed = json.loads(content)
            print(json.dumps(parsed, indent=2, ensure_ascii=False))
        else:
            print(content if content else "No content")
    except (json.JSONDecodeError, AttributeError):
        content = response.choices[i].message.content
        print(content if content else "No content")

# Display likelihoods with better formatting
if hasattr(response, "likelihoods") and response.likelihoods:
    print("\nCONSENSUS LIKELIHOODS:")
    print("-" * 40)

    # Handle both list and dict formats for likelihoods
    if isinstance(response.likelihoods, list):
        for i, likelihood in enumerate(response.likelihoods):
            print(f"Result #{i + 1}: {likelihood:.4f}")

        # Show the most confident result
        if response.likelihoods:
            best_idx = response.likelihoods.index(max(response.likelihoods))
            print(f"\n🎯 Most confident result: #{best_idx + 1} (likelihood: {max(response.likelihoods):.4f})")
    else:
        # Handle dict format
        for key, likelihood in response.likelihoods.items():
            print(f"{key}: {likelihood:.4f}")
