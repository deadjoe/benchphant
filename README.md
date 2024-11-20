# Benchphant

[![Go Report Card](https://goreportcard.com/badge/github.com/joe/benchphant)](https://goreportcard.com/report/github.com/joe/benchphant)
[![GoDoc](https://pkg.go.dev/badge/github.com/joe/benchphant)](https://pkg.go.dev/github.com/joe/benchphant)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://github.com/joe/benchphant/actions/workflows/test.yml/badge.svg)](https://github.com/joe/benchphant/actions)
[![Lint Status](https://github.com/joe/benchphant/actions/workflows/lint.yml/badge.svg)](https://github.com/joe/benchphant/actions)
[![codecov](https://codecov.io/gh/joe/benchphant/branch/main/graph/badge.svg)](https://codecov.io/gh/joe/benchphant)
[![Go Version](https://img.shields.io/github/go-mod/go-version/joe/benchphant)](https://github.com/joe/benchphant)
[![Release](https://img.shields.io/github/v/release/joe/benchphant)](https://github.com/joe/benchphant/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/joe/benchphant)](https://hub.docker.com/r/joe/benchphant)

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
go install github.com/joe/benchphant@latest
```

#### Using Docker

```bash
docker pull joe/benchphant
docker run -p 8080:8080 joe/benchphant
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
- [Architecture Overview](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Development Guide](docs/development.md)
- [Plugin Development](docs/plugins.md)
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

## Contact

- GitHub Issues: For bug reports and feature requests
- Email: benchphant@example.com
- Twitter: [@benchphant](https://twitter.com/benchphant)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=joe/benchphant&type=Date)](https://star-history.com/#joe/benchphant&Date)
