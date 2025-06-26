# ⚡ Quickstart Examples

This folder contains simple, focused examples to help you quickly get started with Retab.

---

## 📥 Message Creation

Learn how to create LLM-ready messages from documents using the `Retab.documents.create_messages()` API.

| File | Description |
|------|-------------|
| `create_messages.py` | Minimal example: create messages from a document and send to an LLM for summarization |

---

## 📤 Schema-Based Extraction

Use a JSON schema to extract structured data from documents using different API styles.

| File | Description |
|------|-------------|
| `extract_api.py` | One-liner `.parse()` method that handles everything — perfect for quick results |
| `openai_completion_api.py` | Full pipeline using **OpenAI Chat Completions API** and schema-based prompting |
| `openai_responses_api.py` | Same as above, but uses the **OpenAI Responses API** for greater flexibility |

---

## 🧪 How to Run

Make sure you’ve installed the SDK:

```bash
pip install retab
```
