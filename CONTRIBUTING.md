# Contributing to Benchphant

First off, thank you for considering contributing to Benchphant! It's people like you that make Benchphant such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* Use a clear and descriptive title
* Describe the exact steps which reproduce the problem
* Provide specific examples to demonstrate the steps
* Describe the behavior you observed after following the steps
* Explain which behavior you expected to see instead and why
* Include screenshots and animated GIFs if possible

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* Use a clear and descriptive title
* Provide a step-by-step description of the suggested enhancement
* Provide specific examples to demonstrate the steps
* Describe the current behavior and explain which behavior you expected to see instead
* Explain why this enhancement would be useful

### Pull Requests

* Fill in the required template
* Do not include issue numbers in the PR title
* Include screenshots and animated GIFs in your pull request whenever possible
* Follow the Go and JavaScript styleguides
* Include thoughtfully-worded, well-structured tests
* Document new code
* End all files with a newline

## Development Process

1. Fork the repo
2. Create a new branch for each feature/fix
3. Make your changes
4. Run tests and linters
5. Submit a Pull Request

### Setting up development environment

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/benchphant.git

# Add upstream remote
git remote add upstream https://github.com/deadjoe/benchphant.git

# Install dependencies
go mod download
cd web && npm install

# Run tests and linters
go test -race -coverprofile=coverage.txt -covermode=atomic ./...
golangci-lint run
cd web && npm test
```

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

### Go Styleguide

* Follow the official Go style guide
* Use `gofmt -s` for simplified code
* Run `golangci-lint` before committing
* Write descriptive comments for exported items
* Include tests with good coverage for new code
* Handle errors appropriately
* Use meaningful variable names
* Follow the project's error handling patterns

### JavaScript Styleguide

* Use ES6+ features
* Follow Vue.js style guide
* Use meaningful variable names
* Write unit tests for new code

## Additional Notes

### Issue and Pull Request Labels

* `bug`: Something isn't working
* `enhancement`: New feature or request
* `good first issue`: Good for newcomers
* `help wanted`: Extra attention is needed
* `documentation`: Improvements or additions to documentation

## Recognition

Contributors are recognized in our README.md file. We value every contribution, no matter how small!
