import datetime
import re
import unicodedata
import requests
from collections import defaultdict
from typing import Any, Literal
from Levenshtein import distance as levenshtein_distance
from io import BytesIO
from motor.motor_asyncio import AsyncIOMotorDatabase
from mimelib.email.read import read_email_data_from_eml_bytes
from pydantic import BaseModel, Field, computed_field
from typing import Any, Optional
import uuid
import termplotlib as tpl # type: ignore
import numpy as np
from typing import Literal
import pandas as pd
# The goal is to leverage this piece of code to open a jsonl file and get an analysis of the performance of the model using a one-liner. 


############# BENCHMARKING MODELS #############

class DictionaryComparisonMetrics(BaseModel):
    # Pure dict comparison
    unchanged_fields: int
    total_fields: int
    is_equal: dict[str, bool]
    similarity: dict[str, float]
    false_positives: list[dict[str, Any]]
    false_negatives: list[dict[str, Any]]
    mismatched_values: list[dict[str, Any]]
    keys_only_on_1: list[str]
    keys_only_on_2: list[str]

    # Some metrics
    avg_similarity: float
    total_similarity: float
    valid_comparisons: int
    total_accuracy: float
    false_positive_rate: float
    false_negative_rate: float
    mismatched_value_rate: float


def flatten_dict(obj: Any, prefix: str = '') -> dict[str, Any]:
    items = []  # type: ignore
    if isinstance(obj, dict):
        for k, v in obj.items():
            new_key = f"{prefix}.{k}" if prefix else k
            items.extend(flatten_dict(v, new_key).items())
    elif isinstance(obj, list):
        for i, v in enumerate(obj):
            new_key = f"{prefix}[{i}]"
            items.extend(flatten_dict(v, new_key).items())
    else:
        items.append((prefix, obj))
    return dict(items)

def normalize_value(val: Any) -> str:
    """Convert value to uppercase and remove all spacing for comparison."""
    if val is None:
        return ""
    prep = re.sub(r'\s+', '', str(val).upper())
    # Remove all accents (é -> e, etc.)
    return unicodedata.normalize('NFKD', prep).encode('ASCII', 'ignore').decode()

def key_normalization(key: str) -> str:
    """This method is useful to compare keys under list indexes (that refers to the same kind of error but on different list index position)"""
    # We will replace all [{i}] with ["_"] where i is the index of the list (using regex for this)
    return re.sub(r'\[\d+\]', "[i]", key)

def should_ignore_key(key: str, exclude_field_patterns: list[str] | None, include_field_patterns: list[str] | None = None, information_presence_per_field: dict[str, bool] | None = None) -> bool:    
    if information_presence_per_field and information_presence_per_field.get(key) is False:
        # If we have the information_presence_per_field dict and the key is marked as false, then we should ignore it
        should_ignore = True
    else:
        # If exclude_field_patterns is None, we should not ignore any key
        normalized_key = key_normalization(key)
        should_ignore = any(normalized_key.startswith(key_normalization(pattern)) for pattern in exclude_field_patterns or [])
        
        if include_field_patterns and not should_ignore:
            # If include_field_patterns is not None, we should ignore the key if it does not start with any of the include_field_patterns and is not in the exclude_field_patterns
            should_ignore = not any(normalized_key.startswith(key_normalization(pattern)) for pattern in include_field_patterns)
    
    return should_ignore

def levenshtein_similarity(val1: Any, val2: Any) -> float:
    """
    Calculate similarity between two values using Levenshtein distance.
    Returns a similarity score between 0.0 and 1.0.
    """
    # Handle None/empty and general cases
    if (val1 or "") == (val2 or ""):
        return 1.0
    
    # Check if both values are numeric, compare with 5% tolerance
    if isinstance(val1, (int, float)) and isinstance(val2, (int, float)):
        return 1.0 if abs(val1 - val2) <= 0.05 * max(abs(val1), abs(val2)) else 0.0

    # Convert to normalized strings
    str1 = normalize_value(val1)
    str2 = normalize_value(val2)
    
    if str1 == str2:
        return 1.0
        
    # Calculate Levenshtein distance
    if str1 and str2:  # Only if both strings are non-empty
        max_len = max(len(str1), len(str2))
        if max_len == 0:
            return 1.0
            
        dist = levenshtein_distance(str1, str2)
        return 1 - (dist / max_len)
        
    return 0.0


