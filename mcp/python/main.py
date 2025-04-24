import json
import os

import requests
from mcp.server.fastmcp import FastMCP
from openai.types.model import Model

from uiform import UiForm

# Initialize FastMCP server
mcp = FastMCP("UiForm")

# Constants
USER_AGENT = "uiform-app/1.0"


@mcp.tool()
async def generate_schema(list_file_download_urls: list[str], model: str = "gpt-4o") -> str:
    """Generates an UiForm promptified JSON Schema from input files.

    This function takes a list of file URLs, downloads their contents, and generates
    a JSON Schema using UiForm's schema generation capabilities. The schema includes
    standard JSON Schema fields and can be used for document extraction.

    Args:
        list_file_download_urls: A list of URLs pointing to files that need to be processed.
        model: The model identifier to use for processing. Defaults to "gpt-4o".
                Use get_uiform_available_models() to see available options.

    Returns:
        str: A JSON Schema as a string (from json.dumps with no formatting or indentations)

    Raises:
        Exception: If there's an error downloading the files or processing them with UiForm.
    """
    uiform_client = UiForm(api_key=api_key, base_url=base_url)
    try:
        # Download the file
        list_binary_data = []
        for file_download_url in list_file_download_urls:
            response = requests.get(file_download_url)
            list_binary_data.append(response.content)
        # FIrst, generate the schema from the file
        response = uiform_client.schemas.generate(documents=list_binary_data, model=model, modality="native")

        raw_schema = response.json_schema

        # # Now the final json schema
        final_schema = raw_schema.copy()

        # Return the processing results
        return json.dumps(final_schema)

    except Exception as e:
        return f"Failed to process PDF file due to an error: {str(e)}"


@mcp.tool()
async def get_uiform_available_models() -> str:
    """Retrieves a list of available models from the UiForm API.

    This function queries the UiForm API to get a list of all available models
    that can be used with other UiForm functions.

    Returns:
        str: A comma-separated string listing all available model IDs.
             Format: "Available models: model1, model2, ..."

    Raises:
        Exception: If there's an error connecting to the UiForm API or retrieving
                  the model list.
    """
    uiform_client = UiForm(api_key=api_key, base_url=base_url)
    try:
        available_models: list[Model] = uiform_client.models.list()
    except Exception as e:
        return "Failed to get available models due to an error: " + str(e)

    return "Available models: " + ", ".join([model.id for model in available_models])


if __name__ == "__main__":
    api_key = os.getenv("UIFORM_API_KEY")
    base_url = os.getenv("UIFORM_BASE_URL", "https://api.uiform.com")
    if not api_key:
        raise ValueError("UIFORM_API_KEY is not set")
    try:
        uiform_client = UiForm(api_key=api_key, base_url=base_url)
        models = uiform_client.models.list()
    except Exception as e:
        print("Failed to create UiForm client due to an error: " + str(e))
        exit(1)

    # Initialize and run the server
    mcp.run(transport='stdio')
