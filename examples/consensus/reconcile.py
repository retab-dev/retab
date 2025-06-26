from retab import Retab

client = Retab()

# Multiple extraction results to reconcile
extractions = [
    {
        "title": "Quantum Algorithms in Interstellar Navigation",
        "authors": ["Dr. Stella Voyager", "Dr. Nova Star", "Dr. Lyra Hunter"],
        "year": 2025,
        "keywords": ["quantum computing", "space navigation", "algorithms"],
    },
    {
        "title": "Quantum Algorithms for Interstellar Navigation",
        "authors": ["Dr. S. Voyager", "Dr. N. Star", "Dr. L. Hunter"],
        "year": 2025,
        "keywords": ["quantum algorithms", "interstellar navigation", "space travel"],
    },
    {
        "title": "Application of Quantum Algorithms in Space Navigation",
        "authors": ["Stella Voyager", "Nova Star", "Lyra Hunter"],
        "year": 2025,
        "keywords": ["quantum computing", "navigation", "space exploration"],
    },
]

# Reconcile the different extraction results into a consensus
response = client.consensus.reconcile(list_dicts=extractions, mode="aligned")

consensus_result = response.consensus_dict
consensus_confidence = response.likelihoods

print(f"Consensus: {consensus_result}")
print(f"Confidence scores: {consensus_confidence}")
