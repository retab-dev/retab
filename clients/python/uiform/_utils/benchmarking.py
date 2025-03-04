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
    Compute the difference between two dictionaries recursively.

    Args:
        dict1: The first dictionary (can be nested)
        dict2: The second dictionary (can be nested)
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
    else:
        raise ValueError(f"Invalid metric: {metric}")

    def compare_values(val1: Any, val2: Any, path: str = "") -> Any:
        # If both are dictionaries, process recursively
        if isinstance(val1, dict) and isinstance(val2, dict):
            nested_result: dict[str, Any] = {}
            all_keys = set(val1.keys()) | set(val2.keys())

            for key in all_keys:
                norm_key = key_normalization(key)
                sub_val1 = val1.get(key, None)
                sub_val2 = val2.get(key, None)

                if sub_val1 is None or sub_val2 is None:
                    nested_result[norm_key] = None
                else:
                    nested_result[norm_key] = compare_values(sub_val1, sub_val2, f"{path}.{norm_key}" if path else norm_key)

            return nested_result

        # If both are lists/arrays, compare items with detailed results
        if isinstance(val1, (list, tuple)) and isinstance(val2, (list, tuple)):
            # If both lists are empty, they're identical
            if not val1 and not val2:
                return 1.0

            # Create a detailed element-by-element comparison
            array_result = {}
            similarities = []

            # Process each position in both arrays
            for i, (item1, item2) in enumerate(zip_longest(val1, val2, fillvalue=None)):
                element_key = str(i)  # Use index as dictionary key
                element_path = f"{path}.{i}" if path else str(i)

                if item1 is None or item2 is None:
                    # Handle lists of different lengths
                    array_result[element_key] = None
                    similarities.append(0.0)  # Penalize missing elements
                else:
                    # Compare the elements
                    comparison_result = compare_values(item1, item2, element_path)
                    array_result[element_key] = comparison_result

                    # Extract similarity metric for this element
                    if isinstance(comparison_result, dict):
                        # Calculate average from nested structure
                        numeric_values = [v for v in _extract_numeric_values(comparison_result) if v is not None]
                        if numeric_values:
                            similarities.append(sum(numeric_values) / len(numeric_values))
                    elif isinstance(comparison_result, (int, float)) and comparison_result is not None:
                        similarities.append(float(comparison_result))

            # Add overall similarity as a special key
            array_result["_similarity"] = sum(similarities) / max(len(similarities), 1) if similarities else 1.0

            return array_result

        # If one is a dict and the other isn't, return None
        if isinstance(val1, dict) or isinstance(val2, dict) or isinstance(val1, (list, tuple)) or isinstance(val2, (list, tuple)):
            return None

        # Handle leaf nodes (primitives) by converting to strings and comparing
        str_val1 = "" if val1 is None else str(val1)
        str_val2 = "" if val2 is None else str(val2)
        return float(metric_function(str_val1, str_val2))  # Ensure we return a float

    def _extract_numeric_values(d: dict) -> list[float]:
        """Extract all numeric values from a nested dictionary."""
        result = []
        for k, v in d.items():
            if isinstance(v, dict):
                # Recursively extract from nested dictionaries
                result.extend(_extract_numeric_values(v))
            elif isinstance(v, (int, float)) and not isinstance(v, bool):
                # Add numeric values
                result.append(v)
            # Skip non-numeric values
        return result

    # Normalize top-level keys
    dict1_normalized = {key_normalization(k): v for k, v in dict1.items()}
    dict2_normalized = {key_normalization(k): v for k, v in dict2.items()}

    # Process all keys from both dictionaries
    all_keys = set(dict1_normalized.keys()) | set(dict2_normalized.keys())

    for key in all_keys:
        val1 = dict1_normalized.get(key, None)
        val2 = dict2_normalized.get(key, None)

        if val1 is None or val2 is None:
            result[key] = None
        else:
            result[key] = compare_values(val1, val2, key)

    return result


def aggregate_dict_differences(dict_differences: list[dict[str, Any]]) -> tuple[dict[str, Any], dict[str, Any]]:
    """
    Aggregate a list of dictionary differences into a single dictionary with average values,
    handling nested dictionaries recursively.
    
    Args:
        dict_differences: A list of dictionaries containing similarity metrics (can be nested)
        
    Returns:
        A tuple containing:
        - A dictionary with the average similarity metrics across all input dictionaries
        - A dictionary with the uncertainty (standard deviation) for each metric
    """
    if not dict_differences:
        return {}, {}
    
    def aggregate_recursively(dicts_list: list[dict[str, Any]]) -> tuple[dict[str, Any], dict[str, Any]]:
        # Initialize result dictionaries
        result: dict[str, Any] = {}
        uncertainty: dict[str, Any] = {}
        
        # Collect all keys across all dictionaries
        all_keys: set[str] = set()
        for d in dicts_list:
            all_keys.update(d.keys())
        
        for key in all_keys:
            # Collect values for this key from all dictionaries
            values = []
            for d in dicts_list:
                if key in d:
                    values.append(d[key])
            
            # Skip if no valid values
            if not values:
                result[key] = None
                uncertainty[key] = None
                continue
                
            # Check if values are nested dictionaries
            if all(isinstance(v, dict) for v in values if v is not None):
                # Filter out None values
                nested_dicts = [v for v in values if v is not None]
                if nested_dicts:
                    nested_result, nested_uncertainty = aggregate_recursively(nested_dicts)
                    result[key] = nested_result
                    uncertainty[key] = nested_uncertainty
                else:
                    result[key] = None
                    uncertainty[key] = None
            else:
                # Handle leaf nodes (numeric values)
                numeric_values = [v for v in values if v is not None and isinstance(v, (int, float))]
                
                if numeric_values:
                    mean = sum(numeric_values) / len(numeric_values)
                    result[key] = mean
                    
                    if len(numeric_values) > 1:
                        variance = sum((x - mean) ** 2 for x in numeric_values) / (len(numeric_values) - 1)
                        uncertainty[key] = max(0, variance) ** 0.5
                    else:
                        uncertainty[key] = 0.0
                else:
                    result[key] = None
                    uncertainty[key] = None
        
        return result, uncertainty
    
    return aggregate_recursively(dict_differences)




class SingleFileEval(BaseModel):
    """
    A class for evaluating metrics between two dictionaries.
    """
    file_id: str
    dataset_membership_id_1: str
    dataset_membership_id_2: str
    schema_id: str
    schema_data_id: str
    dict_1: dict[str, Any]
    dict_2: dict[str, Any]

    hamming_similarity: dict[str, Any]
    jaccard_similarity: dict[str, Any]
    levenshtein_similarity: dict[str, Any]


class EvalMetric(BaseModel):
    average: dict[str, Any]
    std: dict[str, Any]

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
    analysis: dict[str, Any],
    uncertainties: Optional[dict[str, Any]] = None,
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

