# Neo4j Multi-tenant Proxy

A high-performance multi-tenant proxy for Neo4j databases, enabling tenant isolation in open-source Neo4j deployments.

## Features

- **Multi-tenant Support**: Route connections to different Neo4j backends based on tenant identification
- **Bolt Protocol Compatible**: Full support for Neo4j's Bolt protocol (versions 1-4)
- **Flexible Tenant Routing**: Multiple strategies for tenant identification (username-based, database-based, metadata-based)
- **High Performance**: Transparent connection proxying with minimal latency overhead
- **Production Ready**: Comprehensive testing, CI/CD pipeline, graceful shutdown
- **Easy Configuration**: JSON-based configuration with hot-reload support

## Quick Start

### Prerequisites

- Go 1.25 or later
- Neo4j backend instances (3.0+)

### Installation

```bash
git clone https://github.com/your-org/neo4j-proxy
cd neo4j-proxy
make build
```

### Configuration

1. Copy the example configuration:
   ```bash
   cp config.example.json config.json
   ```

2. Edit `config.json` to match your Neo4j backends:
   ```json
   {
     "proxy_port": 7687,
     "tenants": {
       "tenant1": {
         "host": "neo4j-backend-1.example.com",
         "port": 7687
       },
       "tenant2": {
         "host": "neo4j-backend-2.example.com", 
         "port": 7687
       }
     }
   }
   ```

3. Start the proxy:
   ```bash
   CONFIG_FILE=config.json ./neo4j-proxy
   ```

### Usage

Connect to the proxy using standard Neo4j drivers, specifying tenant in username:

```python
from neo4j import GraphDatabase

# Connect to tenant1
driver = GraphDatabase.driver("bolt://localhost:7687", auth=("tenant1@myuser", "password"))

# Connect to tenant2  
driver = GraphDatabase.driver("bolt://localhost:7687", auth=("tenant2@myuser", "password"))
```

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

### Development Mode

```bash
make dev
```

See `make help` for all available commands.

## Architecture

The proxy uses a layered architecture with clear separation of concerns:

- **Connection Handler**: Manages client connections and proxy lifecycle
- **Protocol Parser**: Handles Bolt protocol handshake and message parsing  
- **Tenant Router**: Routes connections based on tenant identification
- **Backend Manager**: Manages connections to Neo4j backends

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.