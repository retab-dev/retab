import asyncio
import hashlib
import json
import re
import shutil
import tempfile
import time
from concurrent.futures import ThreadPoolExecutor
from io import IOBase
from pathlib import Path
from typing import IO, Any, Literal, Optional

from anthropic import Anthropic
from openai import OpenAI
from openai.types.chat.chat_completion_message_param import ChatCompletionMessageParam
from pydantic import BaseModel
from tqdm import tqdm

from .._resource import AsyncAPIResource, SyncAPIResource
from .._utils.ai_models import assert_valid_model_extraction, find_provider_from_model
from .._utils.chat import convert_to_anthropic_format, convert_to_openai_format, separate_messages
from .._utils.display import Metrics, display_metrics, process_dataset_and_compute_metrics
from .._utils.json_schema import load_json_schema
from ..types.chat import ChatCompletionUiformMessage
from ..types.modalities import Modality
from ..types.schemas.object import Schema


class FinetuningJSON(BaseModel):
    messages: list[ChatCompletionUiformMessage]


FinetuningJSONL = list[FinetuningJSON]
from typing import TypedDict


class BatchJSONLResponseFormat(TypedDict):
    type: str
    json_schema: dict[str, Any]


class BatchJSONLBody(TypedDict):
    model: str
    messages: list[ChatCompletionMessageParam]
    temperature: float
    response_format: BatchJSONLResponseFormat


class BatchJSONL(TypedDict):
    custom_id: str
    method: str
    url: str
    body: BatchJSONLBody


class BatchJSONLResponseUsageTokenDetails(BaseModel):
    cached_tokens: int
    audio_tokens: int


class BatchJSONLResponseUsageCompletionDetails(BaseModel):
    reasoning_tokens: int
    audio_tokens: int
    accepted_prediction_tokens: int
    rejected_prediction_tokens: int


class BatchJSONLResponseUsage(BaseModel):
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    prompt_tokens_details: BatchJSONLResponseUsageTokenDetails
    completion_tokens_details: BatchJSONLResponseUsageCompletionDetails


class BatchJSONLResponseChoice(BaseModel):
    index: int
    message: ChatCompletionMessageParam
    logprobs: None | Any
    finish_reason: str


class BatchJSONLResponseBody(BaseModel):
    id: str
    object: str
    created: int
    model: str
    choices: list[BatchJSONLResponseChoice]
    usage: BatchJSONLResponseUsage
    service_tier: str
    system_fingerprint: str


class BatchJSONLResponseInner(BaseModel):
    status_code: int
    request_id: str
    body: BatchJSONLResponseBody


class BatchJSONLResponse(BaseModel):
    id: str
    custom_id: str
    response: BatchJSONLResponseInner
    error: None | str


class BaseDatasetsMixin:
    def _dump_training_set(self, training_set: list[dict[str, Any]], dataset_path: Path | str) -> None:
        with open(dataset_path, 'w', encoding='utf-8') as file:
            for entry in training_set:
                file.write(json.dumps(entry) + '\n')


