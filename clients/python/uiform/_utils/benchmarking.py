import re
import unicodedata
from collections import defaultdict
from typing import Any, Literal
from Levenshtein import distance as levenshtein_distance
from pydantic import BaseModel, computed_field
from typing import Any, Optional
import numpy as np
from typing import Literal
import pandas as pd
# The goal is to leverage this piece of code to open a jsonl file and get an analysis of the performance of the model using a one-liner. 


############# BENCHMARKING MODELS #############
from itertools import zip_longest

def hamming_distance_padded(s: str, t: str) -> int:
    """
    Compute the Hamming distance between two strings, treating spaces as wildcards.
    
    Args:
        s: The first string
        t: The second string
        
    Returns:
        The Hamming distance between the two strings
    """
    return sum(a != b for a, b in zip_longest(s, t, fillvalue=' '))

def hamming_similarity(str_1: str, str_2: str) -> float:
    """
    Compute the Hamming similarity between two strings.
    
    Args:
        str_1: The first string
        str_2: The second string

    Returns:
        A float between 0 and 1, where 1 means the strings are identical
    """
    max_length = max(len(str_1), len(str_2))

    if max_length == 0:
        return 1.0
        
    dist = hamming_distance_padded(str_1, str_2)
    return 1 - (dist / max_length)
    


def jaccard_similarity(str_1: str, str_2: str) -> float:
    """
    Compute the Jaccard similarity between two strings.
    
    Args:
        str_1: The first string
        str_2: The second string

    Returns:
        A float between 0 and 1, where 1 means the strings are identical
    """
    set_a = set(str_1)
    set_b = set(str_2)
    intersection = set_a & set_b
    union = set_a | set_b
    if not union:
        return 1.0
    return len(intersection) / len(union)



def levenshtein_similarity(str_1: str, str_2: str) -> float:
    """
    Calculate similarity between two values using Levenshtein distance.
    Returns a similarity score between 0.0 and 1.0.
    """

    max_length = max(len(str_1), len(str_2))

    if max_length == 0:
        return 1.0
        
    dist = levenshtein_distance(str_1, str_2)
    return 1 - (dist / max_length)


def key_normalization(key: str) -> str:
    """This method is useful to compare keys under list indexes (that refers to the same kind of error but on different list index position)"""
    # We will replace all .{i} with .* where i is the index of the list (using regex for this)
    key_parts = key.split(".")
    new_key_parts = []
    for key_part in key_parts:
        if key_part.isdigit():
            new_key_parts.append("*")
        else:
            new_key_parts.append(key_part)
    return ".".join(new_key_parts)


dictionary_metrics = Literal["levenshtein_similarity", "jaccard_similarity", "hamming_similarity"]
def compute_dict_difference(dict1: dict[str, Any], dict2: dict[str, Any], metric: dictionary_metrics) -> dict[str, Any]:
    """
    Compute the difference between two dictionaries.
    
    Args:
        dict1: The first dictionary
        dict2: The second dictionary
        metric: The metric to use for comparison ("levenshtein_similarity", "jaccard_similarity", "hamming_similarity")
        
    Returns:
        A dictionary containing the difference between the two dictionaries
    """
    result: dict[str, Any] = {}

    if metric == "levenshtein_similarity":
        metric_function = levenshtein_similarity
    elif metric == "jaccard_similarity":
        metric_function = jaccard_similarity
    elif metric == "hamming_similarity":
        metric_function = hamming_similarity
    else : 
        raise ValueError(f"Invalid metric: {metric}")
    
    # Step 1: Coerce the None values to empty strings
    dict1_coerced = {k: "" if v is None else str(v) for k, v in dict1.items()}
    dict2_coerced = {k: "" if v is None else str(v) for k, v in dict2.items()}
    
    # Step 2: Normalize the keys
    dict1_normalized = {key_normalization(k): v for k, v in dict1_coerced.items()}
    dict2_normalized = {key_normalization(k): v for k, v in dict2_coerced.items()}
    
    # Step 3: Compute the metrics
    all_keys = set(dict1_normalized.keys()) | set(dict2_normalized.keys())
    
    for key in all_keys:
        val1 = dict1_normalized.get(key, None)
        val2 = dict2_normalized.get(key, None)
        
        # If the key is only in one dictionary, set the difference to None
        if val1 is None or val2 is None:
            result[key] = None
            continue
            
        # Calculate the appropriate similarity metric
        result[key] = metric_function(val1, val2)
    
    return result

