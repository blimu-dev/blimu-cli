# Blimu CLI

A command-line tool for managing Blimu configurations and generating custom SDKs.

## Features

- 🚀 **Initialize** new `.blimu` configurations
- ✅ **Validate** your resource configurations
- 🔧 **Generate** custom SDKs based on your resources
- 🔐 **Authenticate** with Blimu API

## Installation

```bash
go install github.com/blimu-dev/blimu-cli/cmd/blimucli@latest
```

Or build from source:

```bash
git clone https://github.com/blimu-dev/blimu-cli
cd blimu-cli
go build -o blimucli cmd/blimucli/main.go
```

## Quick Start

### 1. Initialize a new configuration

```bash
blimu init
```

This creates a `.blimu/resources.yml` file with a basic configuration.

### 2. Edit your configuration

Edit `.blimu/resources.yml` to define your resources:

```yaml
organization:
  roles: [admin, editor, viewer]

workspace:
  roles: [admin, editor, viewer]
  roles_inheritance:
    editor: [organization->admin]
    viewer: [organization->editor]
  parents:
    organization:
      required: true

brand:
  roles: [admin, editor, viewer]
  parents:
    workspace:
      required: false
```

### 3. Validate your configuration

```bash
blimucli validate
```

### 4. Set up authentication

Authenticate using OAuth:

```bash
blimu auth login
```

Test your authentication:

```bash
blimu auth test
```

### 5. Generate your custom SDK

```bash
blimucli generate --output ./my-blimu-sdk --package-name my-blimu-client
```

## Commands

### `blimucli init`

Initialize a new `.blimu` configuration directory.

**Options:**

- `--force, -f`: Force initialization even if `.blimu` directory exists

### `blimucli validate`

Validate your `.blimu/resources.yml` configuration.

Checks for:

- Valid resource definitions
- Correct role inheritance syntax
- Valid parent relationships
- No circular dependencies

### `blimucli generate`

Generate a custom SDK based on your resource configuration.

**Options:**

- `--output, -o`: Output directory (default: `./blimu-sdk`)
- `--package-name, -p`: Package name (default: `blimu-client`)
- `--client-name, -c`: Client class name (default: `BlimuClient`)
- `--type, -t`: SDK type, currently only `typescript` (default: `typescript`)
- `--force, -f`: Force generation even if output directory exists

### `blimu auth login`

Authenticate with Blimu using OAuth (Clerk).

**Options:**

- `--environment`: Environment to authenticate with (default: `env_blimu_platform`)
- `--api-url`: Clerk domain for OAuth (default: `https://clerk.blimu.dev`)

### `blimu auth test`

Test your OAuth authentication with the Blimu API.

## Generated SDK Usage

After generating your SDK, you can use it like this:

```typescript
import { BlimuClient } from "./blimu-client";

const client = new BlimuClient({
  baseURL: "https://api.blimu.dev",
  bearerToken: "your-oauth-token",
});

// Use your custom resources
const org = await client.Organization.create({ name: "My Org" });
const workspaces = await client.Workspace.list();
const brand = await client.Brand.get("brand-id");
```

## Configuration Format

The `.blimu/resources.yml` file defines your resources:

```yaml
resource_name:
  roles: [list, of, roles]
  roles_inheritance:
    role_name: [parent_resource->parent_role]
  parents:
    parent_resource:
      required: true|false
```

## Development

This project is structured similar to [sdk-gen](https://github.com/blimu-dev/sdk-gen):

```
blimu-cli/
├── cmd/blimu-cli/          # CLI entry point
├── pkg/                   # Public packages
│   ├── config/           # Configuration handling
│   ├── auth/             # Authentication
│   ├── blimu/            # Blimu-specific types
│   └── generator/        # SDK generation
├── internal/             # Private packages
│   ├── cli/              # CLI commands
│   └── openapi/          # OpenAPI spec generation
└── .blimu/               # Example configuration
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
