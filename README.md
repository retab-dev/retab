# uiform

UiForm is a **modern**, **flexible**, and **AI-powered** document processing API that helps you:

- Create from JSON schemas and Pydantic models with zero boilerplate
- Add AI capabilities for automated document processing, that is compatible with any data structure
- Create annotated datasets to distill or finetune your models

Our goal is to make the process of analyzing documents and unstructured data as **easy** and **transparent** as possible. You come with your own API key from your favorite AI provider, and we handle the rest.

## Quickstart

<Steps>
  <Step title="Load a JSON Schema and some Documents">
    Setup your API keys and load a JSON Schema and some Documents.
  </Step>
  <Step title="Extract data from your documents with our Python SDK">
    Use UiForm to extract data from your documents
  </Step>
</Steps>





### 1- Load a JSON Schema and some Documents

Save this [JSON Schema](https://github.com/UiForm/uiform/blob/main/notebooks/freight/booking_confirmation_json_schema.json) as `json_schema.json`

Download this [example document](https://github.com/UiForm/uiform/blob/main/notebooks/freight/booking_confirmation.jpg) as `example.jpg`

```python main.py
import json

with open("booking_confirmation_json_schema.json", "r") as f:
    json_schema = json.load(f)

document = "booking_confirmation.jpg"
```


### 2 - Extract data from your documents with our Python SDK


To get started, install the `uiform` package using pip:

```bash
pip install uiform
```

Then, populate your `env` variables with your API keys:

```bash .env
OPENAI_API_KEY=YOUR-API-KEY # Your AI provider API key. Compatible with OpenAI, Anthropic, xAI.
UIFORM_API_KEY=sk_xxxxxxxxx
```

Use the `UiForm` client to extract data from your documents:

```python main.py
import json
from uiform.client import UiForm

with open("booking_confirmation_json_schema.json", "r") as f:
    json_schema = json.load(f)

document = "booking_confirmation.jpg"

client = UiForm()
response = client.documents.extract(
    json_schema = json_schema,
    document = document,
    model="gpt-4o-mini-2024-07-18",
    temperature=0
)
```

And that's it ! You can start processing documents at scale ! 
You have 1000 free requests to get started, and you can [subscribe](https://www.uiform.com) to the pro plan to get more.

But this minimalistic example is not much more useful than the bare openAI API. Continue reading to learn more about how to use UiForm **to its full potential**.

----

## Go further

- [Additional parameters](https://docs.uiform.com/document-api/additional-parameters)
- [Finetuning](https://docs.uiform.com/document-api/finetuning)
- Prompt optimization (coming soon)
- Data-Labelling with our AI-powered annotator (coming soon)

----

## Examples

You can view minimal notebooks that demonstrate how to use UiForm to process documents:

- [Quickstart](https://github.com/UiForm/uiform/blob/main/notebooks/Quickstart.ipynb)
- [Schema Creation](https://github.com/UiForm/uiform/blob/main/notebooks/Schema_creation.ipynb)
- [Finetuning](https://github.com/UiForm/uiform/blob/main/notebooks/Finetuning.ipynb)