def aggregate_dict_differences(dict_differences: list[dict[str, Any]]) -> tuple[dict[str, Any], dict[str, Any]]:
    """
    Aggregate a list of dictionary differences into a single dictionary with average values.
    
    Args:
        dict_differences: A list of dictionaries containing similarity metrics
        
    Returns:
        A tuple containing:
        - A dictionary with the average similarity metrics across all input dictionaries
        - A dictionary with the uncertainty (standard deviation) for each metric
    """
    if not dict_differences:
        return {}, {}
    
    # Initialize counters and sum dictionaries
    counts: dict[str, int] = defaultdict(int)
    sums: dict[str, float] = defaultdict(float)
    sum_squares: dict[str, float] = defaultdict(float)
    
    # Collect all keys across all dictionaries
    all_keys: set[str] = set()
    for diff_dict in dict_differences:
        all_keys.update(diff_dict.keys())
    
    # Sum up values and count occurrences for each key
    for key in all_keys:
        for diff_dict in dict_differences:
            if key in diff_dict and diff_dict[key] is not None:
                value = diff_dict[key]
                sums[key] += value
                sum_squares[key] += value * value
                counts[key] += 1
    
    # Calculate averages and uncertainties
    result: dict[str, float | None] = {}
    uncertainty: dict[str, float | None] = {}
    
    for key in all_keys:
        if counts[key] > 0:
            mean = sums[key] / counts[key]
            result[key] = mean
            
            # Calculate standard deviation if we have more than one sample
            if counts[key] > 1:
                variance = (sum_squares[key] - (sums[key] * sums[key] / counts[key])) / (counts[key] - 1)
                uncertainty[key] = max(0, variance) ** 0.5  # Ensure non-negative due to floating point errors
            else:
                uncertainty[key] = 0.0
        else:
            result[key] = None
            uncertainty[key] = None
    
    return result, uncertainty




class SingleFileEval(BaseModel):
    """
    A class for evaluating metrics between two dictionaries.
    """
    file_id: str
    dataset_membership_id_1: str
    dataset_membership_id_2: str
    schema_id: str
    schema_data_id: str
    dict_1: dict[str, float | None]
    dict_2: dict[str, float | None]

    hamming_similarity: dict[str, float | None]
    jaccard_similarity: dict[str, float | None]
    levenshtein_similarity: dict[str, float | None]


class EvalMetric(BaseModel):
    average: dict[str, float | None]
    std: dict[str, float | None]

class EvalMetrics(BaseModel): 
    dataset_membership_id_1: str
    dataset_membership_id_2: str
    schema_id: str
    schema_data_id: str
    distances: dict[dictionary_metrics, EvalMetric]

























import pandas as pd
import shutil

def flatten_dict(d: dict[str, Any], parent_key: str = '', sep: str = '.') -> dict[str, Any]:
    """Flatten a nested dictionary with dot-separated keys."""
    items: list[tuple[str, Any]] = []
    for k, v in d.items():
        new_key = f"{parent_key}{sep}{k}" if parent_key else k
        if isinstance(v, dict):
            items.extend(flatten_dict(v, new_key, sep=sep).items())
        else:
            items.append((new_key, v))
    return dict(items)


def plot_metrics_with_uncertainty(
    analysis: dict[str, float | dict],
    uncertainties: Optional[dict[str, float | dict]] = None,
    top_n: int = 20,
    ascending: bool = False
) -> None:
    """Plot a metric from analysis results using a horizontal bar chart with uncertainty.
    
    Args:
        analysis: Dictionary containing similarity scores (can be nested).
        uncertainties: Dictionary containing uncertainty values (same structure as analysis).
        top_n: Number of top fields to display.
        ascending: Whether to sort in ascending order.
    """
    # Flatten the dictionaries
    flattened_analysis = flatten_dict(analysis)
    if uncertainties:
        flattened_uncertainties = flatten_dict(uncertainties)
    else:
        uncertainties_list = None
    
    # Prepare data by matching fields
    fields = list(flattened_analysis.keys())
    similarities = [flattened_analysis[field] for field in fields]

    if uncertainties:
        uncertainties_list = [flattened_uncertainties.get(field, None) for field in fields]

    
    # Create a DataFrame
    df = pd.DataFrame({
        "field": fields,
        "similarity": similarities,
    })

    if uncertainties: 
        df["uncertainty"] = uncertainties_list
    
    # Sort by similarity and select top N
    df = df.sort_values(by="similarity", ascending=ascending).head(top_n)
    
    # Calculate layout dimensions
    label_width = max(len(field) for field in df["field"]) + 2  # Padding for alignment
    terminal_width = shutil.get_terminal_size().columns
    bar_width = terminal_width - label_width - 3  # Space for '| ' and extra padding
    
    # Determine scaling factor based on maximum similarity
    max_similarity = df["similarity"].max()
    scale = bar_width / max_similarity if max_similarity > 0 else 1
    
    # Generate and print bars
    for index, row in df.iterrows():
        field = row["field"]
        similarity = row["similarity"]
        if uncertainties:
            uncertainty = row["uncertainty"]
        else:
            uncertainty = None
        
        if similarity is None:
            continue  # Skip fields with no similarity value
        
        # Calculate bar length and uncertainty range
        bar_len = round(similarity * scale)
        if uncertainty is not None and uncertainty > 0:
            uncertainty_start = max(0, round((similarity - uncertainty) * scale))
            uncertainty_end = min(bar_width, round((similarity + uncertainty) * scale))
        else:
            uncertainty_start = bar_len
            uncertainty_end = bar_len  # No uncertainty to display
        
        # Build the bar string
        bar_string = ''
        for i in range(bar_width):
            if i < bar_len:
                if i < uncertainty_start:
                    char = '█'  # Solid block for certain part
                else:
                    char = '█'  # Lighter block for uncertainty overlap
            else:
                if i < uncertainty_end:
                    char = '░'  # Dash for upper uncertainty range
                else:
                    char = ' '  # Space for empty area
            bar_string += char
        
        # Print the label and bar
        score_field = f'[{similarity:.4f}]'


        print(f"{field:<{label_width}} {score_field} | {bar_string}")

