# Multi-Environment Example

This example demonstrates how to use klastr's multi-environment support.

## Structure

```
.
├── klastr.yaml              # Base configuration
└── environments/
    ├── dev/
    │   └── overlay.yaml     # Dev-specific patches
    ├── staging/
    │   └── overlay.yaml     # Staging-specific patches
    └── production/
        └── overlay.yaml     # Production-specific patches
```

## Usage

### List environments
```bash
klastr env list
```

### Create a new environment
```bash
klastr env create testing
```

### Show environment configuration
```bash
klastr env show production
```

### Deploy with environment
```bash
# Deploy dev environment
klastr run --environment dev

# Deploy staging
klastr run --environment staging

# Deploy production
klastr run --environment production
```

### Other commands with environment
```bash
klastr lint --environment dev
klastr upgrade --environment staging
klastr status --environment production
```

## How it works

1. **Base configuration** (`klastr.yaml`) defines the common settings
2. **Overlay** (`environments/*/overlay.yaml`) patches the base:
   - `patches`: Strategic merge patches (e.g., change worker count)
   - `values`: Variables for templating
3. The final configuration is computed by applying patches on top of base

## Environment Differences

| Environment | Workers | Monitoring | Cert-Manager | Domain |
|-------------|---------|------------|--------------|--------|
| dev         | 0       | No         | No           | dev.localhost |
| staging     | 2       | Yes        | No           | staging.example.com |
| production  | 5       | Yes        | Yes          | example.com |
