import { Retab } from '@retab/node';

// Initialize the client with your API key
const client = new Retab({ apiKey: "YOUR_RETAB_API_KEY" });

// Extract data from a single document using deployment
const response = await client.projects.extract({
    project_id: "proj_97u_9bbMifNjGH175NTwi",
    iteration_id: "eval_iter_5Rl6RKAUn8jLaiKgnvHie",
    document: "path/to/document.pdf"
});

console.log(response);