def compare_dicts(
        ground_truth: dict[str, Any],
        prediction: dict[str, Any],
        include_fields: list[str] | None = None,
        exclude_fields: list[str] | None = None,
        information_presence_per_field: dict[str, bool] | None = None,
        levenshtein_threshold: float = 0.0  # 0.0 means exact match
    ) -> DictionaryComparisonMetrics:
    flat_ground_truth = flatten_dict(ground_truth)
    flat_prediction = flatten_dict(prediction)

    flat_ground_truth = {k: v for k, v in flat_ground_truth.items() if not should_ignore_key(k, exclude_fields, include_fields, information_presence_per_field)}
    flat_prediction = {k: v for k, v in flat_prediction.items() if not should_ignore_key(k, exclude_fields, include_fields, information_presence_per_field)}
    
    keys_ground_truth = set(flat_ground_truth.keys())
    keys_prediction = set(flat_prediction.keys())
    common_keys = keys_ground_truth & keys_prediction
    
    keys_only_on_1 = sorted(list(keys_ground_truth - keys_prediction))
    keys_only_on_2 = sorted(list(keys_prediction - keys_ground_truth))
    
    total_fields = len(common_keys)
    unchanged_fields = 0
    similarity_per_field = {}
    is_equal_per_field = {}
    
    false_positives = []
    false_negatives = []
    mismatched_values = []
    
    total_similarity = 0.0
    valid_comparisons = 0

    for key in common_keys:
        llm_value = flat_ground_truth[key]
        extraction_value = flat_prediction[key]

        coerced_llm_value = llm_value or ""
        coerced_extraction_value = extraction_value or ""        
        similarity = levenshtein_similarity(llm_value, extraction_value)
        is_equal = similarity >= (1 - levenshtein_threshold)
        similarity_per_field[key] = similarity
        is_equal_per_field[key] = is_equal
        
        # Only count non-empty comparisons for average similarity
        if coerced_llm_value != "" and coerced_extraction_value != "":
            total_similarity += similarity
            valid_comparisons += 1
        
        if is_equal:
            unchanged_fields += 1
        else:
            if coerced_llm_value != "" and coerced_extraction_value == "":
                false_positives.append({
                    "key": key, 
                    "expected": extraction_value, 
                    "got": llm_value,
                    "similarity": similarity
                })
            elif coerced_llm_value == "" and coerced_extraction_value != "":
                false_negatives.append({
                    "key": key, 
                    "expected": extraction_value, 
                    "got": llm_value,
                    "similarity": similarity
                })
            elif coerced_llm_value != "" and coerced_extraction_value != "":
                # Both are non-empty but not equal
                mismatched_values.append({
                    "key": key, 
                    "expected": extraction_value, 
                    "got": llm_value,
                    "similarity": similarity
                })
    # Some metrics
    avg_similarity = total_similarity / valid_comparisons if valid_comparisons > 0 else 1.0
    total_accuracy = unchanged_fields / total_fields if total_fields > 0 else 1.0
    false_positive_rate = len(false_positives) / total_fields if total_fields > 0 else 0.0
    false_negative_rate = len(false_negatives) / total_fields if total_fields > 0 else 0.0
    mismatched_value_rate = len(mismatched_values) / total_fields if total_fields > 0 else 0.0

    return DictionaryComparisonMetrics(
        unchanged_fields=unchanged_fields,
        total_fields=total_fields,
        is_equal=is_equal_per_field,
        similarity=similarity_per_field,
        false_positives=false_positives,
        false_negatives=false_negatives,
        mismatched_values = mismatched_values,
        keys_only_on_1=keys_only_on_1,
        keys_only_on_2=keys_only_on_2,
        avg_similarity=avg_similarity,
        total_similarity=total_similarity,
        valid_comparisons=valid_comparisons,
        total_accuracy=total_accuracy,
        false_positive_rate=false_positive_rate,
        false_negative_rate=false_negative_rate,
        mismatched_value_rate=mismatched_value_rate
    )


class ExtractionAnalysis(BaseModel):
    ground_truth: dict[str, Any]
    prediction: dict[str, Any]
    time_spent: Optional[float] = None
    include_fields: list[str] | None = None
    exclude_fields: list[str] | None = None
    information_presence_per_field: dict[str, bool] | None = None
    levenshtein_threshold: float = 0.0

    @computed_field # type: ignore
    @property
    def comparison(self) -> DictionaryComparisonMetrics:
        return compare_dicts(
            self.ground_truth,
            self.prediction,
            include_fields=self.include_fields,
            exclude_fields=self.exclude_fields,
            information_presence_per_field=self.information_presence_per_field,
            levenshtein_threshold=self.levenshtein_threshold
        )

