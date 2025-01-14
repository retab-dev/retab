from typing import TypedDict, List
import json
import requests
from PIL import Image
from io import BytesIO
import base64
from math import ceil
from pathlib import Path
import tiktoken  # For text tokenization


class TokenStats(TypedDict):
    min: float
    max: float
    mean: float
    median: float
    p5: float
    p95: float

class TokenCounts(TypedDict):
    system_text_tokens: int
    user_text_tokens: int
    assistant_text_tokens: int
    image_tokens: int

class MetricCategory(TypedDict):
    num_examples: int
    total_tokens: TokenStats
    assistant_tokens: TokenStats
    sum_total_tokens: float
    sum_assistant_tokens: float
    num_examples_over_token_limit: int

class Metrics(TypedDict):
    Text: MetricCategory
    Image: MetricCategory
    Total: MetricCategory


def count_text_tokens(content: str, encoding_name: str = "cl100k_base") -> int:
    """
    Count the number of tokens in a given text content using the specified encoding.
    """
    enc = tiktoken.get_encoding(encoding_name)
    return len(enc.encode(content))



def count_image_tokens(image_url: str) -> int:
    """
    Count the number of tokens for an image. If the image is a base64-encoded data URL,
    calculate dimensions from the decoded image metadata (if available).
    For HTTP URLs, try to fetch dimensions; fallback to default tile count if not accessible.
    """
    base_token_cost = 85  # Base token cost for any image
    token_per_tile = 170  # Token cost per 512x512 tile

    if image_url.startswith("data:image"):
        # Parse base64 data
        try:
            header, encoded_data = image_url.split(",", 1)
            image_data = base64.b64decode(encoded_data)
            # Decode image dimensions using PIL
            image = Image.open(BytesIO(image_data))
            width, height = image.size
        except Exception:
            return base_token_cost  # If decoding fails, assume base token cost only
    elif image_url.startswith("http"):
        # Try to fetch the image and get its dimensions
        try:
            response = requests.get(image_url, timeout=5)
            response.raise_for_status()
            image = Image.open(BytesIO(response.content))
            width, height = image.size
        except Exception:
            return base_token_cost + token_per_tile  # Fallback to 1 tile
    else:
        # Unsupported image format
        return base_token_cost + token_per_tile  # Default to 1 tile

    # Calculate number of tiles
    tiles_wide = ceil(width / 512)
    tiles_high = ceil(height / 512)
    total_tiles = tiles_wide * tiles_high

    return base_token_cost + (token_per_tile * total_tiles)


def process_jsonl_file(jsonl_path: str) -> List[TokenCounts]:
    """
    Process a JSONL file and calculate the text and image tokens for each example.
    Returns a list of dictionaries with token counts for system, user, assistant, and images.
    """
    results = []

    with open(jsonl_path, "r", encoding="utf-8") as file:
        for line in file:
            example = json.loads(line)
            system_text_tokens = 0
            user_text_tokens = 0
            assistant_text_tokens = 0
            image_tokens = 0

            for message in example.get("messages", []):
                role = message.get("role")
                content = message.get("content")

                if isinstance(content, str):
                    # Count text tokens based on role
                    if role == "system":
                        system_text_tokens += count_text_tokens(content)
                    elif role == "user":
                        user_text_tokens += count_text_tokens(content)
                    elif role == "assistant":
                        assistant_text_tokens += count_text_tokens(content)

                elif isinstance(content, list):  # Check for images in content
                    for item in content:
                        if item.get("type") == "image_url" and "image_url" in item:
                            image_url = item["image_url"]["url"]
                            image_tokens += count_image_tokens(image_url)

            results.append(TokenCounts(
                system_text_tokens=system_text_tokens,
                user_text_tokens=user_text_tokens,
                assistant_text_tokens=assistant_text_tokens,
                image_tokens=image_tokens
            ))

    return results

import json
import numpy as np
from typing import List, Dict
from rich.table import Table
from rich.console import Console

def calculate_statistics(data: List[int]) -> TokenStats:
    """
    Calculate statistics for a list of numbers.
    """
    if not data:
        return {"min": 0, "max": 0, "mean": 0, "median": 0, "p5": 0, "p95": 0}
    
    return {
        "min": float(min(data)),
        "max": float(max(data)),
        "mean": float(np.mean(data)),
        "median": float(np.median(data)),
        "p5": float(np.percentile(data, 5)),
        "p95": float(np.percentile(data, 95)),
    }

