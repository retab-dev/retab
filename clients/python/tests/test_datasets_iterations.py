from typing import Any, Awaitable, Dict, Literal, TypeVar, get_args

import nanoid  # type: ignore
import pytest

from retab import AsyncRetab, Retab
from retab.types.projects.datasets import Dataset, DatasetDocument
from retab.types.projects.iterations import Iteration, IterationDocument, SchemaOverrides
from retab.types.inference_settings import InferenceSettings

T = TypeVar("T")

TEST_MODEL = "retab-micro"


async def await_or_return(obj: T | Awaitable[T]) -> T:
    if isinstance(obj, Awaitable):
        return await obj
    else:
        return obj


ClientType = Literal["sync", "async"]


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_dataset_crud(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test basic CRUD operations for datasets."""
    client = sync_client if client_type == "sync" else async_client
    project_name = f"test_dataset_crud_{nanoid.generate()}"

    # Create a project first
    project = await await_or_return(
        client.projects.create(name=project_name, json_schema=booking_confirmation_json_schema)
    )
    project_id = project.id

    try:
        # CREATE dataset
        dataset_name = f"dataset_{nanoid.generate()}"
        dataset = await await_or_return(
            client.projects.datasets.create(
                project_id=project_id,
                name=dataset_name,
                base_json_schema=booking_confirmation_json_schema,
            )
        )
        assert isinstance(dataset, Dataset)
        assert dataset.name == dataset_name
        assert dataset.project_id == project_id
        dataset_id = dataset.id

        try:
            # GET dataset
            retrieved = await await_or_return(
                client.projects.datasets.get(project_id=project_id, dataset_id=dataset_id)
            )
            assert retrieved.id == dataset_id
            assert retrieved.name == dataset_name

            # LIST datasets
            datasets = await await_or_return(
                client.projects.datasets.list(project_id=project_id)
            )
            assert any(d.id == dataset_id for d in datasets)

            # UPDATE dataset
            new_name = f"updated_{nanoid.generate()}"
            updated = await await_or_return(
                client.projects.datasets.update(
                    project_id=project_id, dataset_id=dataset_id, name=new_name
                )
            )
            assert updated.name == new_name

        finally:
            # DELETE dataset
            try:
                await await_or_return(
                    client.projects.datasets.delete(project_id=project_id, dataset_id=dataset_id)
                )
            except Exception:
                pass
    finally:
        # Clean up project
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass


@pytest.mark.asyncio
@pytest.mark.parametrize("client_type", get_args(ClientType))
async def test_iteration_draft_workflow(
    sync_client: Retab,
    async_client: AsyncRetab,
    client_type: ClientType,
    booking_confirmation_json_schema: Dict[str, Any],
) -> None:
    """Test the iteration draft workflow: create → update_draft → get_schema → finalize."""
    client = sync_client if client_type == "sync" else async_client
    project_name = f"test_iteration_draft_{nanoid.generate()}"

    # Create project
    project = await await_or_return(
        client.projects.create(name=project_name, json_schema=booking_confirmation_json_schema)
    )
    project_id = project.id

    try:
        # Create dataset
        dataset = await await_or_return(
            client.projects.datasets.create(
                project_id=project_id,
                name=f"dataset_{nanoid.generate()}",
                base_json_schema=booking_confirmation_json_schema,
            )
        )
        dataset_id = dataset.id

        try:
            # CREATE iteration (starts as draft)
            iteration = await await_or_return(
                client.projects.datasets.iterations.create(
                    project_id=project_id,
                    dataset_id=dataset_id,
                )
            )
            assert isinstance(iteration, Iteration)
            assert iteration.status == "draft"
            iteration_id = iteration.id

            try:
                # GET iteration
                retrieved = await await_or_return(
                    client.projects.datasets.iterations.get(
                        project_id=project_id, dataset_id=dataset_id, iteration_id=iteration_id
                    )
                )
                assert retrieved.id == iteration_id
                assert retrieved.status == "draft"

                # LIST iterations
                iterations = await await_or_return(
                    client.projects.datasets.iterations.list(
                        project_id=project_id, dataset_id=dataset_id
                    )
                )
                assert any(i.id == iteration_id for i in iterations)

                # UPDATE DRAFT with new inference settings
                new_settings = InferenceSettings(model="retab-small", n_consensus=1)
                updated = await await_or_return(
                    client.projects.datasets.iterations.update_draft(
                        project_id=project_id,
                        dataset_id=dataset_id,
                        iteration_id=iteration_id,
                        inference_settings=new_settings,
                    )
                )
                assert updated.draft.inference_settings.model == "retab-small"

                # GET SCHEMA (without draft)
                schema_response = await await_or_return(
                    client.projects.datasets.iterations.get_schema(
                        project_id=project_id,
                        dataset_id=dataset_id,
                        iteration_id=iteration_id,
                    )
                )
                assert "json_schema" in schema_response

                # GET SCHEMA (with draft applied)
                draft_schema_response = await await_or_return(
                    client.projects.datasets.iterations.get_schema(
                        project_id=project_id,
                        dataset_id=dataset_id,
                        iteration_id=iteration_id,
                        use_draft=True,
                    )
                )
                assert "json_schema" in draft_schema_response

                # FINALIZE iteration (promotes draft → main, status → completed)
                finalized = await await_or_return(
                    client.projects.datasets.iterations.finalize(
                        project_id=project_id, dataset_id=dataset_id, iteration_id=iteration_id
                    )
                )
                assert finalized.status == "completed"
                assert finalized.inference_settings.model == "retab-small"

            finally:
                # DELETE iteration
                try:
                    await await_or_return(
                        client.projects.datasets.iterations.delete(
                            project_id=project_id, dataset_id=dataset_id, iteration_id=iteration_id
                        )
                    )
                except Exception:
                    pass
        finally:
            try:
                await await_or_return(
                    client.projects.datasets.delete(project_id=project_id, dataset_id=dataset_id)
                )
            except Exception:
                pass
    finally:
        try:
            await await_or_return(client.projects.delete(project_id))
        except Exception:
            pass
