from pathlib import Path

from setuptools import find_packages, setup  # type: ignore


def load_requirements() -> list[str]:
    requirements_path = Path("requirements.txt")
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
    version="0.0.132",
    author="Retab",
    author_email="contact@retab.com",
    description="Retab official python library",
    long_description=Path("README.md").read_text(),
    long_description_content_type="text/markdown",
    url="https://github.com/retab-dev/retab",
    project_urls={"Team website": "https://retab.com"},
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: POSIX :: Linux",
        "Operating System :: MacOS",
        "Intended Audience :: Science/Research",
    ],
    packages=find_packages(),
    python_requires=">=3.6",
    install_requires=requirements_list,
    include_package_data=True,
    package_data={"retab": ["**/*.yaml"]},
)
