import json
import os

import requests
from mcp.server.fastmcp import FastMCP
from openai.types.model import Model

from retab import Retab

# Initialize FastMCP server
mcp = FastMCP("Retab")

# Constants
USER_AGENT = "retab-app/1.0"


@mcp.tool()
async def generate_schema(list_file_download_urls: list[str], model: str = "gpt-5.2") -> str:
    """Generates an Retab promptified JSON Schema from input files.

    This function takes a list of file URLs, downloads their contents, and generates
    a JSON Schema using Retab's schema generation capabilities. The schema includes
    standard JSON Schema fields and can be used for document extraction.

    Args:
        list_file_download_urls: A list of URLs pointing to files that need to be processed.
        model: The model identifier to use for processing. Defaults to "gpt-5.2".
                Use get_retab_available_models() to see available options.

    Returns:
        str: A JSON Schema as a string (from json.dumps with no formatting or indentations)

    Raises:
        Exception: If there's an error downloading the files or processing them with Retab.
    """
    retab_client = Retab(api_key=api_key, base_url=base_url)
    try:
        # Download the file
        list_binary_data = []
        for file_download_url in list_file_download_urls:
            response = requests.get(file_download_url)
            list_binary_data.append(response.content)
        # FIrst, generate the schema from the file
        response = retab_client.schemas.generate(documents=list_binary_data, model=model, modality="native")

        json_schema = response.json_schema

        # # Now the final json schema
        final_schema = json_schema.copy()

        # Return the processing results
        return json.dumps(final_schema)

    except Exception as e:
        return f"Failed to process PDF file due to an error: {str(e)}"


@mcp.tool()
async def get_retab_available_models() -> str:
    """Retrieves a list of available models from the Retab API.

    This function queries the Retab API to get a list of all available models
    that can be used with other Retab functions.

    Returns:
        str: A comma-separated string listing all available model IDs.
             Format: "Available models: model1, model2, ..."

    Raises:
        Exception: If there's an error connecting to the Retab API or retrieving
                  the model list.
    """
    retab_client = Retab(api_key=api_key, base_url=base_url)
    try:
        available_models: list[Model] = retab_client.models.list()
    except Exception as e:
        return "Failed to get available models due to an error: " + str(e)

    return "Available models: " + ", ".join([model.id for model in available_models])


if __name__ == "__main__":
    api_key = os.getenv("RETAB_API_KEY")
    base_url = os.getenv("RETAB_BASE_URL", "https://api.retab.com")
    if not api_key:
        raise ValueError("RETAB_API_KEY is not set")
    try:
        retab_client = Retab(api_key=api_key, base_url=base_url)
        models = retab_client.models.list()
    except Exception as e:
        print("Failed to create Retab client due to an error: " + str(e))
        exit(1)

    # Initialize and run the server
    mcp.run(transport="stdio")
