import re
import unicodedata
from collections import defaultdict
from typing import Any, Literal, Optional

import numpy as np
import pandas as pd
import termplotlib as tpl  # type: ignore
from Levenshtein import distance as levenshtein_distance
from pydantic import BaseModel, computed_field

# The goal is to leverage this piece of code to open a jsonl file and get an analysis of the performance of the model using a one-liner.


############# BENCHMARKING MODELS #############


class DictionaryComparisonMetrics(BaseModel):
    # Pure dict comparison
    unchanged_fields: int
    total_fields: int
    is_equal: dict[str, bool]
    false_positives: list[dict[str, Any]]
    false_negatives: list[dict[str, Any]]
    mismatched_values: list[dict[str, Any]]
    keys_only_on_1: list[str]
    keys_only_on_2: list[str]

    # Some metrics
    valid_comparisons: int
    total_accuracy: float
    false_positive_rate: float
    false_negative_rate: float
    mismatched_value_rate: float

    similarity_levenshtein: dict[str, float]
    similarity_jaccard: dict[str, float]

    avg_similarity_levenshtein: float
    avg_similarity_jaccard: float
    total_similarity_levenshtein: float
    total_similarity_jaccard: float


def flatten_dict(obj: Any, prefix: str = '') -> dict[str, Any]:
    items = []  # type: ignore
    if isinstance(obj, dict):
        for k, v in obj.items():
            new_key = f"{prefix}.{k}" if prefix else k
            items.extend(flatten_dict(v, new_key).items())
    elif isinstance(obj, list):
        for i, v in enumerate(obj):
            new_key = f"{prefix}.{i}"
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
    # We will replace all .{i} with .* where i is the index of the list (using regex for this)
    key_parts = key.split(".")
    new_key_parts = []
    for key_part in key_parts:
        if key_part.isdigit():
            new_key_parts.append("*")
        else:
            new_key_parts.append(key_part)
    return ".".join(new_key_parts)


def should_ignore_key(
    key: str, exclude_field_patterns: list[str] | None, include_field_patterns: list[str] | None = None, information_presence_per_field: dict[str, bool] | None = None
) -> bool:
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


def jaccard_similarity(val1: Any, val2: Any) -> float:
    """
    Calculate Jaccard similarity between two values.
    Returns a similarity score between 0.0 and 1.0.
    """
    # Handle None/empty and general cases
    if (val1 or "") == (val2 or ""):
        return 1.0

    # Check if both values are numeric, compare with 5% tolerance
    if isinstance(val1, (int, float)) and isinstance(val2, (int, float)):
        return 1.0 if abs(val1 - val2) <= 0.05 * max(abs(val1), abs(val2)) else 0.0

    # Convert to normalized strings and split into words
    str1 = set(normalize_value(val1).split())
    str2 = set(normalize_value(val2).split())

    if not str1 and not str2:
        return 1.0

    # Calculate Jaccard similarity
    intersection = len(str1.intersection(str2))
    union = len(str1.union(str2))

    return intersection / union if union > 0 else 0.0


