# 📁 Examples

Welcome to our `examples/` folder, part of the **UiForm** open-source repo.

We have here a curated set of scripts demonstrating how to use UiForm to build document-based LLM workflows.

These examples are grouped by theme to help you get started quickly, learn how to define schemas, and integrate with different providers or webhook backends.

---

## 📂 Folder Structure

```bash
examples/
├── quickstarts/                       # End-to-end examples to get up and running fast
├── schemas/                           # Pydantic & JSON Schema examples
├── airbnb/                            # Structured pitch deck extraction
├── other_providers/                   # Other providers examples
└── webhook_server/                    # FastAPI server for receiving automation results
```


---

## 🏁 Quickstarts

These examples show how to use the `uiform` SDK to create messages and process documents through your preferred LLM provider.

| File | Description |
|------|-------------|
| `create_messages.py` | Minimal flow for converting a document into LLM-ready messages |
| `extract_api.py` | Extract responses from documents using direct API interaction |
| `full_completion_api.py` | End-to-end processing with OpenAI completion |
| `full_create_messages.py` | More detailed `create_messages` example |
| `full_responses_api.py` | Handles full response objects with output formatting |

---

## 🧠 Schema-Based Prompting

Use UiForm’s schema support to guide LLM responses more precisely, using `X-FieldPrompt`, `X-ReasoningPrompt`, and more.

| File | Description |
|------|-------------|
| `pydantic_calendar_event.py` | Define a calendar event schema using Pydantic (clean, type-safe) |
| `json_schema_calendar_event.py` | Same use case, defined using raw JSON Schema (framework-agnostic) |

---

## 🏢 Structured Output from PDF

We let you here with a business-focused examples to extract structured information from startup pitch decks or investor docs (here AirBnB).

| File | Description |
|------|-------------|
| `airbnb_schema.py` | Pydantic model for company structure, team, and business info |
| `run_airbnb_extraction.py` | Full pitch deck parsing into structured data |

---

## 🔄 Other LLM Providers

Demonstrate how to switch out OpenAI with other LLM providers like Gemini or Anthropic.

| File | Description |
|------|-------------|
| `gemini.py` | Use Google’s Gemini API with UiForm messages |
| `openai.py` | Simple OpenAI-based schema extraction (baseline example) |

---

## 🔗 Webhook Integration

UiForm sends structured document outputs to your webhook. This example shows how to receive and process that data using FastAPI.

| File | Description |
|------|-------------|
| `main.py` | FastAPI server with a `/webhook` endpoint to receive structured document results |

---

## 🔍 Fuzzy Matching Utility

We let you here with a standalone script to perform fuzzy matching between a target record and a dataset using Levenshtein similarity — helpful for deduplication or record validation.

| File | Description |
|------|-------------|
| `code_matcher.py` | Load a CSV and find top-k closest records to your input query based on text similarity |

---

## 🚀 How to Use These Examples

1. **Install the SDK**  
   ```bash
   pip install uiform
   ```

2. **Set your environment variables**
   ```bash
   UIFORM_API_KEY=sk_uiform_...
   OPENAI_API_KEY=sk-...
   ```

3. **Run the example**
   Navigate to the example folder and execute:

   ```bash
   python quickstarts/example_create_messages.py
   ```

---

# 🧪 Want to Contribute?

If you have a useful example:

Keep it focused and documented.

Place it in the appropriate folder.

Feel free to open a PR!

---

Made with love by the team at [UiForm](https://uiform.com) 🤍.

Join the community on [Discord](https://discord.com/invite/vc5tWRPqag) and share what you’re building!
