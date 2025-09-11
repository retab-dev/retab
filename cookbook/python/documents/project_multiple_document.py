from retab import Retab

# Initialize the client with your API key
client = Retab(api_key="YOUR_RETAB_API_KEY")

# Submit multiple documents
documents = ["path/to/document1.pdf", "path/to/document2.pdf"]

completion = client.projects.extract(
    project_id="proj_97u_9bbMifNjGH175NTwi",
    iteration_id="eval_iter_5Rl6RKAUn8jLaiKgnvHie",
    documents=documents
)

print(completion)