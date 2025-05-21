import base64
import json
from io import BytesIO
from math import ceil
from pathlib import Path
from typing import List, Literal, Optional, TypedDict

import numpy as np
import requests
import tiktoken  # For text tokenization
from PIL import Image
from rich.console import Console
from rich.table import Table


class TokenStats(TypedDict):
    min: float
    max: float
    mean: float
    median: float
    p5: float
    p95: float


class TokenCounts(TypedDict):
    input_text_tokens: int
    output_text_tokens: int
    input_image_tokens: int
    output_image_tokens: int


class MetricCategory(TypedDict):
    num_examples: int
    total_tokens: TokenStats
    input_tokens: TokenStats
    output_tokens: TokenStats
    sum_total_tokens: float
    sum_input_tokens: float
    sum_output_tokens: float
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


def count_image_tokens(image_url: str, detail: Literal["low", "high", "auto"] = "high") -> int:
    base_token_cost = 85  # cost for all images
    token_per_tile = 170  # cost per 512×512 tile in high detail

    # 1. Decide detail=low or detail=high
    #    If detail=auto, figure out from user input or some heuristic
    if detail == "low":
        # 2. Low detail => always 85 tokens
        return base_token_cost
    else:
        assert detail == "high" or detail == "auto"
        # 3. High detail => 2-step scaling + tile-based cost

        # (a) Get the raw image dimensions
        try:
            if image_url.startswith("data:image"):
                header, encoded_data = image_url.split(",", 1)
                image_data = base64.b64decode(encoded_data)
                img = Image.open(BytesIO(image_data))
            else:
                # HTTP URL or local path
                response = requests.get(image_url, timeout=5)
                response.raise_for_status()
                img = Image.open(BytesIO(response.content))

            width, height = img.size
        except Exception:
            # If we fail to decode or fetch, maybe return the base cost
            # plus one tile as a fallback
            return base_token_cost + token_per_tile

        # (b) Scale so neither dimension exceeds 2048
        max_side = max(width, height)
        if max_side > 2048:
            scale_factor = 2048.0 / max_side
            width = int(width * scale_factor)
            height = int(height * scale_factor)

        # (c) Upscale if shortest side < 768
        min_side = min(width, height)
        if min_side < 768:
            upscale_factor = 768.0 / min_side
            width = int(width * upscale_factor)
            height = int(height * upscale_factor)

        # (d) Count 512×512 tiles in the scaled image
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
            input_text_tokens = 0
            output_text_tokens = 0
            input_image_tokens = 0
            output_image_tokens = 0

            for message in example.get("messages", []):
                role = message.get("role")
                content = message.get("content")

                if isinstance(content, str):
                    # Count text tokens based on role
                    if role in ["developer", "system", "user"]:
                        input_text_tokens += count_text_tokens(content)
                    elif role == "assistant":
                        output_text_tokens += count_text_tokens(content)

                elif isinstance(content, list):  # Check for images in content
                    for item in content:
                        if item.get("type") == "image_url" and "image_url" in item:
                            image_url = item["image_url"]["url"]
                            tokens = count_image_tokens(image_url)
                            if role in ["developer", "system", "user"]:
                                input_image_tokens += tokens
                            elif role == "assistant":
                                output_image_tokens += tokens

            results.append(
                TokenCounts(
                    input_text_tokens=input_text_tokens, output_text_tokens=output_text_tokens, input_image_tokens=input_image_tokens, output_image_tokens=output_image_tokens
                )
            )

    return results


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
            input_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            output_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            sum_total_tokens=0,
            sum_input_tokens=0,
            sum_output_tokens=0,
            num_examples_over_token_limit=0,
        ),
        "Image": MetricCategory(
            num_examples=0,
            total_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            input_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            output_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            sum_total_tokens=0,
            sum_input_tokens=0,
            sum_output_tokens=0,
            num_examples_over_token_limit=0,
        ),
        "Total": MetricCategory(
            num_examples=0,
            total_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            input_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            output_tokens=TokenStats(min=0, max=0, mean=0, median=0, p5=0, p95=0),
            sum_total_tokens=0,
            sum_input_tokens=0,
            sum_output_tokens=0,
            num_examples_over_token_limit=0,
        ),
    }

    # Accumulate token counts
    input_text_tokens = []
    output_text_tokens = []
    messages_text_tokens = []

    input_image_tokens = []
    output_image_tokens = []
    messages_image_tokens = []

    input_total_tokens = []
    output_total_tokens = []
    messages_total_tokens = []

    with open(jsonl_path, "r", encoding="utf-8") as file:
        for line in file:
            example = json.loads(line)

            input_text_tokens_example = 0
            output_text_tokens_example = 0

            input_image_tokens_example = 0
            output_image_tokens_example = 0

            for message in example.get("messages", []):
                role = message.get("role")
                content = message.get("content")

                if isinstance(content, str):
                    if role in ["developer", "system", "user"]:
                        input_text_tokens_example += count_text_tokens(content)
                    elif role == "assistant":
                        output_text_tokens_example += count_text_tokens(content)
                elif isinstance(content, list):  # Handle images
                    for item in content:
                        if item.get("type") == "image_url" and "image_url" in item:
                            image_url = item["image_url"]["url"]
                            detail = item["image_url"]["detail"]
                            tokens = count_image_tokens(image_url, detail)
                            if role in ["developer", "system", "user"]:
                                input_image_tokens_example += tokens
                            elif role == "assistant":
                                output_image_tokens_example += tokens

                        elif item.get("type") == "text":
                            if role in ["developer", "system", "user"]:
                                input_text_tokens_example += count_text_tokens(item["text"])
                            elif role == "assistant":
                                output_text_tokens_example += count_text_tokens(item["text"])

            # Calculate totals for the example
            example_total_tokens = input_text_tokens_example + output_text_tokens_example + input_image_tokens_example + output_image_tokens_example

            # Add to accumulators
            input_text_tokens.append(input_text_tokens_example)
            output_text_tokens.append(output_text_tokens_example)
            messages_text_tokens.append(input_text_tokens_example + output_text_tokens_example)

            input_image_tokens.append(input_image_tokens_example)
            output_image_tokens.append(output_image_tokens_example)
            messages_image_tokens.append(input_image_tokens_example + output_image_tokens_example)

            input_total_tokens.append(input_text_tokens_example + input_image_tokens_example)
            output_total_tokens.append(output_text_tokens_example + output_image_tokens_example)
            messages_total_tokens.append(input_text_tokens_example + output_text_tokens_example + input_image_tokens_example + output_image_tokens_example)

            # Count examples over token limit
            if input_text_tokens_example > token_limit:
                metrics["Text"]["num_examples_over_token_limit"] += 1
            if input_image_tokens_example > token_limit:
                metrics["Image"]["num_examples_over_token_limit"] += 1
            if example_total_tokens > token_limit:
                metrics["Total"]["num_examples_over_token_limit"] += 1
                # print(example_total_tokens, token_limit)

    # Update metrics for Text, Image, and Total
    metrics["Text"]["num_examples"] = len(input_text_tokens)
    metrics["Text"]["total_tokens"] = calculate_statistics(messages_text_tokens)
    metrics["Text"]["input_tokens"] = calculate_statistics(input_text_tokens)
    metrics["Text"]["output_tokens"] = calculate_statistics(output_text_tokens)
    metrics["Text"]["sum_input_tokens"] = sum(input_text_tokens)
    metrics["Text"]["sum_output_tokens"] = sum(output_text_tokens)
    metrics["Text"]["sum_total_tokens"] = sum(messages_text_tokens)

    metrics["Image"]["num_examples"] = len(input_image_tokens)
    metrics["Image"]["total_tokens"] = calculate_statistics(messages_image_tokens)
    metrics["Image"]["input_tokens"] = calculate_statistics(input_image_tokens)
    metrics["Image"]["output_tokens"] = calculate_statistics(output_image_tokens)
    metrics["Image"]["sum_input_tokens"] = sum(input_image_tokens)
    metrics["Image"]["sum_output_tokens"] = sum(output_image_tokens)
    metrics["Image"]["sum_total_tokens"] = sum(messages_image_tokens)

    metrics["Total"]["num_examples"] = len(input_total_tokens)
    metrics["Total"]["total_tokens"] = calculate_statistics(messages_total_tokens)
    metrics["Total"]["input_tokens"] = calculate_statistics(input_total_tokens)
    metrics["Total"]["output_tokens"] = calculate_statistics(output_total_tokens)
    metrics["Total"]["sum_input_tokens"] = sum(input_total_tokens)
    metrics["Total"]["sum_output_tokens"] = sum(output_total_tokens)
    metrics["Total"]["sum_total_tokens"] = sum(messages_total_tokens)

    return metrics


