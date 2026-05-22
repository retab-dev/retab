"""Canonical types module for the ``workflows.blocks.executions`` accessor.

The single canonical access path for block-execution types is
``retab.types.workflows.blocks.executions``, mirroring the resource
accessor ``client.workflows.blocks.executions``. The class bodies live
in :mod:`retab.types.workflows.model` for now — this module surfaces
them under the canonical path.
"""

from ..model import BlockExecutionIteration, StoredBlockExecution

__all__ = ["StoredBlockExecution", "BlockExecutionIteration"]
