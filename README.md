# Benchphant

[![Go Report Card](https://goreportcard.com/badge/github.com/deadjoe/benchphant)](https://goreportcard.com/report/github.com/deadjoe/benchphant)
[![GoDoc](https://pkg.go.dev/badge/github.com/deadjoe/benchphant)](https://pkg.go.dev/github.com/deadjoe/benchphant)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://github.com/deadjoe/benchphant/actions/workflows/test.yml/badge.svg)](https://github.com/deadjoe/benchphant/actions)
[![codecov](https://codecov.io/gh/deadjoe/benchphant/branch/main/graph/badge.svg)](https://codecov.io/gh/deadjoe/benchphant)

Benchphant is a modern, user-friendly database stress testing tool that supports MySQL (including MySQL clusters) and PostgreSQL databases. It provides a command-line interface for configuring and monitoring database performance tests, inspired by industry-standard tools like sysbench and TPC-C.

## Features

- Multi-DB Support - MySQL, PostgreSQL support
- Advanced Metrics - QPS, latency percentiles, resource usage
- Detailed Reports - Test history and comparative analysis
- Plugin System - Extensible architecture for custom workloads
- Comprehensive Testing - Extensive test coverage and static analysis
- Code Quality - Enforced by golangci-lint and continuous integration

## Quick Start

### Prerequisites

- Go 1.21 or later

### Installation

```bash
go install github.com/deadjoe/benchphant@latest
```

### Usage

1. Run a simple OLTP test:

```bash
benchphant oltp --host localhost --port 3306 --user root --password secret --database test
```

2. Run a TPC-C test:

```bash
benchphant tpcc --host localhost --port 3306 --user root --password secret --database test --warehouses 10
```

## Documentation

For detailed documentation, please visit our [Wiki](https://github.com/deadjoe/benchphant/wiki).

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) for details on how to submit pull requests, report issues, and contribute to the project.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
