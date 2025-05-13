
class AsyncEvals(AsyncAPIResource, EvalsMixin):
    """Async Evals API wrapper"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.documents = AsyncDocuments(self._client)
        self.document = AsyncDocument(self._client)
        self.iterations = AsyncIterations(self._client)

    async def create(self, name: str, json_schema: Dict[str, Any], project_id: str) -> Experiment:
        """
        Create a new evaluation.

        Args:
            name: The name of the evaluation
            json_schema: The JSON schema for the evaluation
            project_id: The project ID to associate with the evaluation

        Returns:
            Experiment: The created evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(name, json_schema, project_id)
        response = await self._client._prepared_request(request)
        return Experiment(**response)

    async def get(self, id: str) -> Experiment:
        """
        Get an evaluation by ID.

        Args:
            id: The ID of the evaluation to retrieve

        Returns:
            Experiment: The evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        response = await self._client._prepared_request(request)
        return Experiment(**response)

    async def update(self, id: str, name: str) -> Experiment:
        """
        Update an evaluation.

        Args:
            id: The ID of the evaluation to update
            name: The new name for the evaluation

        Returns:
            Experiment: The updated evaluation
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(id, name)
        response = await self._client._prepared_request(request)
        return Experiment(**response)

    async def list(self, project_id: str) -> List[Experiment]:
        """
        List evaluations for a project.

        Args:
            project_id: The project ID to list evaluations for

        Returns:
            List[Experiment]: List of evaluations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(project_id)
        response = await self._client._prepared_request(request)
        return [Experiment(**item) for item in response.get("data", [])]

    async def delete(self, id: str) -> DeleteResponse:
        """
        Delete an evaluation.

        Args:
            id: The ID of the evaluation to delete

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return await self._client._prepared_request(request)


class AsyncDocuments(AsyncAPIResource, DocumentsMixin):
    """Async Documents API wrapper for evaluations"""

    async def import_jsonl(self, eval_id: str, path: str) -> Experiment:
        """
        Import documents from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Experiment: The updated experiment with imported documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        response = await self._client._prepared_request(request)
        return Experiment(**response)

    async def save_to_jsonl(self, eval_id: str, path: str) -> ExportResponse:
        """
        Save documents to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            ExportResponse: The response containing success status and path
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def create(self, eval_id: str, document: str, ground_truth: Dict[str, Any]) -> ExperimentDocument:
        """
        Create a document for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            document: The document file path or content
            ground_truth: The ground truth for the document

        Returns:
            ExperimentDocument: The created document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, document, ground_truth)
        response = await self._client._prepared_request(request)
        return ExperimentDocument(**response)

    async def list(self, eval_id: str, filename: Optional[str] = None) -> List[ExperimentDocument]:
        """
        List documents for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            filename: Optional filename to filter by

        Returns:
            List[ExperimentDocument]: List of documents
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, filename)
        response = await self._client._prepared_request(request)
        return [ExperimentDocument(**item) for item in response.get("data", [])]


class AsyncDocument(AsyncAPIResource, DocumentMixin):
    """Async Document API wrapper for individual document operations"""

    async def update(self, eval_id: str, id: str, ground_truth: Dict[str, Any]) -> ExperimentDocument:
        """
        Update a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document
            ground_truth: The ground truth for the document

        Returns:
            ExperimentDocument: The updated document
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(eval_id, id, ground_truth)
        response = await self._client._prepared_request(request)
        return ExperimentDocument(**response)

    async def delete(self, eval_id: str, id: str) -> DeleteResponse:
        """
        Delete a document.

        Args:
            eval_id: The ID of the evaluation
            id: The ID of the document

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(eval_id, id)
        return await self._client._prepared_request(request)


class AsyncIterations(AsyncAPIResource, IterationsMixin):
    """Async Iterations API wrapper for evaluations"""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.distances = AsyncDistances(self._client)

    async def import_jsonl(self, eval_id: str, path: str) -> Experiment:
        """
        Import iterations from a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to the JSONL file

        Returns:
            Experiment: The updated experiment with imported iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_import_jsonl(eval_id, path)
        response = await self._client._prepared_request(request)
        return Experiment(**response)

    async def save_to_jsonl(self, eval_id: str, path: str) -> ExportResponse:
        """
        Save iterations to a JSONL file.

        Args:
            eval_id: The ID of the evaluation
            path: The path to save the JSONL file

        Returns:
            ExportResponse: The response containing success status and path
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_save_to_jsonl(eval_id, path)
        return await self._client._prepared_request(request)

    async def get(self, id: str) -> Iteration:
        """
        Get an iteration by ID.

        Args:
            id: The ID of the iteration

        Returns:
            Iteration: The iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(id)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def list(self, eval_id: str, model: Optional[str] = None) -> List[Iteration]:
        """
        List iterations for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            model: Optional model to filter by

        Returns:
            List[Iteration]: List of iterations
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_list(eval_id, model)
        response = await self._client._prepared_request(request)
        return [Iteration(**item) for item in response.get("data", [])]

    async def create(self, eval_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Iteration:
        """
        Create a new iteration for an evaluation.

        Args:
            eval_id: The ID of the evaluation
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Iteration: The created iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_create(eval_id, json_schema, model, temperature, image_settings)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def update(self, iteration_id: str, json_schema: Dict[str, Any], model: str, temperature: float = 0.0, image_settings: Optional[Dict[str, Any]] = None) -> Iteration:
        """
        Update an iteration.

        Args:
            iteration_id: The ID of the iteration
            json_schema: The JSON schema for the iteration
            model: The model to use for the iteration
            temperature: The temperature to use for the model
            image_settings: Optional image settings

        Returns:
            Iteration: The updated iteration
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_update(iteration_id, json_schema, model, temperature, image_settings)
        response = await self._client._prepared_request(request)
        return Iteration(**response)

    async def delete(self, id: str) -> DeleteResponse:
        """
        Delete an iteration.

        Args:
            id: The ID of the iteration

        Returns:
            DeleteResponse: The response containing success status and ID
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_delete(id)
        return await self._client._prepared_request(request)


class AsyncDistances(AsyncAPIResource, DistancesMixin):
    """Async Distances API wrapper for iterations"""

    async def get(self, iteration_id: str, document_id: str) -> MetricResult:
        """
        Get distances for a document in an iteration.

        Args:
            iteration_id: The ID of the iteration
            document_id: The ID of the document

        Returns:
            MetricResult: The distances
        Raises:
            HTTPException if the request fails
        """
        request = self.prepare_get(iteration_id, document_id)
        response = await self._client._prepared_request(request)
        return MetricResult(**response)

