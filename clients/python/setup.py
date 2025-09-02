from setuptools import find_packages, setup  # type: ignore

# Read requirements.txt and use it for the install_requires parameter
with open("requirements.txt") as f:
    requirements_list = f.read().splitlines()

setup(
    name="retab",
    version="0.0.62",
    author="Retab",
    author_email="contact@retab.com",
    description="Retab official python library",
    long_description=open("README.md").read(),
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
