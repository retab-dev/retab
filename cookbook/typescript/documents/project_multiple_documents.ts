import { Retab } from '@retab/node';

// Initialize the client with your API key
const client = new Retab();

// Extract data from multiple documents
const documents = ["path/to/document1.pdf", "path/to/document2.pdf"];

const response = await client.projects.extract({
    project_id: "proj_97u_9bbMifNjGH175NTwi",
    iteration_id: "eval_iter_5Rl6RKAUn8jLaiKgnvHie",
    documents: documents
});

console.log(response);