def compare_dicts(
    ground_truth: dict[str, Any],
    prediction: dict[str, Any],
    include_fields: list[str] | None = None,
    exclude_fields: list[str] | None = None,
    information_presence_per_field: dict[str, bool] | None = None,
    levenshtein_threshold: float = 0.0,  # 0.0 means exact match
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
    is_equal_per_field = {}

    false_positives = []
    false_negatives = []
    mismatched_values = []

    total_similarity_levenshtein = 0.0
    total_similarity_jaccard = 0.0
    similarity_levenshtein_per_field = {}
    similarity_jaccard_per_field = {}

    valid_comparisons = 0

    for key in common_keys:
        llm_value = flat_ground_truth[key]
        extraction_value = flat_prediction[key]

        coerced_llm_value = llm_value or ""
        coerced_extraction_value = extraction_value or ""

        similarity_lev = levenshtein_similarity(llm_value, extraction_value)
        similarity_jac = jaccard_similarity(llm_value, extraction_value)
        # print("Jaccard similarity", similarity_jac)

        # Use Levenshtein for equality comparison (you can adjust this if needed)
        is_equal = similarity_lev >= (1 - levenshtein_threshold)

        similarity_levenshtein_per_field[key] = similarity_lev
        similarity_jaccard_per_field[key] = similarity_jac
        is_equal_per_field[key] = is_equal

        # Only count non-empty comparisons for average similarity
        if coerced_llm_value != "" and coerced_extraction_value != "":
            total_similarity_levenshtein += similarity_lev
            total_similarity_jaccard += similarity_jac
            valid_comparisons += 1

        if is_equal:
            unchanged_fields += 1
        else:
            if coerced_llm_value != "" and coerced_extraction_value == "":
                false_positives.append({"key": key, "expected": extraction_value, "got": llm_value, "similarity": similarity_lev})
            elif coerced_llm_value == "" and coerced_extraction_value != "":
                false_negatives.append({"key": key, "expected": extraction_value, "got": llm_value, "similarity": similarity_lev})
            elif coerced_llm_value != "" and coerced_extraction_value != "":
                # Both are non-empty but not equal
                mismatched_values.append({"key": key, "expected": extraction_value, "got": llm_value, "similarity": similarity_lev})
    # Some metrics
    avg_similarity_levenshtein = total_similarity_levenshtein / valid_comparisons if valid_comparisons > 0 else 1.0
    avg_similarity_jaccard = total_similarity_jaccard / valid_comparisons if valid_comparisons > 0 else 1.0
    total_accuracy = unchanged_fields / total_fields if total_fields > 0 else 1.0
    false_positive_rate = len(false_positives) / total_fields if total_fields > 0 else 0.0
    false_negative_rate = len(false_negatives) / total_fields if total_fields > 0 else 0.0
    mismatched_value_rate = len(mismatched_values) / total_fields if total_fields > 0 else 0.0

    return DictionaryComparisonMetrics(
        unchanged_fields=unchanged_fields,
        total_fields=total_fields,
        is_equal=is_equal_per_field,
        false_positives=false_positives,
        false_negatives=false_negatives,
        mismatched_values=mismatched_values,
        keys_only_on_1=keys_only_on_1,
        keys_only_on_2=keys_only_on_2,
        valid_comparisons=valid_comparisons,
        total_accuracy=total_accuracy,
        false_positive_rate=false_positive_rate,
        false_negative_rate=false_negative_rate,
        mismatched_value_rate=mismatched_value_rate,
        similarity_levenshtein=similarity_levenshtein_per_field,
        similarity_jaccard=similarity_jaccard_per_field,
        avg_similarity_levenshtein=avg_similarity_levenshtein,
        avg_similarity_jaccard=avg_similarity_jaccard,
        total_similarity_levenshtein=total_similarity_levenshtein,
        total_similarity_jaccard=total_similarity_jaccard,
    )


class ExtractionAnalysis(BaseModel):
    ground_truth: dict[str, Any]
    prediction: dict[str, Any]
    time_spent: Optional[float] = None
    include_fields: list[str] | None = None
    exclude_fields: list[str] | None = None
    information_presence_per_field: dict[str, bool] | None = None
    levenshtein_threshold: float = 0.0

    @computed_field  # type: ignore
    @property
    def comparison(self) -> DictionaryComparisonMetrics:
        return compare_dicts(
            self.ground_truth,
            self.prediction,
            include_fields=self.include_fields,
            exclude_fields=self.exclude_fields,
            information_presence_per_field=self.information_presence_per_field,
            levenshtein_threshold=self.levenshtein_threshold,
        )


class BenchmarkMetrics(BaseModel):
    ai_model: str
    accuracy: float
    levenshtein_similarity: float
    jaccard_similarity: float
    false_positive_rate: float
    false_negative_rate: float
    mismatched_value_rate: float


from rich.console import Console
from rich.table import Table


def display_benchmark_metrics(benchmark_metrics: list[BenchmarkMetrics]) -> None:
    """
    Display benchmark metrics for multiple models in a formatted table.

    Args:
        benchmark_metrics: List of BenchmarkMetrics objects containing model performance data
    """
    console = Console(style="on grey23")
    table = Table(title="Model Benchmark Comparison", show_lines=True)

    # Add columns
    table.add_column("Model", justify="left", style="#BDE8F6", no_wrap=True)
    table.add_column("Accuracy", justify="right", style="#C2BDF6")
    table.add_column("Levenshtein", justify="right", style="#F6BDBD")
    table.add_column("Jaccard", justify="right", style="#F6E4BD")
    table.add_column("False Positive Rate", justify="right", style="#BDF6C0")
    table.add_column("False Negative Rate", justify="right", style="#F6BDE4")
    table.add_column("Mismatched Value Rate", justify="right", style="#E4F6BD")

    # Find best values for each metric
    best_values = {
        'accuracy': max(m.accuracy for m in benchmark_metrics),
        'levenshtein': max(m.levenshtein_similarity for m in benchmark_metrics),
        'jaccard': max(m.jaccard_similarity for m in benchmark_metrics),
        'fp_rate': min(m.false_positive_rate for m in benchmark_metrics),
        'fn_rate': min(m.false_negative_rate for m in benchmark_metrics),
        'mismatch_rate': min(m.mismatched_value_rate for m in benchmark_metrics),
    }

    # Add rows for each model's metrics
    for metrics in benchmark_metrics:
        table.add_row(
            metrics.ai_models,
            f"[bold]{metrics.accuracy:.3f}[/bold]" if metrics.accuracy == best_values['accuracy'] else f"[dim]{metrics.accuracy:.3f}[/dim]",
            f"[bold]{metrics.levenshtein_similarity:.3f}[/bold]"
            if metrics.levenshtein_similarity == best_values['levenshtein']
            else f"[dim]{metrics.levenshtein_similarity:.3f}[/dim]",
            f"[bold]{metrics.jaccard_similarity:.3f}[/bold]" if metrics.jaccard_similarity == best_values['jaccard'] else f"[dim]{metrics.jaccard_similarity:.3f}[/dim]",
            f"[bold]{metrics.false_positive_rate:.3f}[/bold]" if metrics.false_positive_rate == best_values['fp_rate'] else f"[dim]{metrics.false_positive_rate:.3f}[/dim]",
            f"[bold]{metrics.false_negative_rate:.3f}[/bold]" if metrics.false_negative_rate == best_values['fn_rate'] else f"[dim]{metrics.false_negative_rate:.3f}[/dim]",
            f"[bold]{metrics.mismatched_value_rate:.3f}[/bold]"
            if metrics.mismatched_value_rate == best_values['mismatch_rate']
            else f"[dim]{metrics.mismatched_value_rate:.3f}[/dim]",
        )

    # Print the table
    console.print(table)


class ComparisonMetrics(BaseModel):
    # Total Values (count or sum) per Field
    false_positive_counts: dict[str, int] = defaultdict(int)
    false_positive_rate_per_field: dict[str, float] = defaultdict(float)

    false_negative_counts: dict[str, int] = defaultdict(int)
    false_negative_rate_per_field: dict[str, float] = defaultdict(float)

    mismatched_value_counts: dict[str, int] = defaultdict(int)
    mismatched_value_rate_per_field: dict[str, float] = defaultdict(float)

    common_presence_counts: dict[str, int] = defaultdict(int)
    accuracy_per_field: dict[str, float] = defaultdict(float)

    jaccard_similarity_per_field: dict[str, float] = defaultdict(float)
    total_jaccard_similarity_per_field: dict[str, float] = defaultdict(float)

    levenshtein_similarity_per_field: dict[str, float] = defaultdict(float)
    total_levenshtein_similarity_per_field: dict[str, float] = defaultdict(float)

    @computed_field  # type: ignore
    @property
    def accuracy(self) -> float:
        return sum(self.accuracy_per_field.values()) / len(self.accuracy_per_field)

    @computed_field  # type: ignore
    @property
    def levenshtein_similarity(self) -> float:
        return sum(self.levenshtein_similarity_per_field.values()) / len(self.levenshtein_similarity_per_field)

    @computed_field  # type: ignore
    @property
    def jaccard_similarity(self) -> float:
        return sum(self.jaccard_similarity_per_field.values()) / len(self.jaccard_similarity_per_field)

    @computed_field  # type: ignore
    @property
    def false_positive_rate(self) -> float:
        return sum(self.false_positive_rate_per_field.values()) / len(self.false_positive_rate_per_field)

    @computed_field  # type: ignore
    @property
    def false_negative_rate(self) -> float:
        return sum(self.false_negative_rate_per_field.values()) / len(self.false_negative_rate_per_field)

    @computed_field  # type: ignore
    @property
    def mismatched_value_rate(self) -> float:
        return sum(self.mismatched_value_rate_per_field.values()) / len(self.mismatched_value_rate_per_field)


def normalized_comparison_metrics(list_analyses: list[ExtractionAnalysis], min_freq: float = 0.2) -> ComparisonMetrics:
    false_positive_counts: dict[str, int] = defaultdict(int)
    false_negative_counts: dict[str, int] = defaultdict(int)
    mismatched_value_counts: dict[str, int] = defaultdict(int)
    common_presence_counts: dict[str, int] = defaultdict(int)
    is_equal_per_field: dict[str, int] = defaultdict(int)

    total_levenshtein_similarity_per_field: dict[str, float] = defaultdict(float)
    total_jaccard_similarity_per_field: dict[str, float] = defaultdict(float)
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

        # Count total errors per field (Levenshtein)
        for key, similarity in analysis.comparison.similarity_levenshtein.items():
            common_presence_counts[key_normalization(key)] += 1
            total_levenshtein_similarity_per_field[key_normalization(key)] += similarity

        for key, is_equal in analysis.comparison.is_equal.items():
            is_equal_per_field[key_normalization(key)] += int(is_equal)

        # Count Jaccard Similarity
        for key, similarity in analysis.comparison.similarity_jaccard.items():
            total_jaccard_similarity_per_field[key_normalization(key)] += similarity

    accuracy_per_field = {
        key: is_equal_per_field[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))
    }
    levenshtein_similarity_per_field = {
        key: total_levenshtein_similarity_per_field[key] / common_presence_counts[key]
        for key in common_presence_counts
        if common_presence_counts[key] > int(min_freq * len(list_analyses))
    }
    jaccard_similarity_per_field = {
        key: total_jaccard_similarity_per_field[key] / common_presence_counts[key]
        for key in common_presence_counts
        if common_presence_counts[key] > int(min_freq * len(list_analyses))
    }
    false_positive_rate_per_field = {
        key: false_positive_counts[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))
    }
    false_negative_rate_per_field = {
        key: false_negative_counts[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))
    }
    mismatched_value_rate_per_field = {
        key: mismatched_value_counts[key] / common_presence_counts[key] for key in common_presence_counts if common_presence_counts[key] > int(min_freq * len(list_analyses))
    }

    return ComparisonMetrics(
        false_positive_counts=false_positive_counts,
        false_negative_counts=false_negative_counts,
        mismatched_value_counts=mismatched_value_counts,
        common_presence_counts=common_presence_counts,
        total_levenshtein_similarity_per_field=total_levenshtein_similarity_per_field,
        total_jaccard_similarity_per_field=total_jaccard_similarity_per_field,
        accuracy_per_field=accuracy_per_field,
        levenshtein_similarity_per_field=levenshtein_similarity_per_field,
        jaccard_similarity_per_field=jaccard_similarity_per_field,
        false_positive_rate_per_field=false_positive_rate_per_field,
        false_negative_rate_per_field=false_negative_rate_per_field,
        mismatched_value_rate_per_field=mismatched_value_rate_per_field,
    )


