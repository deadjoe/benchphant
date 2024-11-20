# BenchPhant

[![Go Report Card](https://goreportcard.com/badge/github.com/deadjoe/benchphant?v=2)](https://goreportcard.com/report/github.com/deadjoe/benchphant)
[![GoDoc](https://pkg.go.dev/badge/github.com/deadjoe/benchphant)](https://pkg.go.dev/github.com/deadjoe/benchphant)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://github.com/deadjoe/benchphant/actions/workflows/ci.yml/badge.svg)](https://github.com/deadjoe/benchphant/actions)
[![codecov](https://codecov.io/gh/deadjoe/benchphant/branch/main/graph/badge.svg)](https://codecov.io/gh/deadjoe/benchphant)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deadjoe/benchphant)](https://github.com/deadjoe/benchphant)
[![Release](https://img.shields.io/github/v/release/deadjoe/benchphant)](https://github.com/deadjoe/benchphant/releases)

<div align="center">
  <img src="docs/assets/logo.png" alt="Benchphant Logo" width="200">
  <p><strong>Modern Database Performance Testing Made Easy</strong></p>
</div>

Benchphant is a modern, user-friendly database stress testing tool that supports MySQL (including MySQL clusters) and PostgreSQL databases. It provides a beautiful web interface for configuring and monitoring database performance tests, inspired by industry-standard tools like sysbench and TPC-C.

## Features

- Modern Web UI - Beautiful, responsive interface with real-time monitoring
- Security First - Built-in authentication and encryption
- Rich Visualizations - Interactive charts and comprehensive reports
- Theme Support - Light/Dark modes for comfortable viewing
- Multi-DB Support - MySQL, PostgreSQL, and more coming soon
- Advanced Metrics - QPS, latency percentiles, resource usage
- Detailed Reports - Test history and comparative analysis
- Local Storage - SQLite-based configuration and results storage
- Plugin System - Extensible architecture for custom workloads
- Docker Ready - Easy deployment with Docker and Docker Compose
- Comprehensive Testing - Extensive test coverage and static analysis
- Code Quality - Enforced by golangci-lint and continuous integration

## Quick Start

### Prerequisites

- Go 1.21 or later
- Node.js 16 or later
- npm or yarn

### Installation

#### Using Go

```bash
go install github.com/deadjoe/benchphant@latest
```

#### Using Docker

```bash
docker pull deadjoe/benchphant
docker run -p 8080:8080 deadjoe/benchphant
```

### Running Locally

```bash
benchphant
```

The application will automatically open in your default web browser at `http://localhost:8080`.

Default credentials:
- Username: `bench`
- Password: `bench`

## Development

### Backend Development

```bash
# Get dependencies
go mod download

# Run tests with coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run linters
go vet ./...
golangci-lint run
```

### Frontend Development

```bash
# Install dependencies
cd web
npm install

# Start development server
npm run dev

# Run tests
npm test

# Build for production
npm run build
```

## Documentation

- [Getting Started Guide](docs/getting-started.md)
- [API Documentation](docs/api.md)
- [Contributing Guide](CONTRIBUTING.md)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [sysbench](https://github.com/akopytov/sysbench) - Inspiration for benchmark workloads
- [TPC](http://www.tpc.org/) - Industry standard database benchmarks
- [Vue.js](https://vuejs.org/) - Frontend framework
- [Chart.js](https://www.chartjs.org/) - Beautiful charts
- [Tailwind CSS](https://tailwindcss.com/) - Styling
