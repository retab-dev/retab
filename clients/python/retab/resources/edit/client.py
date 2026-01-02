"""
Edit SDK client - Wrapper for document editing functionality.

Provides access to:
- edit.agent.fill() - Agent-based document editing (PDF, DOCX, PPTX, XLSX)
- edit.templates.* - Template-based PDF form filling
"""

from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from .templates import Templates, AsyncTemplates
from .agent import Agent, AsyncAgent


class Edit(SyncAPIResource):
    """Edit API wrapper for synchronous usage.
    
    Sub-clients:
        agent: Agent-based document editing (fill any document with AI)
        templates: Template-based PDF form filling (for batch processing)
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.agent = Agent(client=client)
        self.templates = Templates(client=client)


class AsyncEdit(AsyncAPIResource):
    """Edit API wrapper for asynchronous usage.
    
    Sub-clients:
        agent: Agent-based document editing (fill any document with AI)
        templates: Template-based PDF form filling (for batch processing)
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.agent = AsyncAgent(client=client)
        self.templates = AsyncTemplates(client=client)
