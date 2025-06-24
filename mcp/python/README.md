## How to include this tool in your project:


1. Add the following to your claude_desktop_config.json file (usually located in `~/Library/Application\ Support/Claude/claude_desktop_config.json`)

```json
{
    "mcpServers": {
      "retab-dev2": {
        "command": "docker",
        "args": [
          "exec",
          "-i",
          "-e",
          "RETAB_API_KEY",
          "retab-devcontainer-1",
          "/bin/bash",
          "-i",
          "-c",
          "uv --directory /workspaces/retab/open-source/retab/mcp/python run main.py"
        ],
        "env": {
          "RETAB_API_KEY": "[Your Retab API Key]"
        }
      }
    }
}
```

That's it, now you can use the `retab` tool within Claude Desktop App.

I set it up using a devcontainer, but you can run it locally as well. For example:

```json
{
    "mcpServers": {
      "retab-dev2": {
        "command": "uv",
        "args": [
          "--directory",
          "/path/to/retab/mcp/python",
          "run",
          "main.py"
        ],
        "env": {
          "RETAB_API_KEY": "[Your Retab API Key]"
        }
      }
    }
}
```



## Some important notes:

- The `retab` tool only accept downloadable URLs when generating a schema
- You can use http://localhost links to test the tool locally, those might not work when MCP allows remote connections.





