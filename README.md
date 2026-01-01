# finparser

Convert human-friendly CSV to Qlik Sense loadable CSV format with support for multiple currencies and automatic conversion to Russian Rubles.

## Overview

`finparser` is a command-line tool that processes financial transaction data from CSV files. It parses human-readable purchase descriptions and converts them into structured CSV format suitable for analysis tools like Qlik Sense.

## Features

- **Multi-currency support**: USD ($), EUR (€), Belarusian Ruble (Br), and Armenian Dram (֏)
- **Automatic currency conversion** to Russian Rubles using CBR (Central Bank of Russia) rates
- **Flexible price expressions**: Simple numbers, arithmetic operations, currency notations
- **Category mapping**: Automatic categorization with transport category consolidation
- **Person/category parsing**: Support for person-specific transactions
- **Date format flexibility**: Configurable date formats

## Supported Currencies

| Currency | Symbol | Code | Example |
|----------|---------|------|---------|
| US Dollar | $ | USD | `$15`, `$10.50=750` |
| Euro | € | EUR | `€12`, `€8.25=900` |
| Belarusian Ruble | Br | BYN | `Br20`, `Br15=400` |
| Armenian Dram | ֏ | AMD | `֏2000`, `֏1500=300` |

## Input Format

The input CSV should have the following structure:
```csv
Date,Items
15.12.2023,"Food - bread (50), Transport - bus (30)"
16.12.2023,"John/Food - groceries ($25), Mary/Clothes - shirt (€15), Anna/Gifts - flowers (֏2000)"
```

## Purchase Description Format

Each purchase item follows the pattern: `[Person/]Category[ - Name] (Price)`

### Examples

- `Food (100)` → Person: "Общие", Category: "food", Name: "food", Price: 100
- `Food - bread (50)` → Person: "Общие", Category: "food", Name: "bread", Price: 50
- `John/Food - groceries (200)` → Person: "john", Category: "food", Name: "groceries", Price: 200
- `Mary|Clothes - dress ($45)` → Person: "mary", Category: "clothes", Name: "dress", Price: ~1200 RUB
- `Anna/Gifts - flowers (֏2000)` → Person: "anna", Category: "gifts", Name: "flowers", Price: ~400 RUB

### Price Expression Formats

1. **Simple numbers**: `(100)`, `(0)`
2. **Arithmetic expressions**: `(50+25)`, `(2*300)`, `(1000/4)`
3. **Currency with conversion**: `($15)`, `(€20)`, `(Br10)`, `(֏2000)`
4. **Currency with explicit rate**: `($10=750)`, `(€15=1300)`, `(Br5=135)`, `(֏1500=300)`

### Category Auto-mapping

Transport-related categories are automatically consolidated:
- автобус, трамвай, троллейбус, маршрутка, метро, электричка → транспорт
- интернет → связь

## Usage

```bash
# Basic usage with default date format (DD.MM.YYYY)
cat input.csv | go run finparser.go > output.csv

# Custom date format
cat input.csv | go run finparser.go -df "01/02/2006" > output.csv
```

### Command Line Options

- `-df string`: Date format in Go time format (default: "02.01.2006")

## Output Format

The tool outputs CSV with the following columns:
```csv
Date,Person,Category,Name,Price
15.12.2023,общие,food,bread,50
16.12.2023,john,food,groceries,1200
16.12.2023,mary,clothes,shirt,1300
```

## Examples

### Input CSV
```csv
Date,Items
15.12.2023,"Продукты - хлеб (50), Маша/автобус (30), Одежда ($25)"
16.12.2023,"Кафе - кофе (€8), Аптека - лекарства (Br15), Подарки (֏2000), Бензин (2*500)"
```

### Command
```bash
cat example.csv | go run finparser.go
```

### Output
```csv
15.12.2023,общие,продукты,хлеб,50
15.12.2023,маша,транспорт,автобус,30
15.12.2023,общие,одежда,одежда,1350
16.12.2023,общие,кафе,кофе,870
16.12.2023,общие,аптека,лекарства,405
16.12.2023,общие,подарки,подарки,410
16.12.2023,общие,бензин,бензин,1000
```

## Currency Conversion

