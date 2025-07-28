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
    "User-Agent": "YourName/1.0 (your.email@example.com)",
    "Accept-Encoding": "gzip, deflate"
}

# 1) Get the latest 8-K entry
def get_latest_8k_entry():
    feed = feedparser.parse(SEC_RSS_FEED, request_headers=HEADERS)
    for entry in feed.entries:
        if entry.get("edgar_formtype") == "8-K":
            return entry
    raise RuntimeError("No 8-K entries found")

# 2) Resolve the instance URL (Takes the index URL > Finds the filing directory > Downloads the MetaLinks.json > Locates the instance document > Ruturns the actual document URL)
def resolve_instance_url(entry):
    # build the base directory URL for this filing
    idx_url = entry["link"]  # e.g. ...-index.htm
    base_dir = idx_url.rsplit("/", 1)[0] + "/"
    # fetch MetaLinks.json to enumerate instance files
    meta_url = urljoin(SEC_ROOT, base_dir + "MetaLinks.json")
    r = requests.get(meta_url, headers=HEADERS, timeout=15)
    r.raise_for_status()
    meta = r.json()
    instances = meta.get("instance", {})
    # pick the first .htm instance that's not the viewer
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
    soup = BeautifulSoup(raw_html, "xml")
    # extract all XBRL facts (nonNumeric & nonFraction)
    facts = []
    for tag in soup.find_all(re.compile(r"^ix:(nonNumeric|nonFraction)$")):
        facts.append({
            "name":       tag.get("name"),
            "contextRef": tag.get("contextRef"),
            "unitRef":    tag.get("unitRef"),
            "decimals":   tag.get("decimals"),
            "value":      tag.text.strip()
        })
    # extract contexts
    contexts = {}
    for ctx in soup.find_all("xbrli:context"):
        cid = ctx.get("id")
        period = ctx.find("xbrli:period")
        start = period.find("xbrli:startDate")
        end   = period.find("xbrli:endDate")
        instant = period.find("xbrli:instant")
        contexts[cid] = {
            "startDate":  start.text if start else None,
            "endDate":    end.text   if end   else None,
            "instant":    instant.text if instant else None
        }
    # extract narrative sections (plain HTML) under <body>
    body_html = soup.find("body")
    narrative = ""
    if body_html:
        # remove all ix:* tags
        for tag in body_html.find_all(re.compile(r"^ix:")):
            tag.decompose()
        narrative = body_html.get_text("\n", strip=True)
    return facts, contexts, narrative


# 4) Main function—Structure the data with retab 
def main():
    entry = get_latest_8k_entry()
    print("Latest 8-K:", entry["edgar_companyname"])
    inst_url = resolve_instance_url(entry)
    raw_html = download_raw_instance(inst_url)
    facts, contexts, narrative = parse_ixbrl(raw_html)
    
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
