import os
import ast
import nbformat

# Directories to scan
BASE_DIR = "/Users/victorsoto/DocumentsLocal/uiform-open-source"
TARGET_DIRS = ["examples", "notebooks"]

def analyze_python_file(path):
    summary = {
        "type": "python",
        "imports": [],
        "functions": [],
        "classes": [],
        "providers": [],
        "doc": None
    }
    try:
        with open(path, "r", encoding="utf-8") as f:
            source = f.read()
            tree = ast.parse(source)
            summary["doc"] = ast.get_docstring(tree)

            for node in ast.walk(tree):
                if isinstance(node, ast.Import):
                    for n in node.names:
                        summary["imports"].append(n.name)
                elif isinstance(node, ast.ImportFrom):
                    summary["imports"].append(node.module)
                elif isinstance(node, ast.FunctionDef):
                    summary["functions"].append(node.name)
                elif isinstance(node, ast.ClassDef):
                    summary["classes"].append(node.name)

            providers = ["openai", "anthropic", "bedrock", "groq", "cohere", "mistral"]
            for provider in providers:
                if provider in source.lower():
                    summary["providers"].append(provider)

    except Exception as e:
        summary["error"] = str(e)

    return summary

def analyze_notebook_file(path):
    summary = {
        "type": "notebook",
        "imports": [],
        "functions": [],
        "classes": [],
        "providers": [],
        "doc": None
    }
    try:
        with open(path, "r", encoding="utf-8") as f:
            nb = nbformat.read(f, as_version=4)
            code_cells = [cell['source'] for cell in nb.cells if cell['cell_type'] == 'code']
            full_code = "\n".join(code_cells)
            tree = ast.parse(full_code)
            summary["doc"] = ast.get_docstring(tree)

            for node in ast.walk(tree):
                if isinstance(node, ast.Import):
                    for n in node.names:
                        summary["imports"].append(n.name)
                elif isinstance(node, ast.ImportFrom):
                    summary["imports"].append(node.module)
                elif isinstance(node, ast.FunctionDef):
                    summary["functions"].append(node.name)
                elif isinstance(node, ast.ClassDef):
                    summary["classes"].append(node.name)

            providers = ["openai", "anthropic", "bedrock", "groq", "cohere", "mistral"]
            for provider in providers:
                if provider in full_code.lower():
                    summary["providers"].append(provider)

    except Exception as e:
        summary["error"] = str(e)

    return summary

def scan_project():
    results = {}
    for folder in TARGET_DIRS:
        folder_path = os.path.join(BASE_DIR, folder)
        for root, _, files in os.walk(folder_path):
            for file in files:
                if file.endswith(".py"):
                    path = os.path.join(root, file)
                    results[path] = analyze_python_file(path)
                elif file.endswith(".ipynb"):
                    path = os.path.join(root, file)
                    results[path] = analyze_notebook_file(path)
    return results

if __name__ == "__main__":
    import json
    result = scan_project()
    output_path = "summary_output.json"
    with open(output_path, "w", encoding="utf-8") as out:
        json.dump(result, out, indent=2)
    print(f"✔️ Summary saved to {output_path}")
