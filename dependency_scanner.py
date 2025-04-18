import os
import re

EXAMPLES_DIR = "examples"
PROJECT_ROOT = "/Users/victorsoto/DocumentsLocal/uiform-open-source"

def find_all_python_files(base_dir):
    py_files = []
    for root, _, files in os.walk(base_dir):
        for file in files:
            if file.endswith(".py"):
                py_files.append(os.path.join(root, file))
    return py_files

def scan_for_imports(py_files, target_files):
    usage_map = {target: [] for target in target_files}
    relative_paths = [os.path.relpath(tf, PROJECT_ROOT).replace("/", ".").replace(".py", "") for tf in target_files]

    for file in py_files:
        with open(file, "r", encoding="utf-8") as f:
            content = f.read()

        for target, module_path in zip(target_files, relative_paths):
            if (
                re.search(rf"\bimport {re.escape(module_path)}\b", content) or
                re.search(rf"\bfrom {re.escape(module_path)}\b", content)
            ):
                usage_map[target].append(file)

    return usage_map

if __name__ == "__main__":
    print("🔍 Scanning for dependencies...")
    all_py_files = find_all_python_files(PROJECT_ROOT)
    example_files = find_all_python_files(os.path.join(PROJECT_ROOT, EXAMPLES_DIR))

    usage = scan_for_imports(all_py_files, example_files)

    for target, used_by in usage.items():
        if used_by:
            print(f"📌 {target} is used by:")
            for ref in used_by:
                print(f"   - {ref}")
        else:
            print(f"🧹 {target} is NOT referenced anywhere else ✅")