def plot_metric(
    analysis: ComparisonMetrics,
    value_type: Literal["accuracy", "levenshtein_similarity", "jaccard_similarity", "false_positive_rate", "false_negative_rate", "mismatched_value_rate"] = "accuracy",
    top_n: int = 20,
    ascending: bool = False,
) -> None:
    """Plot a metric from analysis results using a horizontal bar chart.

    Args:
        analysis: ComparisonMetrics object containing the analysis results
    """
    # Create dataframe from accuracy data
    df = pd.DataFrame(list(analysis.__getattribute__(value_type + "_per_field").items()), columns=["field", value_type]).sort_values(by=value_type, ascending=ascending)

    # Filter top n fields with the lowest accuracy
    top_n_df = df.head(top_n)

    # Create the plot
    fig = tpl.figure()
    fig.barh(np.array(top_n_df[value_type]).round(4), np.array(top_n_df["field"]), force_ascii=False)

    fig.show()


def plot_comparison_metrics(analysis: ComparisonMetrics, top_n: int = 20) -> None:
    metric_ascendency_dict: dict[
        Literal["accuracy", "levenshtein_similarity", "jaccard_similarity", "false_positive_rate", "false_negative_rate", "mismatched_value_rate"], bool
    ] = {"accuracy": True, "levenshtein_similarity": True, "jaccard_similarity": True, "false_positive_rate": False, "false_negative_rate": False, "mismatched_value_rate": False}

    print(f"#########################################")
    print(f"############ AVERAGE METRICS ############")
    print(f"#########################################")
    print(f"Accuracy: {analysis.accuracy:.2f}")
    print(f"Levenshtein Similarity: {analysis.levenshtein_similarity:.2f}")
    print(f"Jaccard Similarity (IOU): {analysis.jaccard_similarity:.2f}")
    print(f"False Positive Rate: {analysis.false_positive_rate:.2f}")
    print(f"False Negative Rate: {analysis.false_negative_rate:.2f}")
    print(f"Mismatched Value Rate: {analysis.mismatched_value_rate:.2f}")

    for metric, ascending in metric_ascendency_dict.items():
        print(f"\n\n############ {metric.upper()} ############")
        plot_metric(analysis, metric, top_n, ascending)