def display_metrics(metrics: Metrics, input_token_price: Optional[float] = None, output_token_price: Optional[float] = None) -> None:
    """
    Display the metrics dictionary in a compact table with min/max, mean/median, and p5/p95 on the same row.
    """
    console = Console(style="on grey23")
    table = Table(title="Dataset Metrics", show_lines=True)

    # Add columns
    table.add_column("Metric", justify="left", style="#BDE8F6", no_wrap=True)
    table.add_column("Text", justify="right", style="#C2BDF6")
    table.add_column("Image", justify="right", style="#F6BDBD")
    table.add_column("Total", justify="right", style="#F6E4BD")

    # Add rows
    table.add_row("Num Examples", str(metrics["Text"]["num_examples"]), str(metrics["Image"]["num_examples"]), str(metrics["Total"]["num_examples"]))

    table.add_row(
        "Examples Over Limit",
        str(metrics["Text"]["num_examples_over_token_limit"]),
        str(metrics["Image"]["num_examples_over_token_limit"]),
        str(metrics["Total"]["num_examples_over_token_limit"]),
    )

    table.add_row("")

    # Rows for input tokens
    table.add_row(
        "Min / Max Input Tokens",
        f"{metrics['Text']['input_tokens']['min']:.0f} / {metrics['Text']['input_tokens']['max']:.0f}",
        f"{metrics['Image']['input_tokens']['min']:.0f} / {metrics['Image']['input_tokens']['max']:.0f}",
        f"{metrics['Total']['input_tokens']['min']:.0f} / {metrics['Total']['input_tokens']['max']:.0f}",
    )

    table.add_row(
        "Mean / Median Input Tokens",
        f"{metrics['Text']['input_tokens']['mean']:.0f} / {metrics['Text']['input_tokens']['median']:.0f}",
        f"{metrics['Image']['input_tokens']['mean']:.0f} / {metrics['Image']['input_tokens']['median']:.0f}",
        f"{metrics['Total']['input_tokens']['mean']:.0f} / {metrics['Total']['input_tokens']['median']:.0f}",
    )

    table.add_row(
        "P5 / P95 Input Tokens",
        f"{metrics['Text']['input_tokens']['p5']:.0f} / {metrics['Text']['input_tokens']['p95']:.0f}",
        f"{metrics['Image']['input_tokens']['p5']:.0f} / {metrics['Image']['input_tokens']['p95']:.0f}",
        f"{metrics['Total']['input_tokens']['p5']:.0f} / {metrics['Total']['input_tokens']['p95']:.0f}",
    )

    table.add_row("Sum Input Tokens", f"{metrics['Text']['sum_input_tokens']}", f"{metrics['Image']['sum_input_tokens']}", f"{metrics['Total']['sum_input_tokens']}")

    table.add_row("")  # Empty row for spacing

    # Rows for output tokens
    table.add_row(
        "Min / Max Output Tokens",
        f"{metrics['Text']['output_tokens']['min']:.0f} / {metrics['Text']['output_tokens']['max']:.0f}",
        f"{metrics['Image']['output_tokens']['min']:.0f} / {metrics['Image']['output_tokens']['max']:.0f}",
        f"{metrics['Total']['output_tokens']['min']:.0f} / {metrics['Total']['output_tokens']['max']:.0f}",
    )

    table.add_row(
        "Mean / Median Output Tokens",
        f"{metrics['Text']['output_tokens']['mean']:.0f} / {metrics['Text']['output_tokens']['median']:.0f}",
        f"{metrics['Image']['output_tokens']['mean']:.0f} / {metrics['Image']['output_tokens']['median']:.0f}",
        f"{metrics['Total']['output_tokens']['mean']:.0f} / {metrics['Total']['output_tokens']['median']:.0f}",
    )

    table.add_row(
        "P5 / P95 Output Tokens",
        f"{metrics['Text']['output_tokens']['p5']:.0f} / {metrics['Text']['output_tokens']['p95']:.0f}",
        f"{metrics['Image']['output_tokens']['p5']:.0f} / {metrics['Image']['output_tokens']['p95']:.0f}",
        f"{metrics['Total']['output_tokens']['p5']:.0f} / {metrics['Total']['output_tokens']['p95']:.0f}",
    )

    table.add_row("Sum Output Tokens", f"{metrics['Text']['sum_output_tokens']}", f"{metrics['Image']['sum_output_tokens']}", f"{metrics['Total']['sum_output_tokens']}")

    table.add_row("")  # Empty row for spacing

    # Total tokens
    table.add_row(
        "Min / Max Tokens",
        f"{metrics['Text']['input_tokens']['min']:.0f} / {metrics['Text']['input_tokens']['max']:.0f}",
        f"{metrics['Image']['input_tokens']['min']:.0f} / {metrics['Image']['input_tokens']['max']:.0f}",
        f"{metrics['Total']['input_tokens']['min']:.0f} / {metrics['Total']['input_tokens']['max']:.0f}",
    )

    table.add_row(
        "Mean / Median Tokens",
        f"{metrics['Text']['input_tokens']['mean']:.0f} / {metrics['Text']['input_tokens']['median']:.0f}",
        f"{metrics['Image']['input_tokens']['mean']:.0f} / {metrics['Image']['input_tokens']['median']:.0f}",
        f"{metrics['Total']['input_tokens']['mean']:.0f} / {metrics['Total']['input_tokens']['median']:.0f}",
    )

    table.add_row(
        "P5 / P95 Tokens",
        f"{metrics['Text']['input_tokens']['p5']:.0f} / {metrics['Text']['input_tokens']['p95']:.0f}",
        f"{metrics['Image']['input_tokens']['p5']:.0f} / {metrics['Image']['input_tokens']['p95']:.0f}",
        f"{metrics['Total']['input_tokens']['p5']:.0f} / {metrics['Total']['input_tokens']['p95']:.0f}",
    )

    table.add_row("Sum Total Tokens", f"{metrics['Text']['sum_input_tokens']}", f"{metrics['Image']['sum_input_tokens']}", f"{metrics['Total']['sum_input_tokens']}")

    table.add_row("")  # Empty row for spacing

    if input_token_price is not None:
        table.add_row(
            "Input Cost",
            f"{metrics['Text']['sum_input_tokens'] * input_token_price:.2f} USD",
            f"{metrics['Image']['sum_input_tokens'] * input_token_price:.2f} USD",
            f"{metrics['Total']['sum_input_tokens'] * input_token_price:.2f} USD",
        )

    if output_token_price is not None:
        table.add_row(
            "Output Cost",
            f"{metrics['Text']['sum_output_tokens'] * output_token_price:.2f} USD",
            f"{metrics['Image']['sum_output_tokens'] * output_token_price:.2f} USD",
            f"{metrics['Total']['sum_output_tokens'] * output_token_price:.2f} USD",
        )

    if input_token_price is not None and output_token_price is not None:
        table.add_row(
            "Total Cost",
            f"{metrics['Text']['sum_total_tokens'] * input_token_price:.2f} USD",
            f"{metrics['Image']['sum_total_tokens'] * input_token_price:.2f} USD",
            f"{metrics['Total']['sum_total_tokens'] * input_token_price:.2f} USD",
        )

    # Print the table
    console.print(table)
