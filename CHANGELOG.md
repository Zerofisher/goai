# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-10-20

### Added

- **Official SDK Integration**: Migrated to official OpenAI and Anthropic SDKs
  - OpenAI integration using `github.com/openai/openai-go/v2`
  - Anthropic integration using `github.com/anthropics/anthropic-sdk-go`
- **Multi-Provider Support**: Factory pattern for extensible LLM provider support
- **Mock Client**: Testing and development support with `pkg/llm/mock`
- **Performance Tests**: Comprehensive benchmark tests for LLM operations
- **Enhanced Documentation**:
  - Created `docs/LLM_PROVIDERS.md` with detailed provider integration guide
  - Updated `README.md` with multi-provider configuration examples
- **Tool Calling Support**: Full support for function calling across both providers
- **Streaming Support**: Enhanced streaming responses for both OpenAI and Anthropic

### Changed

- **Breaking**: Replaced custom HTTP implementation with official SDKs
- **API**: Updated `llm.Client` interface for better provider abstraction
- **Configuration**: Enhanced config support for multiple providers
- **Default Model**: OpenAI default model changed to `gpt-4.1-mini`
- **Dependencies**:
  - Added `github.com/openai/openai-go/v2 v2.7.1`
  - Added `github.com/anthropics/anthropic-sdk-go v1.14.0`

### Removed

- Old custom HTTP implementation for OpenAI
- Legacy provider-specific code
- Backup files (`.old` files)

### Fixed

- Improved error handling across all LLM operations
- Better type safety with official SDK types
- More reliable streaming implementation

### Testing

- All unit tests passing (55.3% coverage)
- New performance benchmarks added
- Integration tests for both providers
- Zero linting issues

### Performance

- `BenchmarkCreateMessage`: 92.87 ns/op
- `BenchmarkStreamMessage`: 1328 ns/op
- `BenchmarkToolDefinitionConversion`: 131.3 ns/op

## [0.1.0] - 2025-10-19

### Added

- Initial release
- Basic agent system with conversation management
- Tool system (bash, file, edit, search, todo)
- OpenAI integration (custom HTTP client)
- Message history management
- Configuration system (YAML-based)
- Todo and reminder systems
- Interactive CLI interface

### Features

- Multi-round tool execution
- Secure file operations with path validation
- Command filtering for bash tool
- Code search with caching
- Task tracking and progress monitoring

---

## Migration Guide (0.1.0 → 0.2.0)

### Configuration Changes

If you're upgrading from 0.1.0, update your configuration:

**Old (0.1.0):**

```yaml
model:
  provider: "openai"
  name: "gpt-4"
```

**New (0.2.0):**

```yaml
model:
  provider: "openai" # or "anthropic"
  name: "gpt-4.1-mini" # for OpenAI
  # name: "claude-3-7-sonnet-latest"  # for Anthropic
```

### Environment Variables

**OpenAI:**

```bash
export OPENAI_API_KEY="your-api-key"
```

**Anthropic (NEW):**

```bash
export ANTHROPIC_API_KEY="your-api-key"
```

### Breaking Changes

1. The internal `llm.Client` interface has been updated
2. Custom HTTP implementation removed in favor of official SDKs
3. Some internal types have changed to match SDK conventions

### Benefits of Upgrading

- More reliable API communication
- Better error messages
- Official SDK bug fixes and improvements
- Support for Anthropic/Claude models
- Better streaming performance
- Enhanced tool calling support