def process_dataset_and_compute_metrics(jsonl_path: Path | str, token_limit: int = 128000) -> Metrics:
    """
    Process the dataset to compute metrics for Text, Image, and Total tokens.
    """
    # Initialize metrics
    metrics: Metrics = {
        "Text": MetricCategory(
            num_examples=0,
            total_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            assistant_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            sum_total_tokens=0,
            sum_assistant_tokens=0,
            num_examples_over_token_limit=0,
        ),
        "Image": MetricCategory(
            num_examples=0,
            total_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            assistant_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            sum_total_tokens=0,
            sum_assistant_tokens=0,
            num_examples_over_token_limit=0,
        ),
        "Total": MetricCategory(
            num_examples=0,
            total_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            assistant_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            sum_total_tokens=0,
            sum_assistant_tokens=0,
            num_examples_over_token_limit=0,
        )
    }

    # Accumulate token counts
    text_total_tokens = []
    image_total_tokens = []
    text_assistant_tokens = []
    image_assistant_tokens = []
    total_tokens = []
    total_assistant_tokens = []

    with open(jsonl_path, "r", encoding="utf-8") as file:
        for line in file:
            example = json.loads(line)

            system_text_tokens = 0
            system_image_tokens = 0

            user_text_tokens = 0
            user_image_tokens = 0

            assistant_text_tokens = 0
            assistant_image_tokens = 0

            for message in example.get("messages", []):
                role = message.get("role")
                content = message.get("content")

                if isinstance(content, str):
                    if role == "system":
                        system_text_tokens += count_text_tokens(content)
                    elif role == "user":
                        user_text_tokens += count_text_tokens(content)
                    elif role == "assistant":
                        assistant_text_tokens += count_text_tokens(content)
                elif isinstance(content, list):  # Handle images
                    for item in content:
                        if item.get("type") == "image_url" and "image_url" in item:
                            image_url = item["image_url"]["url"]
                            tokens = count_image_tokens(image_url)
                            if role == "system":
                                system_image_tokens += tokens
                            elif role == "user":
                                user_image_tokens += tokens
                            elif role == "assistant":
                                assistant_image_tokens+= tokens

            # Calculate totals for the example
            example_text_tokens = system_text_tokens + user_text_tokens + assistant_text_tokens
            example_total_tokens = example_text_tokens + system_image_tokens + user_image_tokens

            # Add to accumulators
            text_total_tokens.append(example_text_tokens)
            text_assistant_tokens.append(assistant_text_tokens)
            image_total_tokens.append(system_image_tokens + user_image_tokens)
            image_assistant_tokens.append(assistant_image_tokens)
            total_tokens.append(example_total_tokens)
            total_assistant_tokens.append(assistant_text_tokens + assistant_image_tokens)

            # Count examples over token limit
            if example_text_tokens > token_limit:
                metrics["Text"]["num_examples_over_token_limit"] += 1
            if system_image_tokens + user_image_tokens > token_limit:
                metrics["Image"]["num_examples_over_token_limit"] += 1
            if example_total_tokens > token_limit:
                metrics["Total"]["num_examples_over_token_limit"] += 1
                #print(example_total_tokens, token_limit)

    # Update metrics for Text, Image, and Total
    metrics["Text"]["num_examples"] = len(text_total_tokens)
    metrics["Text"]["total_tokens"] = calculate_statistics(text_total_tokens)
    metrics["Text"]["assistant_tokens"] = calculate_statistics(text_assistant_tokens)
    metrics["Text"]["sum_assistant_tokens"] = sum(text_assistant_tokens)
    metrics["Text"]["sum_total_tokens"] = sum(text_total_tokens)

    metrics["Image"]["num_examples"] = len(image_total_tokens)
    metrics["Image"]["total_tokens"] = calculate_statistics(image_total_tokens)
    metrics["Image"]["assistant_tokens"] = calculate_statistics(image_assistant_tokens)
    metrics["Image"]["sum_assistant_tokens"] = sum(image_assistant_tokens)
    metrics["Image"]["sum_total_tokens"] = sum(image_total_tokens)

    metrics["Total"]["num_examples"] = len(total_tokens)
    metrics["Total"]["total_tokens"] = calculate_statistics(total_tokens)
    metrics["Total"]["assistant_tokens"] = calculate_statistics(total_assistant_tokens)
    metrics["Total"]["sum_assistant_tokens"] = sum(total_assistant_tokens)
    metrics["Total"]["sum_total_tokens"] = sum(total_tokens)

    return metrics

