from pathlib import Path

from setuptools import find_packages, setup  # type: ignore

# Resolve data files relative to this file, not the caller's cwd: `pip install
# .` (or an sdist build) can run from a different directory, where a bare
# Path("requirements.txt") raises FileNotFoundError.
HERE = Path(__file__).parent.resolve()


def load_requirements() -> list[str]:
    requirements_path = HERE / "requirements.txt"
    requirements: list[str] = []
    for raw_line in requirements_path.read_text().splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        requirements.append(line)
    return requirements


requirements_list = load_requirements()

setup(
    name="retab",
    version="0.0.154",
    author="Retab",
    author_email="contact@retab.com",
    description="Retab official python library",
    long_description=(HERE / "README.md").read_text(),
    long_description_content_type="text/markdown",
    url="https://github.com/retab-dev/retab",
    project_urls={"Team website": "https://retab.com"},
    classifiers=[
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "License :: OSI Approved :: MIT License",
        "Operating System :: POSIX :: Linux",
        "Operating System :: MacOS",
        "Intended Audience :: Science/Research",
    ],
    packages=find_packages(),
    python_requires=">=3.11",
    install_requires=requirements_list,
    include_package_data=True,
    package_data={"retab": ["**/*.yaml"]},
)
