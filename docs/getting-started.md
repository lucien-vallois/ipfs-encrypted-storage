# Getting Started

This guide will help you get started with IPFS Encrypted Storage.

## Prerequisites

Before installing, ensure you have the following:

- **Go 1.21 or later** - Required for building the application
- **IPFS daemon** - Required for decentralized storage operations
- **Git** - For cloning the repository

### Installing IPFS

```bash
# Download and install IPFS
curl -L https://dist.ipfs.io/go-ipfs/v0.16.0/go-ipfs_v0.16.0_linux-amd64.tar.gz | tar xz
cd go-ipfs
sudo ./install.sh

# Initialize IPFS
ipfs init

# Start IPFS daemon
ipfs daemon
```

For other platforms, see the [IPFS installation guide](https://docs.ipfs.io/install/).

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/lucien-vallois/ipfs-encrypted-storage.git
cd ipfs-encrypted-storage

# Build the application
go build -o ipfs-storage ./src

# Or use the Makefile
make build
```

## Quick Start

### 1. Initialize the System

Generate your cryptographic key pair:

```bash
./ipfs-storage init
```

This creates `~/.ipfs-encrypted-storage/keys.json` with your Ed25519 key pair.

### 2. Upload a File

Encrypt and upload a file:

```bash
./ipfs-storage upload myfile.txt
```

You will be prompted for an encryption password.

### 3. Download a File

Download and decrypt a file:

```bash
./ipfs-storage download <CID> --metadata myfile.meta.json
```

## Basic Concepts

### Content Identifiers (CIDs)

IPFS uses content identifiers to address files. When you upload a file, you receive a CID that uniquely identifies that content:

```
QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE7
```

### Metadata Files

When you upload encrypted files, a `.meta.json` file is created containing:
- The IPFS CID
- Encryption parameters
- Digital signatures
- File metadata

Keep this file safe - it's required for decryption.

### Key Management

- **Private keys** are stored locally in `~/.ipfs-encrypted-storage/keys.json`
- **Passwords** are used for file encryption (not stored)
- **Public keys** are embedded in metadata for signature verification

## Next Steps

- Explore the [basic usage examples](examples/basic-usage.md)
- Review the [HTTP API status](api-rest.md)
- Read the [programmatic API and current limitations](programmatic-api.md)
- Check the [developer guide](developer-guide.md) for development workflows