class Datasets(SyncAPIResource, BaseDatasetsMixin):
    """Datasets API wrapper"""

    # TODO : Maybe at some point we could add some visualization methods... but the multimodality makes it hard... # client.datasets.plot.tsne()... # client.datasets.plot.umap()...
    def pprint(self, dataset_path: Path | str, input_token_price: Optional[float] = None, output_token_price: Optional[float] = None) -> Metrics:
        """Print a summary of the contents and statistics of a JSONL file.

        This method analyzes the JSONL file and displays various metrics and statistics
        about the dataset contents.

        Inspired from : https://gist.github.com/nmwsharp/54d04af87872a4988809f128e1a1d233

        Args:
            dataset_path: Path to the JSONL file to analyze
            completion_token_price: Price per completion token
            input_token_price: Price per input token
        """

        computed_metrics = process_dataset_and_compute_metrics(dataset_path)
        display_metrics(computed_metrics, input_token_price=input_token_price, output_token_price=output_token_price)
        return computed_metrics

    def save(
        self,
        json_schema: dict[str, Any] | Path | str,
        document_annotation_pairs_paths: list[dict[str, Path | str]],
        dataset_path: Path | str,
        image_resolution_dpi: int | None = None,
        browser_canvas: Literal['A3', 'A4', 'A5'] | None = None,
        modality: Modality = "native",
    ) -> None:
        """Save document-annotation pairs to a JSONL training set.

        Args:
            json_schema: The JSON schema for validation, can be a dict, Path, or string
            document_annotation_pairs_paths: {document_fpath: Path | str, annotation_fpath: Path | str} List of dictionaries containing document and annotation file paths
            jsonl_path: Output path for the JSONL training file
            modality: The modality to use for document processing ("native" by default)
        """
        json_schema = load_json_schema(json_schema)
        schema_obj = Schema(json_schema=json_schema)

        with open(dataset_path, 'w', encoding='utf-8') as file:
            for pair_paths in tqdm(document_annotation_pairs_paths, desc="Processing pairs", position=0):
                document_message = self._client.documents.create_messages(document=pair_paths['document_fpath'], modality=modality, image_resolution_dpi=image_resolution_dpi, browser_canvas=browser_canvas)

                with open(pair_paths['annotation_fpath'], 'r') as f:
                    annotation = json.loads(f.read())
                assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

                entry = {"messages": schema_obj.messages + document_message.messages + [assistant_message]}
                file.write(json.dumps(entry) + '\n')

    def change_schema(
        self,
        input_dataset_path: Path | str,
        json_schema: dict[str, Any] | Path | str,
        output_dataset_path: None | Path | str = None,
        inplace: bool = False,
    ) -> None:
        """Change the system prompt in a dataset to match a new schema.

        Args:
            input_dataset_path: Path to the input JSONL dataset file
            output_dataset_path: Path to the output JSONL dataset file
            json_schema: The new JSON schema for validation, can be a dict, Path, or string
            inplace: If True, overwrite the input dataset with the changes
        """
        json_schema = load_json_schema(json_schema)
        schema_obj = Schema(json_schema=json_schema)

        # Determine the path to write to
        if inplace:
            assert output_dataset_path is None, "Cannot specify inplace=True and output_dataset_path not None"
            target_path = input_dataset_path
        else:
            assert output_dataset_path, "Cannot save the file if inplace=False and output_dataset_path is None"
            target_path = output_dataset_path

        assert isinstance(target_path, Path) or isinstance(target_path, str)

        # Use a temporary file to write the updated content
        with tempfile.NamedTemporaryFile('w', delete=False, encoding='utf-8') as temp_file:
            with open(input_dataset_path, 'r', encoding='utf-8') as infile:
                for line in infile:
                    entry = json.loads(line)
                    messages = entry.get("messages", [])

                    # Remove existing system prompt if it exists
                    if messages and messages[0].get("role") in ("developer", "system"):
                        messages = messages[1:]

                    # Add the new system prompt from schema_obj.messages
                    updated_messages = schema_obj.messages + messages
                    updated_entry = {"messages": updated_messages}

                    # Write the updated entry to the temporary file
                    temp_file.write(json.dumps(updated_entry) + '\n')

        # Replace the original file with the temporary file
        shutil.move(temp_file.name, target_path)

    def stich_and_save(
        self,
        json_schema: dict[str, Any] | Path | str,
        pairs_paths: list[dict[str, Path | str | list[Path | str] | list[str] | list[Path]]],
        dataset_path: Path | str,
        modality: Modality = "native",
    ) -> None:
        """Stitch multiple documents and their annotations into a JSONL training set.

        This method processes and combines multiple documents into messages, creating document-annotation
        pairs that are saved to a JSONL file. Each document is processed according to the specified
        modality and combined with its corresponding annotation.


        Args:
            json_schema: The JSON schema for validation, can be a dict, Path, or string
            pairs_paths: List of dictionaries containing document and annotation file paths
            jsonl_path: Output path for the JSONL training file
            modality: The modality to use for document processing ("native" by default)
        """

        json_schema = load_json_schema(json_schema)
        schema_obj = Schema(json_schema=json_schema)
        training_set = []

        for pair_paths in tqdm(pairs_paths):
            document_messages: list[ChatCompletionUiformMessage] = []

            if isinstance(pair_paths['document_fpath'], str) or isinstance(pair_paths['document_fpath'], Path):
                document_message = self._client.documents.create_messages(document=pair_paths['document_fpath'], modality=modality)
                document_messages.extend(document_message.messages)

            else:
                assert isinstance(pair_paths['document_fpath'], list)
                for document_fpath in pair_paths['document_fpath']:
                    document_message = self._client.documents.create_messages(document=document_fpath, modality=modality)
                    document_messages.extend(document_message.messages)

            # Use context manager to properly close the file
            assert isinstance(pair_paths['annotation_fpath'], Path) or isinstance(pair_paths['annotation_fpath'], str)
            with open(pair_paths['annotation_fpath'], 'r') as f:
                annotation = json.loads(f.read())
            assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

            # Add the complete message set as an entry
            training_set.append({"messages": schema_obj.messages + document_messages + [assistant_message]})

        self._dump_training_set(training_set, dataset_path)

    #########################################
    ##### ENDPOINTS THAT MAKE LLM CALLS #####
    #########################################

    def _initialize_model_client(self, model: str) -> tuple[OpenAI | Anthropic, str]:
        """Initialize the appropriate client based on the model provider.

        Args:
            model: The model identifier string

        Returns:
            A tuple of (client instance, provider type string)
        """
        provider = find_provider_from_model(model)

        if provider == "OpenAI":
            return OpenAI(api_key=self._client.headers["OpenAI-Api-Key"]), provider
        elif provider == "xAI":
            return OpenAI(api_key=self._client.headers["XAI-Api-Key"], base_url="https://api.x.ai/v1"), provider
        elif provider == "Gemini":
            return OpenAI(
                api_key=self._client.headers["Gemini-Api-Key"],
                base_url="https://generativelanguage.googleapis.com/v1beta/openai/",
            ), provider
        else:
            assert provider == "Anthropic", f"Unsupported model: {model}"
            return Anthropic(api_key=self._client.headers["Anthropic-Api-Key"]), provider

    def _get_model_completion(
        self,
        client: OpenAI | Anthropic,
        provider: str,
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage],
        schema_obj: Schema,
    ) -> str:
        """Get completion from the appropriate model provider.

        Args:
            client: The initialized client instance
            provider: The provider type string
            model: The model identifier
            temperature: Temperature setting for generation
            messages: The messages to send to the model
            schema_obj: The schema object containing format information

        Returns:
            The completion string in JSON format
        """
        if provider in ["OpenAI", "xAI"]:
            assert isinstance(client, OpenAI)
            completion = client.chat.completions.create(
                model=model,
                temperature=temperature,
                messages=convert_to_openai_format(messages),
                response_format={"type": "json_schema", "json_schema": {"name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True}},
            )
            assert completion.choices[0].message.content is not None
            return completion.choices[0].message.content

        elif provider == "Gemini":
            assert isinstance(client, OpenAI)
            gemini_completion = client.chat.completions.create(
                model=model,
                temperature=temperature,
                messages=convert_to_openai_format(messages),
                response_format={"type": "json_schema", "json_schema": {"name": schema_obj.id, "schema": schema_obj.inference_gemini_json_schema, "strict": True}},
            )
            assert gemini_completion.choices[0].message.content is not None
            return gemini_completion.choices[0].message.content

        else:  # Anthropic
            assert isinstance(client, Anthropic)
            system_messages, other_messages = convert_to_anthropic_format(messages)
            anthropic_completion = client.messages.create(
                model="claude-3-5-sonnet-20241022", max_tokens=8192, temperature=temperature, system=system_messages, messages=other_messages
            )
            from anthropic.types.text_block import TextBlock

            assert isinstance(anthropic_completion.content[0], TextBlock)
            return anthropic_completion.content[0].text

    def annotate(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: list[Path | str | IOBase],
        dataset_path: Path,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0.0,
        batch_size: int = 5,
        max_concurrent: int = 3,
        root_dir: Path = Path("annotations"),
        modality: Modality = "native",
    ) -> None:
        json_schema = load_json_schema(json_schema)
        assert_valid_model_extraction(model)

        client, provider = self._initialize_model_client(model)
        schema_obj = Schema(json_schema=json_schema)

        """
        Generate annotations from document files or in-memory documents
        and create a JSONL training set in one go.

        Args:
            json_schema: The JSON schema for validation
            documents: list of documents, each can be a Path/str or an IOBase object
            dataset_path: Output path for the JSONL training file
            model: The model to use for processing
            temperature: Model temperature (0-1)
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
            root_dir: Where to store the per-document JSON annotations
        """

        def process_example(doc: Path | str | IOBase) -> dict[str, Any]:
            """
            Process a single document (either a file path or an in-memory file-like object).
            Returns a dict with pointers to the original doc and the stored annotation JSON.
            """
            if isinstance(doc, (str, Path)):
                doc_path = Path(doc)
                if not doc_path.is_file():
                    raise ValueError(f"Invalid file path: {doc_path}")
                hash_str = hashlib.md5(doc_path.as_posix().encode()).hexdigest()
            elif isinstance(doc, IO):
                file_bytes = doc.read()
                hash_str = hashlib.md5(file_bytes).hexdigest()
                doc.seek(0)
            else:
                raise ValueError(f"Unsupported document type: {type(doc)}")

            doc_msg = self._client.documents.create_messages(
                document=doc,
                modality=modality,
            )

            # Use _get_model_completion instead of duplicating provider-specific logic
            string_json = self._get_model_completion(
                client=client, provider=provider, model=model, temperature=temperature, messages=schema_obj.messages + doc_msg.messages, schema_obj=schema_obj
            )

            annotation_path = Path(root_dir) / f"annotations_{hash_str}.json"
            annotation_path.parent.mkdir(parents=True, exist_ok=True)

            with open(annotation_path, 'w', encoding='utf-8') as f:
                json.dump(string_json, f, ensure_ascii=False, indent=2)

            return {"document_fpath": str(doc_path), "annotation_fpath": str(annotation_path)}

        # Make sure output directory exists
        Path(root_dir).mkdir(parents=True, exist_ok=True)

        pairs_paths: list[dict[str, Path | str]] = []
        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            futures = []
            # Split documents into batches
            for batch in tqdm([documents[i : i + batch_size] for i in range(0, len(documents), batch_size)], desc="Processing batches"):
                # Submit batch of tasks
                batch_futures = []
                for doc in batch:
                    try:
                        future = executor.submit(process_example, doc)
                        batch_futures.append(future)
                    except Exception as e:
                        print(f"Error submitting document for processing: {e}")
                futures.extend(batch_futures)

                # Wait for batch to finish (rate limit)
                for future in batch_futures:
                    try:
                        pair = future.result()
                        pairs_paths.append(pair)
                    except Exception as e:
                        print(f"Error processing example: {e}")

        # Generate final training set from all results
        self.save(json_schema=json_schema, document_annotation_pairs_paths=pairs_paths, dataset_path=dataset_path)

    def eval(
        self,
        json_schema: dict[str, Any] | Path | str,
        dataset_path: str | Path,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0.0,
        batch_size: int = 5,
        max_concurrent: int = 3,
        display: bool = True,
    ) -> ComparisonMetrics:
        """Evaluate model performance on a test dataset.

        Args:
            json_schema: JSON schema defining the expected data structure
            dataset_path: Path to the JSONL file containing test examples
            model: The model to use for benchmarking
            temperature: Model temperature setting (0-1)
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
        """

        json_schema = load_json_schema(json_schema)
        assert_valid_model_extraction(model)
        schema_obj = Schema(json_schema=json_schema)

        # Initialize appropriate client
        client, provider = self._initialize_model_client(model)

        # Read all lines from the JSONL file
        with open(dataset_path, 'r') as f:
            lines = [json.loads(line) for line in f]

        extraction_analyses: list[ExtractionAnalysis] = []
        total_batches = (len(lines) + batch_size - 1) // batch_size

        # Create main progress bar for batches
        batch_pbar = tqdm(total=total_batches, desc="Processing batches", position=0)

        # Track running metrics
        class RunningMetrics(BaseModel):
            model: str
            accuracy: float
            levenshtein: float
            jaccard: float
            false_positive: float
            mismatched: float
            processed: int

        running_metrics: RunningMetrics = RunningMetrics(
            model=model,
            accuracy=0.0,
            levenshtein=0.0,
            jaccard=0.0,
            false_positive=0.0,
            mismatched=0.0,
            processed=0,  # number of processed examples - used in the loop to compute the running averages
        )

        def update_running_metrics(analysis: ExtractionAnalysis) -> None:
            comparison = normalized_comparison_metrics([analysis])
            running_metrics.processed += 1
            n = running_metrics.processed
            # Update running averages
            running_metrics.accuracy = (running_metrics.accuracy * (n - 1) + comparison.accuracy) / n
            running_metrics.levenshtein = (running_metrics.levenshtein * (n - 1) + comparison.levenshtein_similarity) / n
            running_metrics.jaccard = (running_metrics.jaccard * (n - 1) + comparison.jaccard_similarity) / n
            running_metrics.false_positive = (running_metrics.false_positive * (n - 1) + comparison.false_positive_rate) / n
            running_metrics.mismatched = (running_metrics.mismatched * (n - 1) + comparison.mismatched_value_rate) / n
            # Update progress bar description
            batch_pbar.set_description(
                f"Processing batches | Model: {running_metrics.model} | Acc: {running_metrics.accuracy:.2f} | "
                f"Lev: {running_metrics.levenshtein:.2f} | "
                f"IOU: {running_metrics.jaccard:.2f} | "
                f"FP: {running_metrics.false_positive:.2f} | "
                f"Mism: {running_metrics.mismatched:.2f}"
            )

        def process_example(jsonline: dict) -> ExtractionAnalysis | None:
            line_number = jsonline['line_number']
            try:
                messages = jsonline['messages']
                ground_truth = json.loads(messages[-1]['content'])
                inference_messages = messages[:-1]

                # Use _get_model_completion instead of duplicating provider-specific logic
                string_json = self._get_model_completion(client=client, provider=provider, model=model, temperature=temperature, messages=inference_messages, schema_obj=schema_obj)

                prediction = json.loads(string_json)
                analysis = ExtractionAnalysis(
                    ground_truth=ground_truth,
                    prediction=prediction,
                )
                update_running_metrics(analysis)
                return analysis
            except Exception as e:
                print(f"\nWarning: Failed to process line number {line_number}: {str(e)}")
                return None

        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            # Split entries into batches
            for batch_idx in range(0, len(lines), batch_size):
                batch = lines[batch_idx : batch_idx + batch_size]

                # Submit and process batch
                futures = [executor.submit(process_example, entry | {"line_number": batch_idx * batch_size + i}) for i, entry in enumerate(batch)]
                for future in futures:
                    result = future.result()
                    if result is not None:
                        extraction_analyses.append(result)

                batch_pbar.update(1)

        batch_pbar.close()

        # Analyze error patterns across all examples
        analysis = normalized_comparison_metrics(extraction_analyses)

        if display:
            plot_comparison_metrics(analysis=analysis, top_n=10)

        return analysis

    def benchmark(
        self,
        json_schema: dict[str, Any] | Path | str,
        dataset_path: str | Path,
        models: list[str],
        temperature: float = 0.0,
        batch_size: int = 5,
        max_concurrent: int = 3,
        print: bool = True,
        verbose: bool = False,
    ) -> list[BenchmarkMetrics]:
        """Benchmark multiple models on a test dataset.

        Args:
            json_schema: JSON schema defining the expected data structure
            dataset_path: Path to the JSONL file containing test examples
            models: List of models to benchmark
            temperature: Model temperature setting (0-1)
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
            print: Whether to print the metrics
            verbose: Whether to print all the metrics of all the function calls

        Returns:
            Dictionary mapping model names to their evaluation metrics
        """
        results: list[BenchmarkMetrics] = []

        for model in models:
            metrics: ComparisonMetrics = self.eval(
                json_schema=json_schema, dataset_path=dataset_path, model=model, temperature=temperature, batch_size=batch_size, max_concurrent=max_concurrent, display=verbose
            )
            results.append(
                BenchmarkMetrics(
                    ai_model=model,
                    accuracy=metrics.accuracy,
                    levenshtein_similarity=metrics.levenshtein_similarity,
                    jaccard_similarity=metrics.jaccard_similarity,
                    false_positive_rate=metrics.false_positive_rate,
                    false_negative_rate=metrics.false_negative_rate,
                    mismatched_value_rate=metrics.mismatched_value_rate,
                )
            )

        if print:
            display_benchmark_metrics(results)

        return results

    def update_annotations(
        self,
        json_schema: dict[str, Any] | Path | str,
        old_dataset_path: str | Path,
        new_dataset_path: str | Path,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0.0,
        batch_size: int = 5,
        max_concurrent: int = 3,
    ) -> None:
        """Update annotations in a JSONL file using a new model.

        Args:
            json_schema: The JSON schema for validation
            old_dataset_path: Path to the JSONL file to update
            new_dataset_path: Path for saving updated annotations
            model: The model to use for new annotations
            temperature: Model temperature (0-1)
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
        """
        json_schema = load_json_schema(json_schema)
        assert_valid_model_extraction(model)
        schema_obj = Schema(json_schema=json_schema)

        # Initialize appropriate client
        client, provider = self._initialize_model_client(model)

        # Read all lines from the JSONL file
        with open(old_dataset_path, 'r') as f:
            lines = [json.loads(line) for line in f]

        updated_entries = []
        total_batches = (len(lines) + batch_size - 1) // batch_size

        batch_pbar = tqdm(total=total_batches, desc="Processing batches", position=0)

        def process_entry(entry: dict) -> dict:
            messages = entry['messages']
            system_message, user_messages, assistant_messages = separate_messages(messages)
            system_and_user_messages = messages[:-1]

            previous_annotation_message: ChatCompletionUiformMessage = {
                "role": "user",
                "content": "Here is an old annotation using a different schema. Use it as a reference to update the annotation: " + messages[-1]['content'],
            }

            string_json = self._get_model_completion(
                client=client,
                provider=provider,
                model=model,
                temperature=temperature,
                messages=schema_obj.messages + user_messages + [previous_annotation_message],
                schema_obj=schema_obj,
            )

            return {"messages": system_and_user_messages + [{"role": "assistant", "content": string_json}]}

        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            futures = []
            for batch_idx in range(0, len(lines), batch_size):
                batch = lines[batch_idx : batch_idx + batch_size]

                batch_futures = [executor.submit(process_entry, entry | {"line_number": batch_idx * batch_size + i}) for i, entry in enumerate(batch)]
                futures.extend(batch_futures)

                for future in batch_futures:
                    try:
                        result = future.result()
                        updated_entries.append(result)
                    except Exception as e:
                        print(f"Error processing example: {e}")

                batch_pbar.update(1)
                time.sleep(1)

        batch_pbar.close()

        with open(new_dataset_path, 'w') as f:
            for entry in updated_entries:
                f.write(json.dumps(entry) + '\n')

    #########################
    ##### BATCH METHODS #####
    #########################

    def save_batch_annotate_requests(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: list[Path | str | IOBase],
        batch_requests_path: Path,
        model: str = "gpt-4o-mini",
        temperature: float = 0.0,
        modality: Modality = "native",
    ) -> None:
        """Create a JSONL file containing requests for OpenAI batch processing API.

        Args:
            json_schema: The JSON schema for validation
            documents: List of documents to process
            batch_requests_path: Output path for the JSONL requests file
            model: The model to use for processing
            temperature: Model temperature (0-1)
            modality: The modality to use for document processing
        """
        loaded_json_schema = load_json_schema(json_schema)
        schema_obj = Schema(json_schema=loaded_json_schema)
        assert_valid_model_extraction(model)

        with open(batch_requests_path, 'w', encoding='utf-8') as f:
            for i, doc in tqdm(enumerate(documents)):
                # Create document messages
                doc_msg = self._client.documents.create_messages(
                    document=doc,
                    modality=modality,
                )

                # Construct the request object
                request: BatchJSONL = {
                    "custom_id": f"request-{i}",
                    "method": "POST",
                    "url": "/v1/chat/completions",
                    "body": {
                        "model": model,
                        "messages": schema_obj.openai_messages + doc_msg.openai_messages,
                        "temperature": temperature,
                        "response_format": {"type": "json_schema", "json_schema": {"name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True}},
                    },
                }

                # Write the request as a JSON line
                f.write(json.dumps(request) + '\n')

    def save_batch_update_annotation_requests(
        self,
        json_schema: dict[str, Any] | Path | str,
        old_dataset_path: str | Path,
        batch_requests_path: str | Path,
        model: str = "gpt-4o-mini",
        temperature: float = 0.0,
    ) -> None:
        """Create a JSONL file containing requests for OpenAI batch processing API to update annotations.

        Args:
            json_schema: The JSON schema for validation
            old_dataset_path: Path to the JSONL file to update
            batch_requests_path: Output path for the updated JSONL file
            model: The model to use for processing
            temperature: Model temperature (0-1)
            modality: The modality to use for document processing
        """
        loaded_json_schema = load_json_schema(json_schema)
        schema_obj = Schema(json_schema=loaded_json_schema)

        # Read existing annotations
        with open(old_dataset_path, 'r') as f:
            entries = [json.loads(line) for line in f]

        # Create new JSONL with update requests
        with open(batch_requests_path, 'w', encoding='utf-8') as f:
            for i, entry in enumerate(entries):
                existing_messages = entry['messages']
                system_and_user_messages = existing_messages[:-1]

                previous_annotation_message: ChatCompletionMessageParam = {
                    "role": "user",
                    "content": "Here is an old annotation using a different schema. Use it as a reference to update the annotation: " + existing_messages[-1]['content'],
                }

                # Construct the request object
                response_format: BatchJSONLResponseFormat = {
                    "type": "json_schema",
                    "json_schema": {"name": schema_obj.id, "schema": schema_obj.inference_json_schema, "strict": True},
                }

                body: BatchJSONLBody = {
                    "model": model,
                    "messages": schema_obj.openai_messages + system_and_user_messages + [previous_annotation_message],
                    "temperature": temperature,
                    "response_format": response_format,
                }

                request: BatchJSONL = {"custom_id": f"request-{i}", "method": "POST", "url": "/v1/chat/completions", "body": body}

                # Write the request as a JSON line
                f.write(json.dumps(request) + '\n')

    def build_dataset_from_batch_results(
        self,
        batch_requests_path: str | Path,
        batch_results_path: str | Path,
        dataset_results_path: str | Path,
    ) -> None:
        with open(batch_requests_path, 'r') as f:
            input_lines: list[BatchJSONL] = [json.loads(line) for line in f]
        with open(batch_results_path, 'r') as f:
            batch_results_lines: list[BatchJSONLResponse] = [json.loads(line) for line in f]

        assert len(input_lines) == len(batch_results_lines), "Input and batch results must have the same number of lines"

        for input_line, batch_result in zip(input_lines, batch_results_lines):
            messages = input_line['body']['messages']

            # Filter out messages containing the old annotation reference to remove messages that come from "update annotation"
            if isinstance(messages[-1].get('content'), str):
                if re.search(r'Here is an old annotation using a different schema\. Use it as a reference to update the annotation:', str(messages[-1].get('content', ''))):
                    print("found keyword")
                    input_line['body']['messages'] = messages[:-1]

            input_line['body']['messages'].append(batch_result['response']['body']['choices'][0]['message'])

        with open(dataset_results_path, 'w') as f:
            for input_line in input_lines:
                f.write(json.dumps({'messages': input_line['body']['messages']}) + '\n')

        print(f"Dataset saved to {dataset_results_path}")

    #############################
    ##### END BATCH METHODS #####
    #############################


class AsyncDatasets(AsyncAPIResource, BaseDatasetsMixin):
    """Asynchronous wrapper for Datasets using thread execution."""

    async def save(
        self,
        json_schema: dict[str, Any] | Path | str,
        pairs_paths: list[dict[str, Path | str]],
        jsonl_path: Path | str,
        modality: Modality = "native",
    ) -> None:
        json_schema = load_json_schema(json_schema)
        training_set = []

        for pair_paths in tqdm(pairs_paths):
            document_message = await self._client.documents.create_messages(document=pair_paths['document_fpath'], modality=modality)

            with open(pair_paths['annotation_fpath'], 'r') as f:
                annotation = json.loads(f.read())
            assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

            training_set.append({"messages": document_message.messages + [assistant_message]})

        self._dump_training_set(training_set, jsonl_path)

    async def annotate(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: list[Path | str | IOBase],
        jsonl_path: str | Path,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0.0,
        batch_size: int = 5,
        max_concurrent: int = 3,
        root_dir: Path = Path("annotations"),
        modality: Modality = "native",
    ) -> None:
        json_schema = load_json_schema(json_schema)
        assert_valid_model_extraction(model)
        """
        Generate annotations from document files or in-memory documents
        and create a JSONL training set in one go.

        Args:
            json_schema: The JSON schema for validation
            documents: list of documents, each can be a Path/str or an IOBase object
            jsonl_path: Output path for the JSONL training file
            model: The model to use for processing
            temperature: Model temperature (0-1)
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
            root_dir: Where to store the per-document JSON annotations
        """

        async def process_example(doc: Path | str | IOBase, semaphore: asyncio.Semaphore) -> dict[str, Any]:
            """
            Process a single document (either a file path or an in-memory file-like object).
            Returns a dict with pointers to the original doc and the stored annotation JSON.
            """
            if isinstance(doc, (str, Path)):
                # Handle path or string
                doc_path = Path(doc)
                if not doc_path.is_file():
                    raise ValueError(f"Invalid file path: {doc_path}")

                # Extract results
                async with semaphore:
                    result = await self._client.documents.extractions.parse(
                        json_schema=json_schema,
                        document=doc_path,  # pass the actual Path to .extract
                        model=model,
                        temperature=temperature,
                        modality=modality,
                    )
                if result.choices[0].message.content is None:
                    print(f"Failed to extract content from {doc_path}")
                    return {"document_fpath": str(doc_path), "annotation_fpath": None}
                # Generate a unique filename for the annotation
                hash_str = hashlib.md5(doc_path.as_posix().encode()).hexdigest()
                annotation_path = Path(root_dir) / f"annotations_{hash_str}.json"

                annotation_path.parent.mkdir(parents=True, exist_ok=True)

                with open(annotation_path, 'w', encoding='utf-8') as f:
                    json.dump(result.choices[0].message.content, f, ensure_ascii=False, indent=2)

                return {"document_fpath": str(doc_path), "annotation_fpath": str(annotation_path)}

            elif isinstance(doc, IO):
                # Handle in-memory file-like object
                # 1) Read file content (but be careful with read pointer!)
                file_bytes = doc.read()

                # 2) Attempt to get a name; default to "uploaded_file" if none
                doc_name = getattr(doc, "name", "uploaded_file")

                # 3) Reset the file pointer if you plan to reuse `doc`
                #    (optional, depending on how you're using it)
                doc.seek(0)

                # 4) Call extract with the same doc object
                async with semaphore:
                    result = await self._client.documents.extractions.parse(
                        json_schema=json_schema,
                        document=doc,  # pass the IO object directly
                        model=model,
                        temperature=temperature,
                        modality=modality,
                    )

                if result.choices[0].message.content is None:
                    print(f"Failed to extract content from {doc_name}")
                    return {"document_fpath": doc_name, "annotation_fpath": None}

                # 5) Create a unique hash from the content
                hash_str = hashlib.md5(file_bytes).hexdigest()
                annotation_path = Path(root_dir) / f"annotations_{hash_str}.json"

                annotation_path.parent.mkdir(parents=True, exist_ok=True)

                with open(annotation_path, 'w', encoding='utf-8') as f:
                    json.dump(result.choices[0].message.content, f, ensure_ascii=False, indent=2)

                return {
                    "document_fpath": doc_name,  # or "in_memory_file"
                    "annotation_fpath": str(annotation_path),
                }

            else:
                raise ValueError(f"Unsupported document type: {type(doc)}")

        # Make sure output directory exists
        Path(root_dir).mkdir(parents=True, exist_ok=True)

        pairs_paths: list[dict[str, Path | str]] = []
        futures = []
        semaphore = asyncio.Semaphore(max_concurrent)
        # Split documents into batches
        for batch in tqdm([documents[i : i + batch_size] for i in range(0, len(documents), batch_size)], desc="Processing batches"):
            # Submit batch of tasks
            for doc in batch:
                futures.append(process_example(doc, semaphore))
            pairs_paths = await asyncio.gather(*futures)

        # Generate final training set from all results
        await self.save(json_schema=json_schema, pairs_paths=pairs_paths, jsonl_path=jsonl_path)

    async def benchmark(self, **kwargs: Any) -> None:
        """Benchmark model performance on a test dataset.

        Args:
            json_schema: JSON schema defining the expected data structure
            jsonl_path: Path to the JSONL file containing test examples
            model: The AI model to use for benchmarking
            temperature: Model temperature setting (0-1)

        Raises:
            NotImplementedError: This method is not implemented yet
        """

        # TODO

        raise NotImplementedError("Benchmarking is not implemented yet")

    async def filter(self, **kwargs: Any) -> None:
        """Filter examples from a JSONL file based on specified parameters.

        Args:
            json_schema: JSON schema defining the data structure
            jsonl_path: Path to the JSONL file to filter
            output_path: Optional path for the filtered output
            inplace: Whether to modify the file in place
            filter_parameters: Parameters to filter examples by (e.g., {"confidence": 0.8})

        Note:
            Filter parameters can include:
            - Number of tokens
            - Modality
            - Other custom parameters
        """
        raise NotImplementedError("Filtering is not implemented yet")

    async def print(self, jsonl_path: Path, output_path: Path = Path("annotations")) -> None:
        """Print a summary of the contents and statistics of a JSONL file.

        This method analyzes the JSONL file and displays various metrics and statistics
        about the dataset contents.

        Args:
            jsonl_path: Path to the JSONL file to analyze
            output_path: Directory where to save any generated reports
        """
        raise NotImplementedError("Printing is not implemented yet")

    async def stitch(self, **kwargs: Any) -> None:
        """Stitch annotations from a list of MIMEData objects into a single MIMEData object.

        This method combines multiple MIMEData annotations into a single object to avoid
        nested list structures (list[list[MIMEData]]) and maintain a simpler list[MIMEData] structure.

        Args:
            json_schema: The JSON schema for validation
            jsonl_path: Path to the JSONL file
            output_path: Optional path for the output file
            inplace: Whether to modify the file in place
            filter_parameters: Optional parameters for filtering
            modality: The modality to use for processing
        """

        raise NotImplementedError("Stitching is not implemented yet")
