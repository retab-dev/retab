# Retab

<div align="center" style="margin-bottom: 1em;">

<img src="https://raw.githubusercontent.com/Retab/retab/refs/heads/main/assets/retab-logo.png" alt="Retab Logo" width="150">


  *The AI Automation Platform*

Made with love by the team at [Retab](https://retab.dev) 🤍.

[Our Website](https://retab.dev) | [Documentation](https://docs.retab.dev/get-started/introduction) | [Discord](https://discord.com/invite/vc5tWRPqag) | [Twitter](https://x.com/retabAPI)


</div>

---

### What is Retab?

Retab solves all the major challenges in document processing with LLMs:

1. **Universal Document Processing**: Convert any file type (PDFs, Excel, emails, etc.) into LLM-ready format without writing custom parsers
2. **Structured, Schema-driven Extraction**: Get consistent, reliable outputs using schema-based prompt engineering
3. **Deployments**: Create custom mailboxes and links to process documents at scale
4. **Evaluations**: Evaluate the performance of models against annotated datasets
5. **Optimizations**: Identify the most used deployments and help you finetune models to reduce costs and improve performance



We are offering you all the software-defined primitives to build your own document processing solutions. We see it as **Stripe** for document processing.

Our goal is to make the process of analyzing documents and unstructured data as **easy** and **transparent** as possible.

Many people haven't yet realized how powerful LLMs have become at document processing tasks - we're here to help **unlock these capabilities**.

---

## How it works

Retab allows you to easily create document processing deployments. Here is the general workflow:

```mermaid
sequenceDiagram
    User ->> Retab: File Upload
    Retab -->> Retab: Preprocessing
    Retab ->> AI Provider: Request on your behalf
    AI Provider -->> Retab:  Structured Generation
    Retab ->> Webhook: Send result
    Retab ->> User: Send Confirmation
```

---



We currently support [OpenAI](https://platform.openai.com/docs/overview), [Anthropic](https://www.anthropic.com/api), [Gemini](https://aistudio.google.com/) and [xAI](https://x.ai/api) models.

You come with your own API key from your favorite AI provider, and we handle the rest.

---

## Go further

- [Quickstart](/get-started/quickstart)
- [Prompt Engineering Guide](/get-started/prompting-with-the-JSON-schema)
- [Overview of the core concepts](/core/Overview)
- [Create a deployment](/core/Deployments#mailbox)

---

## Jupyter Notebooks

You can view minimal notebooks that demonstrate how to use Retab to process documents:
- [Mailbox creation quickstart](https://github.com/Retab-dev/retab/blob/main/notebooks/mailboxes_quickstart.ipynb)
- [Upload Links creation quickstart](https://github.com/Retab-dev/retab/blob/main/notebooks/links_quickstart.ipynb)
- [Document Extractions quickstart](https://github.com/Retab-dev/retab/blob/main/notebooks/Quickstart.ipynb)
- [Document Extractions quickstart - Async](https://github.com/Retab-dev/retab/blob/main/notebooks/Quickstart-Async.ipynb)

--- 


## Community

Let's create the future of document processing together!

Join our [discord community](https://discord.com/invite/vc5tWRPqag) to share tips, discuss best practices, and showcase what you build. Or just [tweet](https://x.com/retabAPI) at us.

We can't wait to see how you'll use Retab.

- [Discord](https://discord.com/invite/vc5tWRPqag)
- [Twitter](https://x.com/retabAPI)

---

## Roadmap

We share our roadmap publicly on [Github](https://github.com/Retab-dev/retab)

Among the features we're working on:

- [ ] Node.js SDK
- [ ] Chat-based interface in the evaluation platform
- [ ] Low-level speed optimizations for Evals Frontend
- [ ] New JSON Reconciliation Model
- [ ] Add more templates
- [ ] DoubleCheck API (Agentic verification of the results)
- [ ] Deep Research