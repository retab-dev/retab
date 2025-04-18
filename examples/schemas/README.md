# 🧠 Schema Examples

This folder showcases two ways to define schemas for extracting structured data from documents using **UiForm**:

---

## ✨ What’s in Here

| File | Approach | Description |
|------|----------|-------------|
| `pydantic_calendar_event.py` | **Pydantic** | Define a `CalendarEvent` schema using Python classes with field-level control and built-in validation. Recommended for Python developers. |
| `json_schema_calendar_event.py` | **JSON Schema** | The same use case, but using raw JSON Schema as a Python dictionary — ideal for non-Python contexts or integration-first workflows. |

---

## 🧪 How to Run

Make sure you’ve installed the SDK:

```bash
pip install uiform
