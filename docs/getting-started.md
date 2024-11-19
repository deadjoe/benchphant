# Getting Started with Benchphant

This guide will help you get up and running with Benchphant quickly.

## Installation

### Prerequisites

Before installing Benchphant, ensure you have:

- Go 1.21 or later
- Node.js 16 or later
- npm or yarn
- Git (optional, for development)

### Installation Methods

#### Using Go

```bash
go install github.com/deadjoe/benchphant@latest
```

#### Using Docker

```bash
docker pull deadjoe/benchphant
docker run -p 8080:8080 deadjoe/benchphant
```

#### Building from Source

```bash
# Clone the repository
git clone https://github.com/deadjoe/benchphant.git
cd benchphant

# Build backend
go build -o benchphant cmd/benchphant/main.go

# Build frontend
cd web
npm install
npm run build
```

## First Steps

1. **Start the Application**

```bash
benchphant
```

2. **Access the Web Interface**

Open your browser and navigate to `http://localhost:8080`

Default credentials:
- Username: `bench`
- Password: `bench`

3. **Add Your First Database Connection**

- Click "Add Connection" in the dashboard
- Fill in your database details:
  - Name: A friendly name for your connection
  - Host: Database host address
  - Port: Database port
  - Username: Database user
  - Password: Database password
  - Database: Database name
  - Type: MySQL or PostgreSQL

4. **Run Your First Benchmark**

- Select your connection from the dashboard
- Click "New Benchmark"
- Configure your test:
  - Duration: How long to run the test
  - Threads: Number of concurrent connections
  - Workload Type: Type of operations to perform
- Click "Start" to begin the benchmark

## Understanding Results

Benchphant provides several metrics:

- **QPS (Queries Per Second)**: The number of queries executed per second
- **Latency**: Response time statistics
  - Average
  - P50 (Median)
  - P95 (95th percentile)
  - P99 (99th percentile)
- **Error Rate**: Percentage of failed queries
- **Resource Usage**: CPU, Memory, and I/O statistics

## Next Steps

- Read the [API Documentation](api.md) to integrate Benchphant with your tools
- Check out [Development Guide](development.md) to contribute
- Join our community on [Discord](#) or [Slack](#)
- Follow us on [Twitter](https://twitter.com/benchphant)

## Troubleshooting

### Common Issues

1. **Cannot connect to database**
   - Check if the database is running
   - Verify connection details
   - Ensure firewall rules allow connection

2. **Performance issues**
   - Check system resources
   - Verify network connectivity
   - Adjust thread count

3. **Web interface not loading**
   - Clear browser cache
   - Check console for errors
   - Verify port availability

### Getting Help

- Check our [FAQ](faq.md)
- Search [GitHub Issues](https://github.com/deadjoe/benchphant/issues)
- Join our [Community Forums](#)
- Contact support at support@benchphant.com

## Best Practices

1. **Start Small**
   - Begin with a small number of threads
   - Use short duration tests first
   - Gradually increase load

2. **Monitor Resources**
   - Watch system CPU and memory
   - Monitor network bandwidth
   - Check database metrics

3. **Regular Testing**
   - Schedule regular benchmarks
   - Compare results over time
   - Document changes and impacts

## Security Considerations

1. **Credentials**
   - Use dedicated benchmark user accounts
   - Limit database permissions
   - Rotate passwords regularly

2. **Network**
   - Use SSL/TLS when possible
   - Consider VPN for remote testing
   - Monitor network traffic

3. **Data**
   - Use test databases only
   - Avoid production data
   - Clean up test data regularly
