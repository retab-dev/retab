from retab.types.documents.classify import ClassifyResponse


def test_classify_response_model_coerces_legacy_shape_to_native_consensus_envelope() -> None:
    response = ClassifyResponse.model_validate(
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

    assert response.classification.category == "invoice"
    assert response.consensus.likelihood == 0.75
    assert [choice.classification.category for choice in response.consensus.choices] == [
        "invoice",
        "invoice",
        "invoice",
        "receipt",
    ]
