from __future__ import annotations

from typing import Any


def _normalize_classify_decision_payload(value: Any) -> dict[str, Any] | None:
    if not isinstance(value, dict):
        return None

    normalized = dict(value)
    if "category" not in normalized:
        legacy_category = normalized.get("classification")
        if legacy_category is not None:
            normalized["category"] = legacy_category

    reasoning = normalized.get("reasoning")
    category = normalized.get("category")
    if not isinstance(reasoning, str) or not isinstance(category, str):
        return None

    return {
        "reasoning": reasoning,
        "category": category,
    }


def _normalize_classify_choice_payload(value: Any) -> dict[str, Any] | None:
    decision = _normalize_classify_decision_payload(value)
    if decision is not None:
        return decision

    if not isinstance(value, dict):
        return None
    return _normalize_classify_decision_payload(value.get("classification"))


def normalize_classify_response_payload(value: Any) -> Any:
    if not isinstance(value, dict):
        return value

    raw_classification = value.get("classification")
    if not isinstance(raw_classification, dict):
        raw_classification = value.get("result")
    classification = _normalize_classify_decision_payload(raw_classification)
    if classification is None:
        return value

    raw_consensus = value.get("consensus")
    raw_choices = raw_consensus.get("choices") if isinstance(raw_consensus, dict) else None
    choices: list[dict[str, Any]] = []

    if isinstance(raw_choices, list):
        for raw_choice in raw_choices:
            normalized_choice = _normalize_classify_choice_payload(raw_choice)
            if normalized_choice is not None:
                choices.append(normalized_choice)
    else:
        raw_likelihood = value.get("likelihood")
        raw_votes = value.get("votes")
        if isinstance(raw_likelihood, (int, float)) or isinstance(raw_votes, list):
            choices.append(dict(classification))
        if isinstance(raw_votes, list):
            for raw_vote in raw_votes:
                normalized_vote = _normalize_classify_decision_payload(raw_vote)
                if normalized_vote is not None:
                    choices.append(normalized_vote)

    likelihood = None
    if isinstance(raw_consensus, dict):
        raw_likelihood = raw_consensus.get("likelihood")
        if isinstance(raw_likelihood, (int, float)):
            likelihood = float(raw_likelihood)
    else:
        raw_likelihood = value.get("likelihood")
        if isinstance(raw_likelihood, (int, float)):
            likelihood = float(raw_likelihood)

    return {
        "classification": classification,
        "consensus": {
            "choices": choices,
            "likelihood": likelihood,
        },
        "usage": value.get("usage", {}),
    }
