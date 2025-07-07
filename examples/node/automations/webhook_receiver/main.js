// ---------------------------------------------
// Express app to receive Retab webhook events (with ngrok tunnel)
// ---------------------------------------------

import express from 'express';
import { config } from 'dotenv';
import ngrok from '@ngrok/ngrok';

// Load environment variables
config();

const app = express();
app.use(express.json());

const port = 8000;

// Example webhook request structure
const exampleBody = {
  completion: {
    id: "id",
    created: 0,
    model: "gpt-4.1-nano",
    object: "chat.completion",
    likelihoods: {},
    choices: [
      {
        index: 0,
        message: {
          content: '{"message": "Hello, World!"}',
          role: "assistant"
        },
        finish_reason: null,
        logprobs: null
      }
    ]
  },
  file_payload: {
    filename: "example.pdf",
    url: "data:application/pdf;base64,the_content_of_the_pdf_file"
  }
};

app.post('/webhook', (req, res) => {
  try {
    const request = req.body;
    const content = request.completion?.choices?.[0]?.message?.content || "{}";
    const parsedData = JSON.parse(content);
    
    console.log('\n✅ Webhook received:');
    console.log(JSON.stringify(parsedData, null, 2));
    
    res.json({ status: "success", data: parsedData });
  } catch (error) {
    console.error('Error processing webhook:', error);
    res.status(400).json({ status: "error", message: error.message });
  }
});

async function startServer() {
  const authToken = process.env.NGROK_AUTH_TOKEN;
  if (!authToken) {
    throw new Error('NGROK_AUTH_TOKEN environment variable is required');
  }

  // Start Express server
  app.listen(port, '0.0.0.0', async () => {
    console.log(`Server running on port ${port}`);
    
    try {
      // Create ngrok tunnel
      const listener = await ngrok.connect({ 
        addr: port, 
        authtoken: authToken 
      });
      
      const publicUrl = listener.url();
      const webhookUrl = `${publicUrl}/webhook`;

      console.log('\n🌍 Ngrok tunnel established!');
      console.log(`📬 Webhook URL: ${webhookUrl}`);
      console.log('\n📬 Test with curl:');
      console.log('-'.repeat(80));
      console.log(`curl -X POST ${webhookUrl} -H "Content-Type: application/json" -d '${JSON.stringify(exampleBody)}'`);
      console.log('-'.repeat(80));
    } catch (error) {
      console.error('Failed to create ngrok tunnel:', error);
    }
  });
}

startServer().catch(console.error);