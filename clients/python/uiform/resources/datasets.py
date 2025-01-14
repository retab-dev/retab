from typing import IO, Any, Optional
import hashlib
import time
import json
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
from tqdm import tqdm

from .._utils.json_schema import load_json_schema
from .._utils.ai_model import assert_valid_model_extraction
from .._resource import SyncAPIResource, WrappedAsyncAPIResource
from ..types.modalities import Modality
from ..types.documents.create_messages import ChatCompletionUiformMessage
from io import IOBase

class Datasets(SyncAPIResource):
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
        training_set = []

        for elmt in tqdm(pairs_paths):
            print(elmt)
            doc_obj = self._client.documents.create_messages(
                document=elmt['document_fpath'], 
                modality=modality, 
                text_operations=text_operations)
            system_and_user_messages = doc_obj.messages + messages if doc_obj.messages else messages

            # Use context manager to properly close the file
            with open(elmt["annotation_fpath"], 'r') as f:
                annotation = json.loads(f.read())

            assistant_message = {"role": "assistant", "content": json.dumps(annotation, ensure_ascii=False, indent=2)}

            # Add the complete message set as an entry
            training_set.append({"messages": system_and_user_messages + [assistant_message]})

        with open(jsonl_path, 'w', encoding='utf-8') as file:
            for entry in training_set:
                file.write(json.dumps(entry) + '\n')

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

    def benchmark(self, **kwargs: Any) -> None:
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

    def filter(self, **kwargs: Any) -> None:
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
        #json_schema: dict[str, Any] | Path | str,
        #jsonl_path: Path,
        #text_operations: Optional[dict[str, Any]] = None,
        #output_path: Optional[Path] = None,
        #inplace: bool = False,
        #filter_parameters: Optional[dict[str, Any]] = None,

        #json_schema = load_json_schema(json_schema)

        # filter parameters:
        # num tokens
        # modality

        # TODO

        raise NotImplementedError("Filtering is not implemented yet")

    def print(self, jsonl_path: Path, output_path: Path = Path("annotations")) -> None:
        """Print a summary of the contents and statistics of a JSONL file.

        This method analyzes the JSONL file and displays various metrics and statistics
        about the dataset contents.

        Args:
            jsonl_path: Path to the JSONL file to analyze
            output_path: Directory where to save any generated reports
        """

        # client.datasets.statistics(json_schema, jsonl_path, text_operations) -> this is a valid name as well for the method.
        # inspiration: https://x.com/nmwsharp/status/1629205292096557056/photo/1
        # https://gist.github.com/nmwsharp/54d04af87872a4988809f128e1a1d233

        from uiform.display import display_metrics, process_dataset_and_compute_metrics

        computed_metrics = process_dataset_and_compute_metrics(jsonl_path)
        display_metrics(computed_metrics)
        raise NotImplementedError("Printing is not implemented yet")

    def stitch(self, **kwargs: Any) -> None:
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
        #json_schema: dict[str, Any] | Path | str,
        #jsonl_path: Path,
        #text_operations: Optional[dict[str, Any]] = None,
        #output_path: Optional[Path] = None,
        #inplace: bool = False,
        #filter_parameters: Optional[dict[str, Any]] = None,
        #modality: Modality = "native",
        #json_schema = load_json_schema(json_schema)

        # TODO

        raise NotImplementedError("Stitching is not implemented yet")


class AsyncDatasets(WrappedAsyncAPIResource):
    """Asynchronous wrapper for Datasets using thread execution."""

    def __init__(self, client: Any, max_workers: Optional[int] = None) -> None:
        super().__init__(client, Datasets(client=client), max_workers=max_workers)

    async def save(
        self,
        json_schema: dict[str, Any] | Path | str,
        pairs_paths: list[dict[str, Path | str]],
        jsonl_path: Path | str,
        text_operations: Optional[dict[str, Any]] = None,
        is_lazy: bool = False,
    ) -> None:
        return await self.__getattr__("save")(json_schema, pairs_paths, jsonl_path, text_operations, is_lazy)

    async def annotate(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: list[Path | str | IO[bytes]],
        jsonl_path: Path,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        batch_size: int = 5,
        max_concurrent: int = 3,
        root_dir: Path = Path("annotations"),
        modality: Modality = "native",
    ) -> Any:
        return await self.__getattr__("annotate")(json_schema, documents, jsonl_path, text_operations, model, temperature, batch_size, max_concurrent, root_dir, modality)

    async def benchmark(self, json_schema: dict[str, Any], jsonl_path: Path, text_operations: Optional[dict[str, Any]], model: str, temperature: float) -> Any:
        return await self.__getattr__("benchmark")(json_schema, jsonl_path, text_operations, model, temperature)

    async def filter(
        self,
        json_schema: dict[str, Any] | Path | str,
        jsonl_path: Path,
        text_operations: Optional[dict[str, Any]] = None,
        output_path: Optional[Path] = None,
        inplace: bool = False,
        filter_parameters: Optional[dict[str, Any]] = None,
    ) -> Any:
        return await self.__getattr__("filter")(json_schema, jsonl_path, text_operations, output_path, inplace, filter_parameters)

    async def print(self, jsonl_path: Path, output_path: Path = Path("annotations")) -> Any:
        return await self.__getattr__("print")(jsonl_path, output_path)

    async def stitch(
        self,
        json_schema: dict[str, Any] | Path | str,
        jsonl_path: Path,
        text_operations: Optional[dict[str, Any]] = None,
        output_path: Optional[Path] = None,
        inplace: bool = False,
        filter_parameters: Optional[dict[str, Any]] = None,
        modality: Modality = "native",
    ) -> Any:
        return await self.__getattr__("stitch")(json_schema, jsonl_path, text_operations, output_path, inplace, filter_parameters, modality)