class ComparisonMetrics(BaseModel):
    # Total Values (count or sum) per Field
    false_positive_counts: dict[str, int] = defaultdict(int)
    false_negative_counts: dict[str, int] = defaultdict(int)
    mismatched_value_counts: dict[str, int] = defaultdict(int)
    common_presence_counts: dict[str, int] = defaultdict(int)
    total_similarity_per_field: dict[str, float] = defaultdict(float)

    # Rates (percentage) per Field
    accuracy_per_field: dict[str, float] = defaultdict(float)
    similarity_per_field: dict[str, float] = defaultdict(float)
    false_positive_rate_per_field: dict[str, float] = defaultdict(float)
    false_negative_rate_per_field: dict[str, float] = defaultdict(float)
    mismatched_value_rate_per_field: dict[str, float] = defaultdict(float)

    @computed_field
    def accuracy(self) -> float:
        return sum(self.accuracy_per_field.values()) / len(self.accuracy_per_field)
    
    @computed_field
    def similarity(self) -> float:
        return sum(self.similarity_per_field.values()) / len(self.similarity_per_field)
    
    @computed_field
    def false_positive_rate(self) -> float:
        return sum(self.false_positive_rate_per_field.values()) / len(self.false_positive_rate_per_field)
    
    @computed_field
    def false_negative_rate(self) -> float:
        return sum(self.false_negative_rate_per_field.values()) / len(self.false_negative_rate_per_field)
    
    @computed_field
    def mismatched_value_rate(self) -> float:
        return sum(self.mismatched_value_rate_per_field.values()) / len(self.mismatched_value_rate_per_field)

def analyze_comparison_metrics(list_analyses: list[ExtractionAnalysis], min_freq: float = 0.2) -> ComparisonMetrics:
    false_positive_counts: dict[str, int] = defaultdict(int)
    false_negative_counts: dict[str, int] = defaultdict(int)
    mismatched_value_counts: dict[str, int] = defaultdict(int)
    common_presence_counts: dict[str, int] = defaultdict(int)
    is_equal_per_field: dict[str, int] = defaultdict(int)

    total_similarity_per_field: dict[str, float] = defaultdict(float)
    false_positive_rate_per_field: dict[str, float] = defaultdict(float)
    false_negative_rate_per_field: dict[str, float] = defaultdict(float)
    mismatched_value_rate_per_field: dict[str, float] = defaultdict(float)


    for analysis in list_analyses:
        # Count false positives
        for error in analysis.comparison.false_positives:
            key = error["key"]
            false_positive_counts[key_normalization(key)] += 1
            
        # Count false negatives
        for error in analysis.comparison.false_negatives:
            key = error["key"]
            false_negative_counts[key_normalization(key)] += 1
        
        # Count Wrong Predictions
        for error in analysis.comparison.mismatched_values:
            key = error["key"]
            mismatched_value_counts[key_normalization(key)] += 1

        # Count total errors per field
        for key, similarity in analysis.comparison.similarity.items():
            common_presence_counts[key_normalization(key)] += 1
            total_similarity_per_field[key_normalization(key)] += similarity
        for key, is_equal in analysis.comparison.is_equal.items():
            is_equal_per_field[key_normalization(key)] += int(is_equal)

    accuracy_per_field = {key: is_equal_per_field[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))}
    similarity_per_field = {key: total_similarity_per_field[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))}
    false_positive_rate_per_field = {key: false_positive_counts[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))}
    false_negative_rate_per_field = {key: false_negative_counts[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))}
    mismatched_value_rate_per_field = {key: mismatched_value_counts[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))}
    return ComparisonMetrics(
        false_positive_counts=false_positive_counts,
        false_negative_counts=false_negative_counts,
        mismatched_value_counts=mismatched_value_counts,
        common_presence_counts=common_presence_counts,
        total_similarity_per_field=total_similarity_per_field,
        accuracy_per_field=accuracy_per_field,
        similarity_per_field=similarity_per_field,
        false_positive_rate_per_field=false_positive_rate_per_field,
        false_negative_rate_per_field=false_negative_rate_per_field,
        mismatched_value_rate_per_field=mismatched_value_rate_per_field
    )




def plot_metric(analysis: ComparisonMetrics, value_type: Literal["accuracy", "similarity", "false_positive_rate", "false_negative_rate", "mismatched_value_rate"] = "accuracy", top_n: int = 20, ascending: bool = False) -> None:
    """Plot a metric from analysis results using a horizontal bar chart.
    
    Args:
        analysis: ComparisonMetrics object containing the analysis results
    """
    # Create dataframe from accuracy data
    df = pd.DataFrame(
        list(analysis.__getattribute__(value_type + "_per_field").items()), 
        columns=["field", value_type]
    ).sort_values(by=value_type, ascending=ascending)

    # Filter top n fields with the lowest accuracy
    top_n_df = df.head(top_n)

    # Create the plot
    fig = tpl.figure()
    fig.barh(
        np.array(top_n_df[value_type]).round(4), 
        np.array(top_n_df["field"]), 
        force_ascii=False
    )

    fig.show()

def plot_comparison_metrics(analysis: ComparisonMetrics, top_n: int = 20)-> None:
    metric_ascendency_dict: dict[Literal["accuracy", "similarity", "false_positive_rate", "false_negative_rate", "mismatched_value_rate"], bool] = {
        "accuracy": True,
        "similarity": True,
        "false_positive_rate": False,
        "false_negative_rate": False,
        "mismatched_value_rate": False
    }
    
    
    for metric, ascending in metric_ascendency_dict.items():
        print(f"\n############ {metric.upper()} ############")
        plot_metric(analysis, metric, top_n, ascending)