import asyncio
from typing import IO, Any, Optional, Literal
from concurrent.futures import ThreadPoolExecutor
import hashlib
import time
import json
from pathlib import Path
from tqdm import tqdm
from io import IOBase

from .._utils.json_schema import load_json_schema
from .._utils.ai_model import assert_valid_model_extraction
from .._utils.display import display_metrics, process_dataset_and_compute_metrics, Metrics
from ..types.modalities import Modality
from ..types.documents.create_messages import DocumentMessage, ChatCompletionUiformMessage
from ..types.schemas.object import Schema
from .._resource import SyncAPIResource, AsyncAPIResource
from .benchmarking import analyze_comparison_metrics, ComparisonMetrics, ExtractionAnalysis, compare_dicts, plot_comparison_metrics



class BaseDatasetsMixin:

    def _dump_training_set(self, training_set: list[dict[str, Any]], jsonl_path: Path | str) -> None:
        with open(jsonl_path, 'w', encoding='utf-8') as file:
            for entry in training_set:
                file.write(json.dumps(entry) + '\n')

class Datasets(SyncAPIResource, BaseDatasetsMixin):
    """Datasets API wrapper"""

    # TODO : Maybe at some point we could add some visualization methods... but the multimodality makes it hard...
    # client.datasets.plot.tsne()...
    # client.datasets.plot.umap()...

    def save(
        self,
        json_schema: dict[str, Any] | Path | str,
        pairs_paths: list[dict[str, Path | str]],
        jsonl_path: Path | str,
        text_operations: Optional[dict[str, Any]] = None,
        modality: Modality = "native",
        messages: list[ChatCompletionUiformMessage] = [],
    ) -> None:
        """Save document-annotation pairs to a JSONL training set.

        Args:
            json_schema: The JSON schema for validation, can be a dict, Path, or string
            pairs_paths: List of dictionaries containing document and annotation file paths
            jsonl_path: Output path for the JSONL training file
            text_operations: Optional context for prompting
            modality: The modality to use for document processing ("native" by default)
            messages: List of additional chat messages to include
        """
        json_schema = load_json_schema(json_schema)
        schema_obj = Schema(
            json_schema=json_schema
        )

        training_set = []

        for pair_paths in tqdm(pairs_paths):
            document_message = self._client.documents.create_messages(
                document=pair_paths['document_fpath'], 
                modality=modality, 
                text_operations=text_operations
            )
            
            # Use context manager to properly close the file
            with open(pair_paths['annotation_fpath'], 'r') as f:
                annotation = json.loads(f.read())
            assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

            # Add the complete message set as an entry
            training_set.append({"messages": document_message.messages + messages + [assistant_message]})
    
            

        self._dump_training_set(training_set, jsonl_path)


    def update_annotations(
            self, 
            json_schema: dict[str, Any] | Path | str,
            jsonl_path: Path, 
            save_path: Path,
            model: str = "gpt-4o-2024-08-06",
            temperature: float = 0,
            batch_size: int = 5,
            max_concurrent: int = 3,
            modality: Modality = "native",
            ) -> None:
        """Update annotations in a JSONL file using a new model.
        
        Args:
            json_schema: The JSON schema for validation
            jsonl_path: Path to the JSONL file to update
            model: The model to use for new annotations
            temperature: Model temperature (0-1)
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
            modality: The modality to use for processing
        """
        json_schema = load_json_schema(json_schema)
        assert_valid_model_extraction(model)
        
        # Read all lines from the JSONL file
        with open(jsonl_path, 'r') as f:
            lines = [json.loads(line) for line in f]
        
        updated_entries = []
        total_batches = (len(lines) + batch_size - 1) // batch_size  # Calculate total number of batches
        
        # Create main progress bar for batches
        batch_pbar = tqdm(total=total_batches, desc="Processing batches", position=0)
        
        def process_entry(entry: dict) -> dict:
            messages = entry['messages']
            system_and_user_messages = messages[:-1]

            previous_annotation_message = {
                "role": "user",
                "content": "Here is an old annotation using a different schema. Use it as a reference to update the annotation: " + messages[-1]['content']
            }
            
            result = self._client.documents.extractions.parse(
                json_schema=json_schema,
                document=None,
                model=model,
                temperature=temperature,
                messages=system_and_user_messages + [previous_annotation_message],
                modality=modality,
            )
            
            if result.choices[0].message.parsed is None:
                print(f"Failed to update annotation for entry")
                return entry
            
            new_assistant_message = {
                "role": "assistant",
                "content": result.choices[0].message.parsed.model_dump_json() 
            }
            return {"messages": system_and_user_messages + [new_assistant_message]}
        
        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            futures = []
            # Split entries into batches
            for batch_idx in range(0, len(lines), batch_size):
                batch = lines[batch_idx:batch_idx + batch_size]
                
                # Submit batch of tasks
                batch_futures = [executor.submit(process_entry, entry) for entry in batch]
                futures.extend(batch_futures)
                
                # Wait for batch to complete
                for future in batch_futures:
                    updated_entries.append(future.result())
                
                batch_pbar.update(1)  # Update batch progress
                time.sleep(1)  # Rate limiting between batches
        
        batch_pbar.close()  # Close batch progress bar
        
        # Write updated entries back to file
        with open(save_path, 'w') as f:
            for entry in updated_entries:
                f.write(json.dumps(entry) + '\n')


    def annotate(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: list[Path | str | IOBase],
        jsonl_path: Path,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
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
            text_operations: Optional context for prompting
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
                # Handle path or string
                doc_path = Path(doc)
                if not doc_path.is_file():
                    raise ValueError(f"Invalid file path: {doc_path}")

                # Extract results
                result = self._client.documents.extractions.parse(
                    json_schema=json_schema,
                    document=doc_path,  # pass the actual Path to .extract
                    text_operations=text_operations,
                    model=model,
                    temperature=temperature,
                    messages=messages,
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
                result = self._client.documents.extractions.parse(
                    json_schema=json_schema,
                    document=doc,  # pass the IO object directly
                    text_operations=text_operations,
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
        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            futures = []
            # Split documents into batches
            for batch in tqdm([documents[i : i + batch_size] for i in range(0, len(documents), batch_size)], desc="Processing batches"):
                # Submit batch of tasks
                batch_futures = [executor.submit(process_example, doc) for doc in batch]
                futures.extend(batch_futures)

                # Wait for batch to finish (rate limit)
                for future in batch_futures:
                    pair = future.result()
                    pairs_paths.append(pair)
                time.sleep(1)  # simple rate-limiting pause between batches

        # Generate final training set from all results
        self.save(json_schema=json_schema, text_operations=text_operations, pairs_paths=pairs_paths, jsonl_path=jsonl_path)

    def benchmark(
        self,
        json_schema: dict[str, Any] | Path | str,
        jsonl_path: Path,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0, 
        text_operations: Optional[dict[str, Any]] = None,
        batch_size: int = 5,
        max_concurrent: int = 3,
        #metric: Literal["Levenstein", "Jaccard", "Accuracy"] = "Accuracy", # TODO
    ) -> ComparisonMetrics:
        
        """Benchmark model performance on a test dataset.

        Args:
            json_schema: JSON schema defining the expected data structure
            jsonl_path: Path to the JSONL file containing test examples
            model: The model to use for benchmarking
            temperature: Model temperature setting (0-1)
            text_operations: Optional context with regex instructions
            batch_size: Number of examples to process in each batch
            max_concurrent: Maximum number of concurrent API calls
        """
        json_schema = load_json_schema(json_schema)
        assert_valid_model_extraction(model)

        # Read all lines from the JSONL file
        with open(jsonl_path, 'r') as f:
            lines = [json.loads(line) for line in f]
        
        extraction_analyses: list[ExtractionAnalysis] = []
        total_batches = (len(lines) + batch_size - 1) // batch_size

        # Create main progress bar for batches
        batch_pbar = tqdm(total=total_batches, desc="Processing batches", position=0)

        def process_example(jsonline: dict) -> ExtractionAnalysis:
            messages = jsonline['messages']
            ground_truth = json.loads(messages[-1]['content'])
            inference_messages = messages[:-1]

            result = self._client.documents.extractions.parse(
                json_schema=json_schema,
                document=None,
                text_operations=text_operations,
                model=model,
                temperature=temperature,
                messages=inference_messages,
            )

            assert result.choices[0].message.parsed is not None
            prediction = result.choices[0].message.parsed.model_dump()

            return ExtractionAnalysis(
                ground_truth=ground_truth,
                prediction=prediction,
            )

        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            # Split entries into batches
            for batch_idx in range(0, len(lines), batch_size):
                batch = lines[batch_idx:batch_idx + batch_size]
                
                # Submit and process batch
                futures = [executor.submit(process_example, entry) for entry in batch]
                for future in futures:
                    extraction_analyses.append(future.result())

                batch_pbar.update(1)
                time.sleep(1)  # Rate limiting between batches

        batch_pbar.close()

        # Analyze error patterns across all examples
        analysis = analyze_comparison_metrics(extraction_analyses)
        plot_comparison_metrics(
            analysis=analysis, 
            top_n=10
        )

        return analysis


    def pprint(self, jsonl_path: Path, output_path: Path = Path("annotations")) -> Metrics:
        """Print a summary of the contents and statistics of a JSONL file.

        This method analyzes the JSONL file and displays various metrics and statistics
        about the dataset contents.

        Inspired from : https://gist.github.com/nmwsharp/54d04af87872a4988809f128e1a1d233

        Args:
            jsonl_path: Path to the JSONL file to analyze
            output_path: Directory where to save any generated reports
        """

        computed_metrics = process_dataset_and_compute_metrics(jsonl_path)
        display_metrics(computed_metrics)
        return computed_metrics

    def stich_and_save(
        self,
        json_schema: dict[str, Any] | Path | str,
        pairs_paths: list[dict[str, Path | str | list[Path | str] | list[str] | list[Path]]],
        jsonl_path: Path | str,
        text_operations: Optional[dict[str, Any]] = None,
        modality: Modality = "native",
        messages: list[ChatCompletionUiformMessage] = [],
    ) -> None:        
        
        """Stitch multiple documents and their annotations into a JSONL training set.

        This method processes and combines multiple documents into messages, creating document-annotation
        pairs that are saved to a JSONL file. Each document is processed according to the specified
        modality and combined with its corresponding annotation.


        Args:
            json_schema: The JSON schema for validation, can be a dict, Path, or string
            pairs_paths: List of dictionaries containing document and annotation file paths
            jsonl_path: Output path for the JSONL training file
            text_operations: Optional context for prompting
            modality: The modality to use for document processing ("native" by default)
            messages: List of additional chat messages to include
        """

        json_schema = load_json_schema(json_schema)
        training_set = []

        for pair_paths in tqdm(pairs_paths):
            document_messages: list[ChatCompletionUiformMessage] = []
            
            if isinstance(pair_paths['document_fpath'], str) or isinstance(pair_paths['document_fpath'], Path):
                document_message = self._client.documents.create_messages(
                    document=pair_paths['document_fpath'], 
                    modality=modality, 
                    text_operations=text_operations
                )
                document_messages.extend(document_message.messages)

            else: 
                assert isinstance(pair_paths['document_fpath'], list)
                for document_fpath in pair_paths['document_fpath']:
                    document_message = self._client.documents.create_messages(
                        document=document_fpath, 
                        modality=modality, 
                        text_operations=text_operations
                    )
                    document_messages.extend(document_message.messages)
            
            # Use context manager to properly close the file
            assert isinstance(pair_paths['annotation_fpath'], Path) or isinstance(pair_paths['annotation_fpath'], str)
            with open(pair_paths['annotation_fpath'], 'r') as f:
                annotation = json.loads(f.read())
            assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

            # Add the complete message set as an entry
            training_set.append({"messages": document_messages + messages + [assistant_message]})

        self._dump_training_set(training_set, jsonl_path)

   

class AsyncDatasets(AsyncAPIResource, BaseDatasetsMixin):
    """Asynchronous wrapper for Datasets using thread execution."""

    async def save(
        self,
        json_schema: dict[str, Any] | Path | str,
        pairs_paths: list[dict[str, Path | str]],
        jsonl_path: Path | str,
        text_operations: Optional[dict[str, Any]] = None,
        modality: Modality = "native",
        messages: list[ChatCompletionUiformMessage] = [],
    ) -> None:
        json_schema = load_json_schema(json_schema)
        training_set = []

        for pair_paths in tqdm(pairs_paths):
            document_message = await self._client.documents.create_messages(document=pair_paths['document_fpath'], modality=modality, text_operations=text_operations)
            
            with open(pair_paths['annotation_fpath'], 'r') as f:
                annotation = json.loads(f.read())
            assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

            training_set.append({"messages": document_message.messages + messages + [assistant_message]})

        self._dump_training_set(training_set, jsonl_path)

    async def annotate(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: list[Path | str | IOBase],
        jsonl_path: Path,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
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
            text_operations: Optional context for prompting
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
                        text_operations=text_operations,
                        model=model,
                        temperature=temperature,
                        messages=messages,
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
                        text_operations=text_operations,
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
        await self.save(json_schema=json_schema, text_operations=text_operations, pairs_paths=pairs_paths, jsonl_path=jsonl_path)

    async def benchmark(self, **kwargs: Any) -> None:
        #json_schema: dict[str, Any], jsonl_path: Path, text_operations: Optional[dict[str, Any]], model: str, temperature: float

        """Benchmark model performance on a test dataset.

        Args:
            json_schema: JSON schema defining the expected data structure
            jsonl_path: Path to the JSONL file containing test examples
            text_operations: Optional context with regex instructions or other metadata
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
            text_operations: Optional context with processing instructions
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
            text_operations: Optional context with processing instructions
            output_path: Optional path for the output file
            inplace: Whether to modify the file in place
            filter_parameters: Optional parameters for filtering
            modality: The modality to use for processing
        """

        raise NotImplementedError("Stitching is not implemented yet")
