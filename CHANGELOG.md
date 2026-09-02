# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Implemented `Close()` method in IPFSClient for proper resource management
- Implemented `PublishName()` for IPNS name publishing
- Complete `CreateDirectory()` implementation for multiple files
- Local P2P message parsers for storage, file-request, and topic payloads
- Configuration system with `config.json` support
- Retry logic with exponential backoff for network operations
- `verify` command for integrity validation
- Improved logging with JSON output and configurable log levels
- GitHub Actions CI/CD pipeline
- CHANGELOG.md following Keep a Changelog format

### Changed
- Improved error handling in all IPFS client operations
- Enhanced integrity validation with real implementations (removed placeholders)
- Better configuration management with validation and defaults
- Improved local P2P payload parsers

### Fixed
- Fixed error handling in `NewIPFSClient()` calls throughout the codebase
- Fixed incomplete implementations in IPFS client methods
- Fixed placeholder implementations in integrity validation
- Added structured parsing to the local P2P stubs

## [1.0.0] - Initial Release

### Added
- AES-256-GCM encryption with Argon2 key derivation
- Ed25519 digital signatures for integrity verification
- IPFS integration for decentralized storage
- In-memory P2P stub using libp2p peer ID types
- CLI interface with multiple commands
- Experimental Schnorr proof demo (not production authorization)
- Decentralized Identity (DID) support
- Content-addressable storage patterns
- Comprehensive test suite

