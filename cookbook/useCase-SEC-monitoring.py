"""
Business Use Cases:

- Financial Monitoring: Track major corporate events as they're filed
- Investment Research: Quickly extract key information from SEC filings
- Compliance Tracking: Monitor regulatory filings automatically
- Data Pipeline: Feed structured data into analysis systems or databases

Why This Matters?

8-K filings contain critical information like:
- Mergers & acquisitions
- Executive changes
- Material agreements
- Earnings releases
- Legal proceedings

Share your own use-cases with us on our [Discord](https://discord.com/invite/vc5tWRPqag) or [Twitter](https://x.com/retabdev)!
"""

import re
import json, textwrap
import feedparser
import requests
from bs4 import BeautifulSoup
from urllib.parse import urljoin
from dotenv import load_dotenv

from retab import Retab

load_dotenv()

# Constants
SEC_RSS_FEED = "https://www.sec.gov/Archives/edgar/xbrlrss.all.xml"
SEC_ROOT     = "https://www.sec.gov"
HEADERS      = {
    "User-Agent": "YourName/1.0 (your.email@example.com)", # Replace with your own email
    "Accept-Encoding": "gzip, deflate"
}

# 1) Get the latest 8-K entry
def get_latest_8k_entry():
    feed = feedparser.parse(SEC_RSS_FEED, request_headers=HEADERS)
    for entry in feed.entries:
        if entry.get("edgar_formtype") == "8-K":
            return entry
    raise RuntimeError("No 8-K entries found")

# 2) Resolve the instance URL (Takes the index URL > Finds the filing directory > Downloads the MetaLinks.json > Locates the instance document > Returns the actual document URL)
def resolve_instance_url(entry):
    idx_url  = entry["link"]
    base_dir = idx_url.rsplit("/", 1)[0] + "/"
    meta_url = urljoin(SEC_ROOT, base_dir + "MetaLinks.json")
    r = requests.get(meta_url, headers=HEADERS, timeout=15)
    r.raise_for_status()
    meta = r.json()
    instances = list(meta.get("instance", {}).keys())

    # prefer likely primary docs
    instances.sort(key=lambda f: (
        0 if ("8-k" in f.lower() or "primary" in f.lower() or "document" in f.lower()) else 1,
        len(f)
    ))

    for fname in instances:
        if fname.lower().endswith(".htm") and not fname.lower().endswith("-index.htm"):
            return urljoin(SEC_ROOT, base_dir + fname)

    raise RuntimeError("No instance HTML file found in MetaLinks.json")

def download_raw_instance(url):
    r = requests.get(url, headers=HEADERS, timeout=15)
    r.raise_for_status()
    return r.text

# 3) Parse the IXBRL document (Removes all ix:* tags and returns the narrative)
def parse_ixbrl(raw_html):
    # parse as HTML
    soup = BeautifulSoup(raw_html, "html.parser")

    # remove non-content tags
    for tag in soup(["script", "style", "noscript"]):
        tag.decompose()

    # KEEP the text inside ix:* tags 
    for tag in soup.find_all(re.compile(r"^ix:"), recursive=True):
        tag.unwrap()

    # grab body text if present, otherwise whole doc
    node = soup.body or soup
    text = node.get_text("\n", strip=True)

    # normalize whitespace a bit
    text = re.sub(r"\n{2,}", "\n", text)
    return text


# 4) Main function—Structure the data with retab 
def main():
    entry = get_latest_8k_entry()
    print("Latest 8-K:", entry["edgar_companyname"])
    inst_url = resolve_instance_url(entry)
    raw_html = download_raw_instance(inst_url)
    narrative = parse_ixbrl(raw_html)
    
    # Save narrative for Retab processing
    with open("narrative.txt", "w", encoding="utf-8") as f:
        f.write(narrative)
    
    # Process with Retab
    try:
        client = Retab()

        completion = client.deployments.extract(
            project_id="proj_6Z821A5qH5P4zDNFd73zN", # Use the processor configured on retab's platform
            iteration_id="base-configuration",       # Use the iteration that gives you the best accuracy / inference speed / price
            document="narrative.txt"
        )
        
        # Extract and print clean structured data
        parsed_data = json.loads(completion.choices[0].message.content)
        formatted_json = json.dumps(parsed_data, indent=2, ensure_ascii=False)
        print(textwrap.indent(formatted_json, "  "))
        
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