def display_metrics(metrics: Metrics) -> None:
    """
    Display the metrics dictionary in a compact table with min/max, mean/median, and p5/p95 on the same row.
    """
    console = Console(style="on grey23")
    table = Table(title="Token Metrics", show_lines=True)

    # Add columns
    table.add_column("Metric", justify="left", style="cyan", no_wrap=True)
    table.add_column("Text", justify="right", style="magenta")
    table.add_column("Image", justify="right", style="green")
    table.add_column("Total", justify="right", style="yellow")

    # Add rows
    table.add_row(
        "Num Examples", 
        str(metrics["Text"]["num_examples"]), 
        str(metrics["Image"]["num_examples"]), 
        str(metrics["Total"]["num_examples"])
    )

    table.add_row(
        "Min / Max Tokens", 
        f"{metrics['Text']['total_tokens']['min']:.2f} / {metrics['Text']['total_tokens']['max']:.2f}", 
        f"{metrics['Image']['total_tokens']['min']:.2f} / {metrics['Image']['total_tokens']['max']:.2f}", 
        f"{metrics['Total']['total_tokens']['min']:.2f} / {metrics['Total']['total_tokens']['max']:.2f}"
    )

    table.add_row(
        "Mean / Median Tokens", 
        f"{metrics['Text']['total_tokens']['mean']:.2f} / {metrics['Text']['total_tokens']['median']:.2f}", 
        f"{metrics['Image']['total_tokens']['mean']:.2f} / {metrics['Image']['total_tokens']['median']:.2f}", 
        f"{metrics['Total']['total_tokens']['mean']:.2f} / {metrics['Total']['total_tokens']['median']:.2f}"
    )

    table.add_row(
        "P5 / P95 Tokens", 
        f"{metrics['Text']['total_tokens']['p5']:.2f} / {metrics['Text']['total_tokens']['p95']:.2f}", 
        f"{metrics['Image']['total_tokens']['p5']:.2f} / {metrics['Image']['total_tokens']['p95']:.2f}", 
        f"{metrics['Total']['total_tokens']['p5']:.2f} / {metrics['Total']['total_tokens']['p95']:.2f}"
    )

    table.add_row(
        "Sum Total Tokens", 
        f"{metrics['Text']['sum_total_tokens']:.2f}", 
        f"{metrics['Image']['sum_total_tokens']:.2f}", 
        f"{metrics['Total']['sum_total_tokens']:.2f}"
    )

    # Rows for assistant_tokens
    table.add_row(
        "Min / Max Assistant Tokens", 
        f"{metrics['Text']['assistant_tokens']['min']:.2f} / {metrics['Text']['assistant_tokens']['max']:.2f}", 
        f"{metrics['Image']['assistant_tokens']['min']:.2f} / {metrics['Image']['assistant_tokens']['max']:.2f}", 
        f"{metrics['Total']['assistant_tokens']['min']:.2f} / {metrics['Total']['assistant_tokens']['max']:.2f}"
    )

    table.add_row(
        "Mean / Median Assistant Tokens", 
        f"{metrics['Text']['assistant_tokens']['mean']:.2f} / {metrics['Text']['assistant_tokens']['median']:.2f}", 
        f"{metrics['Image']['assistant_tokens']['mean']:.2f} / {metrics['Image']['assistant_tokens']['median']:.2f}", 
        f"{metrics['Total']['assistant_tokens']['mean']:.2f} / {metrics['Total']['assistant_tokens']['median']:.2f}"
    )

    table.add_row(
        "P5 / P95 Assistant Tokens", 
        f"{metrics['Text']['assistant_tokens']['p5']:.2f} / {metrics['Text']['assistant_tokens']['p95']:.2f}", 
        f"{metrics['Image']['assistant_tokens']['p5']:.2f} / {metrics['Image']['assistant_tokens']['p95']:.2f}", 
        f"{metrics['Total']['assistant_tokens']['p5']:.2f} / {metrics['Total']['assistant_tokens']['p95']:.2f}"
    )

    table.add_row(
        "Sum Assistant Tokens", 
        f"{metrics['Text']['sum_assistant_tokens']:.2f}", 
        f"{metrics['Image']['sum_assistant_tokens']:.2f}", 
        f"{metrics['Total']['sum_assistant_tokens']:.2f}"
    )
    
    table.add_row(
        "Examples Over Limit", 
        str(metrics["Text"]["num_examples_over_token_limit"]), 
        str(metrics["Image"]["num_examples_over_token_limit"]), 
        str(metrics["Total"]["num_examples_over_token_limit"])
    )

    # Print the table
    console.print(table)
