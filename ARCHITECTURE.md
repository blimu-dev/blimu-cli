# Blimu CLI Architecture

## Overview

The Blimu CLI follows a **thin client** architecture where the CLI handles basic validation and user experience, while the Blimu API is responsible for authoritative validation and OpenAPI spec generation.

## Architecture Principles

### 🎯 **Single Source of Truth**

- The Blimu API is the authoritative source for validation rules and spec generation
- CLI provides fast local validation for development workflow
- Server-side validation ensures consistency and security

### 📦 **Separation of Concerns**

- **CLI**: User experience, local validation, configuration management
- **API**: Authoritative validation, OpenAPI spec generation, versioning control
- **SDK-Gen**: OpenAPI spec to SDK code generation

### 🔄 **Workflow**

```
.blimu/*.yml → CLI (merge + local validate) → API (validate + generate spec) → SDK-Gen → Custom SDK
```

## Component Responsibilities

### CLI (`blimucli`)

- **Configuration Management**: Load, merge, and save YAML configurations
- **Local Validation**: Fast client-side validation for development feedback
- **API Integration**: Send configurations to Blimu API for processing
- **User Experience**: Rich CLI interface with helpful error messages

### Blimu API (`/v1/config/*`)

- **Authoritative Validation**: Server-side validation with business rules
- **OpenAPI Generation**: Convert user configs to custom OpenAPI specs
- **Versioning Control**: Ensure generated specs match current API version
- **Security**: Validate and sanitize user configurations

### SDK Generator (`sdk-gen`)

- **Code Generation**: Convert OpenAPI specs to type-safe SDKs
- **Multiple Languages**: Support TypeScript, Go, etc.
- **Template Management**: Maintain code generation templates

## Configuration Flow

### 1. Local Development

```bash
# Fast local validation for development
blimucli validate

# Local validation + server validation
blimucli validate --remote
```

### 2. SDK Generation

```bash
# Generate SDK via API
blimucli generate --type typescript --output ./my-sdk
```

**Flow:**

1. CLI merges all YAML files into single JSON payload
2. CLI sends JSON to `/v1/config/generate-sdk`
3. API validates config and generates custom OpenAPI spec
4. API returns generated spec to CLI
5. CLI uses `sdk-gen` to generate SDK from spec

## API Endpoints

### `/v1/config/validate`

**Request:**

```json
{
  "resources": {
    /* resources.yml content */
  },
  "entitlements": {
    /* entitlements.yml content */
  },
  "features": {
    /* features.yml content */
  },
  "plans": {
    /* plans.yml content */
  },
  "version": "1.0"
}
```

**Response:**

```json
{
  "valid": true,
  "errors": [],
  "spec": {
    /* Generated OpenAPI spec */
  }
}
```

### `/v1/config/generate-sdk`

**Request:**

```json
{
  "resources": {
    /* ... */
  },
  "entitlements": {
    /* ... */
  },
  "features": {
    /* ... */
  },
  "plans": {
    /* ... */
  },
  "version": "1.0",
  "sdk_options": {
    "type": "typescript",
    "package_name": "my-client",
    "client_name": "MyClient"
  }
}
```

**Response:**

```json
{
  "success": true,
  "spec": {
    /* Custom OpenAPI spec */
  },
  "errors": []
}
```

## Benefits

### ✅ **Consistency**

- Same validation logic for CLI and web interface
- Server-side validation prevents malicious configurations
- Versioned API ensures compatibility

### ✅ **Performance**

- Fast local validation for development feedback
- Optional remote validation for authoritative checks
- Cached API responses for common configurations

### ✅ **Maintainability**

- Clear separation of concerns
- API evolution doesn't require CLI updates for core logic
- Centralized business rules in API

### ✅ **Security**

- Server-side validation and sanitization
- API authentication for all operations
- Controlled access to spec generation

## File Structure

```
blimucli/
├── pkg/
│   ├── config/          # Configuration loading and merging
│   │   ├── config.go    # YAML file loading
│   │   └── merger.go    # JSON payload creation
│   ├── api/             # API client
│   │   └── client.go    # Blimu API integration
│   ├── auth/            # Authentication
│   │   └── auth.go      # API key management
│   └── blimu/           # Local validation
│       └── validator.go # Client-side validation rules
├── internal/
│   └── cli/             # CLI commands
│       ├── validate.go  # Validation command
│       ├── generate.go  # SDK generation command
│       └── init.go      # Initialization command
└── cmd/
    └── blimucli/        # CLI entry point
        └── main.go
```

## Future Enhancements

- **Caching**: Cache API responses for faster repeated operations
- **Offline Mode**: Enhanced local validation for offline development
- **Config Diff**: Show differences between local and deployed configs
- **Multi-Environment**: Support for different API environments (dev, staging, prod)
