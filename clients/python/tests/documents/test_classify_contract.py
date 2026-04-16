import pytest

from retab.types.documents.classify import ClassifyResponse


def test_classify_response_model_rejects_legacy_shape() -> None:
    with pytest.raises(Exception):
        ClassifyResponse.model_validate(
            {
                "result": {
                    "reasoning": "Detected invoice keywords",
                    "classification": "invoice",
                },
                "likelihood": 0.75,
                "votes": [
                    {"classification": "invoice", "reasoning": "invoice vote one"},
                    {"classification": "invoice", "reasoning": "invoice vote two"},
                    {"classification": "receipt", "reasoning": "receipt vote"},
                ],
                "usage": {"credits": 0.12},
            }
        )


def test_classify_response_model_accepts_native_shape() -> None:
    response = ClassifyResponse.model_validate(
        {
            "classification": {
                "reasoning": "Detected invoice keywords",
                "category": "invoice",
            },
            "consensus": {
                "likelihood": 0.75,
                "choices": [
                    {"classification": {"reasoning": "Detected invoice keywords", "category": "invoice"}},
                    {"classification": {"reasoning": "invoice vote one", "category": "invoice"}},
                    {"classification": {"reasoning": "invoice vote two", "category": "invoice"}},
                    {"classification": {"reasoning": "receipt vote", "category": "receipt"}},
                ],
            },
            "usage": {"credits": 0.12},
        }
    )

    assert response.classification.category == "invoice"
    assert response.consensus.likelihood == 0.75
    assert [choice.classification.category for choice in response.consensus.choices] == [
        "invoice",
        "invoice",
        "invoice",
        "receipt",
    ]
