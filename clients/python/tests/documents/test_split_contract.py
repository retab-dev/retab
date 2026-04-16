from retab.types.documents.split import SplitResponse


def test_split_response_model_accepts_native_consensus_envelope() -> None:
    response = SplitResponse.model_validate(
        {
            "splits": [
                {
                    "name": "properties",
                    "metric_key": "properties",
                    "display_label": "properties",
                    "instance_index": 1,
                    "pages": [1, 2, 3],
                    "page_start": 1,
                    "page_end": 3,
                    "partitions": [
                        {
                            "key": "10*310146",
                            "pages": [1],
                        }
                    ],
                }
            ],
            "consensus": {
                "likelihoods": {
                    "splits": [
                        {
                            "likelihood": 0.9703,
                            "name": 0.9703,
                            "pages": [1.0, 1.0, 1.0],
                            "partitions": [
                                {
                                    "likelihood": 1.0,
                                    "key": 1.0,
                                    "pages": [1.0],
                                }
                            ],
                        }
                    ]
                },
                "choices": [
                    {
                        "splits": [
                            {
                                "name": "properties",
                                "metric_key": "properties",
                                "display_label": "properties",
                                "instance_index": 1,
                                "pages": [1, 2, 3],
                                "page_start": 1,
                                "page_end": 3,
                                "partitions": [{"key": "10*310146", "pages": [1]}],
                            }
                        ]
                    },
                    {
                        "splits": [
                            {
                                "name": "properties",
                                "metric_key": "properties",
                                "display_label": "properties",
                                "instance_index": 1,
                                "pages": [1, 2, 3],
                                "page_start": 1,
                                "page_end": 3,
                                "partitions": [{"key": "10*310146", "pages": [1]}],
                            }
                        ]
                    },
                    {
                        "splits": [
                            {
                                "name": "properties",
                                "metric_key": "properties",
                                "display_label": "properties",
                                "instance_index": 1,
                                "pages": [1, 2, 3, 4],
                                "page_start": 1,
                                "page_end": 4,
                                "partitions": [{"key": "10*310146", "pages": [1]}],
                            }
                        ]
                    },
                    {
                        "splits": [
                            {
                                "name": "properties",
                                "metric_key": "properties",
                                "display_label": "properties",
                                "instance_index": 1,
                                "pages": [1, 2, 3],
                                "page_start": 1,
                                "page_end": 3,
                                "partitions": [{"key": "10*310146", "pages": [1]}],
                            }
                        ]
                    },
                ],
            },
            "usage": {"credits": 4.0},
        }
    )

    assert response.splits[0].model_dump(exclude_none=True) == {
        "name": "properties",
        "pages": [1, 2, 3],
        "partitions": [
            {
                "key": "10*310146",
                "pages": [1],
            }
        ],
    }
    assert response.splits[0].partitions[0].model_dump() == {
        "key": "10*310146",
        "pages": [1],
    }
    assert response.consensus is not None
    assert response.consensus.likelihoods is not None
    assert response.consensus.likelihoods.model_dump(exclude_none=True) == {
        "splits": [
            {
                "likelihood": 0.9703,
                "name": 0.9703,
                "pages": [1.0, 1.0, 1.0],
                "partitions": [
                    {
                        "likelihood": 1.0,
                        "key": 1.0,
                        "pages": [1.0],
                    }
                ],
            }
        ]
    }
    assert [choice.splits[0].pages for choice in response.consensus.choices] == [
        [1, 2, 3],
        [1, 2, 3],
        [1, 2, 3, 4],
        [1, 2, 3],
    ]
    assert response.model_dump(exclude_none=True)["splits"][0] == {
        "name": "properties",
        "pages": [1, 2, 3],
        "partitions": [{"key": "10*310146", "pages": [1]}],
    }
