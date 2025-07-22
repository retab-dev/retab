# Retab

<div align="center" style="margin-bottom: 1em;">

<img src="https://raw.githubusercontent.com/Retab-dev/retab/refs/heads/main/assets/retab-logo.png" alt="Retab Logo" width="150">


  *The AI Automation Platform*

Made with love by the team at [Retab](https://retab.com) 🤍.

[Our Website](https://retab.com) | [Documentation](https://docs.retab.com/get-started/introduction) | [Discord](https://discord.com/invite/vc5tWRPqag) | [Twitter](https://x.com/retabdev)


</div>

---

### What is Retab?

[Retab](https://retab.com) is the complete developer platform and SDK for shipping state-of-the-art document processing in the age of LLMs. 

We provide the best-in-class preprocessing, help you generate prompts & extraction schemas that fits your preferred model providers, iterate & evaluate the accuracy of your configuration, and **shit fast** your automation directly in your code or with your prefered platforms such as [n8n](https://n8n.io/) or [Dify](https://dify.ai/).

**A new, lighter paradigm**
Large Language Models collapse entire layers of legacy OCR pipelines into a single, elegant abstraction. When a model can read, reason, and structure text natively, we no longer need brittle heuristics, handcrafted parsers, or heavyweight ETL jobs. Instead, we can expose a small, principled API: "give me the document, tell me the schema, and get back structured truth." Complexity evaporates, reliability rises, speed follows, and costs fall—because every component you remove is one that can no longer break. LLM‑first design lets us focus less on plumbing and more on the questions we actually want answered.

Many people haven't yet realized how powerful LLMs have become at document processing tasks - we're here to help **unlock these capabilities**.

We are convinced that **Human in the loop is dogsh*t**, therefore offering you all the software-defined primitives to build your own document processing solutions. We see it as **Stripe** for document processing.

Check our [documentation](https://docs.retab.com/overview/introduction).

_Join our stargazers! ⭐️_

---

## Useful Links

* [x] **API**: [Documentation](https://docs.retab.com/api-reference/introduction)
* [x] **SDKs**: [Python & JavaScript SDK](https://docs.retab.com/overview/quickstart)
* [x] **Low-code Frameworks**: [Dify](https://marketplace.dify.ai/plugins/retab_team/retab)

**Roadmap**
* [ ] n8n plugin
* [ ] Node.js SDK
* [ ] Schema optimization autopilot
* [ ] Sources API

---

## Features

1. **Universal Document Preprocessing**: Convert any file type (PDFs, Excel, emails, etc.) into LLM-ready format without writing custom parsers
2. **Structured, Schema-driven Extraction**: Get consistent, reliable outputs using schema-based prompt engineering
3. **Processors**: Publish a live, stable, shareable document processor.
4. **Automations**: Create document processing workflows that can be triggered by events (mailbox, upload link, endpoint, outlook plugin).
5. **Projects**: Evaluate the performance of models against annotated datasets
6. **Optimizations**: Identify the most used processors and help you finetune models to reduce costs and improve performance

---

## API Key

To use the API, you need to sign up on [Retab](https://www.retab.com/).

<img src="./assets/API-key.png" alt="API Key" width="150">

---

## SDK

1. Install the SDK
```bash
pip install retab
```

2. Generate a Schema
```bash
from pathlib import Path
from retab import Retab
client = Retab(api_key="YOUR_RETAB_API_KEY")

response = client.schemas.generate(
    documents=["Invoice.pdf"],
    model="gpt-4.1",          # or any model your plan supports
    temperature=0.0,          # keep the generation deterministic
    modality="native",        # "native" = let the API decide best modality
)
```

3. Exrtact Data
```bash
from pathlib import Path
from retab import Retab

from retab import Retab
client = Retab()

response = client.documents.extract(
    json_schema = "Invoice_schema.json",
    document = "Invoice.pdf",
    model="gpt-4.1-nano",
    temperature=0
)

print(response)
```

---

## Project Vision

On the [Platform](https://www.retab.com/), *Projects* provide a systematic way to test and validate your extraction schemas against known ground truth data. Think of it as unit testing for document AI—you can measure accuracy, compare different models, and optimize your extraction pipelines with confidence.

The project workflow for schema optimization:
1. Run initial project → identify low-accuracy fields
2. Refine descriptions and add reasoning prompts → re-run project
3. Compare accuracy improvements → iterate until satisfied
4. Deploy optimized schema to production

```bash
from retab import Retab

client = Retab()

# Submit a single document
completion = client.deployments.extract(
    project_id="eval_***",
    iteration_id="base-configuration", # or the configuration that gave you the best precision score
    document="path/to/document.pdf"
)

print(completion)
```

Projects give you an easy-to-use automation engine easy to integrate in your codebase and workflows.

Check our [documentation](https://docs.retab.com/core-concepts/Projects).

---

## Code examples

You can check our Github repository to see code examples: [python examples](https://github.com/Retab-dev/retab/tree/main/examples) and [jupyter notebooks](https://github.com/Retab-dev/retab-nodejs/tree/main/notebooks).

---

## Community

Let's create the future of document processing together!

Join our [discord community](https://discord.com/invite/vc5tWRPqag) to share tips, discuss best practices, and showcase what you build. Or just [tweet](https://x.com/retabdev) at us.

We can't wait to see how you'll use Retab.

* [Discord](https://discord.com/invite/vc5tWRPqag)
* [Twitter](https://x.com/retabdev)