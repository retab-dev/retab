from setuptools import setup, find_packages # type: ignore

# Read requirements.txt and use it for the install_requires parameter
with open('requirements.txt') as f:
    requirements_list = f.read().splitlines()

setup(
    name='uiform',
    version='0.0.1',
    author='Sacha Ichbiah',
    author_email='sacha@getcube.ai',
    description='Cube official python library',
    long_description=open('README.md').read(),
    long_description_content_type='text/markdown',
    url='https://github.com/cubelogistics/cubeblock',
    project_urls={
        'Team website': 'https://github.com/cubelogistics'
    },
    classifiers=[
        'Programming Language :: Python :: 3',
        'License :: OSI Approved :: MIT License',
        'Operating System :: POSIX :: Linux',
        'Operating System :: MacOS',
        'Intended Audience :: Science/Research'
    ],
    packages=find_packages(),
    python_requires='>=3.6',
    install_requires=requirements_list,

    include_package_data=True,
)