- **Current rates**: Uses live CBR exchange rates for transactions without specific dates
- **Historical rates**: Fetches historical rates for dated transactions
- **Fallback behavior**: If historical rates are unavailable (e.g., BYN before 2016), conversion returns 0
- **AMD support**: Armenian Dram has both current and historical rates available in CBR API
- **Explicit rates**: When using `Currency=Amount` format, uses the specified rate instead of CBR

## Requirements

- Go 1.25 or later

## Building

### Local Build
```bash
go build ./...
```

### Docker Build

The project supports building Docker images for multiple architectures using Docker Buildx.

#### Prerequisites
```bash
# Ensure Docker Buildx is available (included in Docker Desktop)
docker buildx version

# One-time setup for multiarch builds
make setup-buildx
```

#### Single Architecture Build
```bash
# Build for current architecture
make build

# Build and tag with version
make release version=1.0.0

# Push to registry
make push version=1.0.0
```

#### Multi-Architecture Build
```bash
# Build for multiple architectures (linux/amd64, linux/arm64)
make build-multiarch

# Build and push multiarch images to registry
make build-multiarch-push

# Build and push with version tag
make push-multiarch version=1.0.0
```

#### Direct Docker Commands
```bash
# Build multiarch without Make
docker buildx build --platform linux/amd64,linux/arm64 --tag dddpaul/finparser:latest .

# Build and push multiarch
docker buildx build --platform linux/amd64,linux/arm64 --tag dddpaul/finparser:latest --push .

# Build specific architecture
docker buildx build --platform linux/arm64 --tag dddpaul/finparser:arm64 --load .
```

#### Available Make Targets
- `make build` - Build single-arch Docker image for current platform
- `make build-multiarch` - Build multiarch image (amd64, arm64) - cache only
- `make build-multiarch-push` - Build and push multiarch image to registry
- `make setup-buildx` - Set up Docker Buildx for multiarch builds (one-time)
- `make release version=X.Y.Z` - Tag single-arch image with version
- `make release-multiarch version=X.Y.Z` - Build and push multiarch with version tag
- `make push version=X.Y.Z` - Push single-arch images with version
- `make push-multiarch version=X.Y.Z` - Push multiarch images with version

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestParsePriceExpr ./...
```

## Docker Support

The application supports multi-architecture Docker images built with Go 1.25 for:
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM64/AArch64 - Apple Silicon, AWS Graviton, etc.)

### Docker Hub
Pre-built multiarch images are available at: `dddpaul/finparser`

```bash
# Pull and run latest version (automatically selects correct architecture)
docker pull dddpaul/finparser:latest
docker run -i dddpaul/finparser:latest < input.csv > output.csv

# Pull specific version
docker pull dddpaul/finparser:1.0.0

# Run with specific platform (if needed)
docker run --platform linux/amd64 -i dddpaul/finparser:latest < input.csv > output.csv
```

### Local Testing with Docker
```bash
# Test with sample data
echo "Date,Items
01.01.2024,\"Food (100), Transport (\$25), Кафе (€8)\"" | docker run -i dddpaul/finparser:latest

# Expected output:
# 01.01.2024,общие,food,food,100
# 01.01.2024,общие,транспорт,транспорт,1350
# 01.01.2024,общие,кафе,кафе,870
```

### Build Architecture Notes
- Uses `golang:1.25-alpine` image with separated dependency download for reliable compilation
- Final runtime image uses `alpine:latest` with CA certificates and timezone data
- Cross-compilation handled automatically by Docker Buildx with proper build arguments
- Build process optimized for Docker layer caching (go.mod/go.sum copied separately)
- Compatible with both local Docker builds and CI/CD multiarch builds

## Dependencies

- `github.com/soniah/evaler` - Mathematical expression evaluation
- `github.com/dddpaul/cbr-currency-go` v1.0.7+ - CBR currency rate fetching
- `github.com/stretchr/testify` v1.11.1+ - Testing framework

## Error Handling

The tool continues processing even when encountering errors, logging them to stderr:
- Invalid date formats are skipped
- Malformed purchase descriptions are logged with row numbers
- Currency conversion failures use fallback rates or explicit rates when available

## Notes

- All text is converted to lowercase for consistency
- Person names default to "Общие" (Common) when not specified
- Empty rows in input CSV are automatically skipped
- The tool expects UTF-8 encoded input files