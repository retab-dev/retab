# ---------------------------------------------
## Utility: Find and display fuzzy matches between structured records using Levenshtein distance
# ---------------------------------------------

import re
import unicodedata
from typing import Any, Dict, Hashable, List, TypedDict

import pandas as pd
from Levenshtein import distance as levenshtein_distance
from rich.console import Console
from rich.table import Table


def normalize_value(val: Any) -> str:
    """Convert a value to uppercase and remove all spacing and accents for comparison."""
    if val is None:
        return ""
    prep = re.sub(r'\s+', '', str(val).upper())
    return unicodedata.normalize('NFKD', prep).encode('ASCII', 'ignore').decode()


def levenshtein_similarity(val1: Any, val2: Any) -> float:
    """
    Returns a similarity score between 0.0 and 1.0.
    Calculate similarity between two values using the Levenshtein distance.
    """
    if (val1 or "") == (val2 or ""):
        return 1.0

    if isinstance(val1, (int, float)) and isinstance(val2, (int, float)):
        return 1.0 if abs(val1 - val2) <= 0.05 * max(abs(val1), abs(val2)) else 0.0

    str1 = normalize_value(val1)
    str2 = normalize_value(val2)

    if str1 == str2:
        return 1.0

    if str1 and str2:
        max_len = max(len(str1), len(str2))
        dist = levenshtein_distance(str1, str2)
        return 1 - (dist / max_len)

    return 0.0


class MatchResult(TypedDict):
    record: Any
    similarity: float


def find_top_k_neighbors(query: Dict[str, Any], database: List[Dict[Hashable, Any]], k: int = 5) -> List[MatchResult]:
    """Find top k closest matches in a dataset using Levenshtein similarity."""
    compare_fields = list(query.keys())
    results: List[MatchResult] = []

    for record in database:
        score_sum = 0.0
        count = 0

        for field in compare_fields:
            val1 = query.get(field, "")
            val2 = record.get(field, "")
            sim = levenshtein_similarity(val1, val2)
            score_sum += sim
            count += 1

        if count > 0:
            avg_sim = score_sum / count
            results.append({"record": record, "similarity": avg_sim})

    results.sort(key=lambda x: x["similarity"], reverse=True)
    return results[:k]


def print_results_table(results: List[MatchResult]) -> None:
    """Print results in a formatted table using Rich."""
    console = Console()
    table = Table(title=f"Top {len(results)} Matches")

    fields = set()
    for res in results:
        fields.update(res["record"].keys())
    fields = sorted(fields)

    table.add_column("Similarity", justify="right", style="cyan")
    for field in fields:
        table.add_column(field, style="magenta")

    for res in results:
        row = [f"{res['similarity']:.3f}"]
        for field in fields:
            row.append(str(res["record"].get(field, "")))
        table.add_row(*row)

    console.print(table)


def example_usage() -> None:
    """Example showing how to use this module with a CSV file."""
    df = pd.read_csv("data.csv")  # Replace with your own CSV
    database = df.to_dict('records')

    query = {
        'Code Client': 'UIFOR001',
        'Nom': 'UiForm',
        'Adresse 1': '5 Parvis Alan Turing',
        'Adresse 2': '',
        'CP': 75013,
        'Ville': 'PARIS',
        'Pays': 'FR',
        'Num TVA': 'FR00348236132',
        'SIRET': '32323219080032',
    }

    results = find_top_k_neighbors(query, database, k=5)
    print_results_table(results)


if __name__ == "__main__":
    example_usage()