def get_aggregation_metrics(metric: dict[str, float | int], _hierarchy_level: int) -> dict[str, float]:
    if _hierarchy_level == 0:
        # For level 0, aggregate all values under empty string key
        return {"": sum(metric.values()) / len(metric)}

    aggregated_metrics: dict[str, list[float | int]] = {}
    for key, value in metric.items():
        # Split key and handle array notation by replacing array indices with '*'
        key_parts: list[str] = []
        for part in key.split('.'):
            if part.isdigit():
                key_parts.append('*')
            else:
                key_parts.append(part)
        actual_depth = len([part for part in key_parts if part != '*'])
        if actual_depth < _hierarchy_level:
            continue

        used_depth = 0
        aggregation_prefix_parts: list[str] = []
        for part in key_parts:
            if part == "*":
                aggregation_prefix_parts.append("*")
            else:
                aggregation_prefix_parts.append(part)
                used_depth += 1
            if used_depth == _hierarchy_level:
                break
        aggregation_prefix = '.'.join(aggregation_prefix_parts)

        # Aggregate metrics
        if aggregation_prefix not in aggregated_metrics:
            aggregated_metrics[aggregation_prefix] = []
        aggregated_metrics[aggregation_prefix].append(value)

    # Calculate averages
    return {key: sum(values) / len(values) if len(values) > 0 else 0 for key, values in aggregated_metrics.items()}


