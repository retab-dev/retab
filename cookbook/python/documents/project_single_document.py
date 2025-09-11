from retab import Retab

# Initialize the client with your API key
client = Retab(api_key="YOUR_RETAB_API_KEY")

# Submit a single document
completion = client.projects.extract(
    project_id="proj_97u_9bbMifNjGH175NTwi",
    iteration_id="eval_iter_5Rl6RKAUn8jLaiKgnvHie",
    document="path/to/document.pdf"
)

print(completion)