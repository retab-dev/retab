# Contributing to Retab

> "You cannot use 1 model that drives everything, at least today" ~ not a Retab team member, but definitely a Retab enthusiast

We're thrilled you want to contribute to Retab. Every contribution—whether fixing bugs, adding features, or improving documentation—makes Retab better for everyone.

## Getting Started

### Before You Dive In

1. **Check existing issues** or open a new one to start a discussion
2. **Read [Retab's documentation](https://docs.retab.com/)** and core [concepts](https://docs.retab.com/core-concepts/Core_Concepts)

## Contribution Opportunities

### Cookbook

- Add examples and tutorials in our [cookbook folder](https://github.com/retab-dev/retab/tree/main/cookbook).
- Add community integrations
- Fix typos

### Code Improvements

- Implement new document processing strategies
- Improve test coverage

### Performance Enhancements

- Add benchmarks
- Improve memory usage

### New Features

Look for [issues](https://github.com/retab-dev/retab/issues)

## PR Process

```bash
# Fork and clone the repository
git clone https://github.com/retab-dev/retab.git
cd retab
```

### 1. Branch Naming

- `feature/description` for new features
- `fix/description` for bug fixes
- etc.

### 2. Commit Messages

Write clear, descriptive commit messages:

```
feat: add batch preprocessing to Retab

- Implement batch_process method
- Add tests for batch processing
- Update documentation
```

### 3. Dependencies

- Core dependencies go in `project.dependencies`
- Optional features go in `project.optional-dependencies`
- Development tools go in the `dev` optional dependency group

### 4. Code Review

- All PRs need at least one review
- Maintainers will review for:
  - Code quality
  - Test coverage
  - Performance impact
  - Documentation completeness

## Getting Help

- **Questions?** Open an issue or ask in [Discord](https://discord.com/invite/vc5tWRPqag)
- **Bugs?** Open an issue or report in [Discord](https://discord.com/invite/vc5tWRPqag)
- **Chat?** Join our [Discord](https://discord.com/invite/vc5tWRPqag)!
- **Email?** Contact [contact@retab.com](mailto:contact@retab.com)

## Thank You 🩷

Every contribution helps make Retab better! We appreciate your time and effort in helping make Retab the most performant document processing framework!