def get_max_depth(metric: dict[str, float | int]) -> int:
    max_depth_upper_bound = max(len(key.split('.')) for key in metric.keys())
    for max_depth in range(max_depth_upper_bound, 0, -1):
        if get_aggregation_metrics(metric, max_depth):
            return max_depth
    return 0


def aggregate_metric_per_hierarchy_level(metric: dict[str, float | int], hierarchy_level: int) -> dict[str, float | int]:
    """Aggregates metrics by grouping and averaging values at a specified hierarchy level in the key structure.

    Args:
        metric: Dictionary mapping hierarchical keys (dot-separated strings) to numeric values.
               Array indices in keys are treated as wildcards.
        hierarchy_level: The depth level at which to aggregate the metrics.
                        E.g. level 1 aggregates at first dot separator.

    Returns:
        Dictionary mapping aggregated keys to averaged values. Keys are truncated to the specified
        hierarchy level with array indices replaced by '*'.

    Raises:
        ValueError: If the requested hierarchy level exceeds the maximum depth in the data.
    """
    max_depth = get_max_depth(metric)

    if hierarchy_level > max_depth:
        raise ValueError(f"Hierarchy level {hierarchy_level} is greater than the maximum depth {max_depth}")

    return get_aggregation_metrics(metric, hierarchy_level)
