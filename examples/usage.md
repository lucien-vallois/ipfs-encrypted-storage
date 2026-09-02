# IPFS Encrypted Storage Examples

This document provides examples of using the encrypted IPFS storage system for secure, decentralized file storage.

## Prerequisites

1. **IPFS Node**: Install and run an IPFS node
   ```bash
   # Install IPFS
   wget https://dist.ipfs.io/go-ipfs/v0.16.0/go-ipfs_v0.16.0_linux-amd64.tar.gz
   tar -xzf go-ipfs_v0.16.0_linux-amd64.tar.gz
   cd go-ipfs
   ./ipfs init
   ./ipfs daemon
   ```

2. **Go**: Install Go 1.21+
   ```bash
   wget https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.19.5.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

3. **Build the CLI**:
   ```bash
   cd ipfs-encrypted-storage
   go build -o ipfs-storage ./src
   ```

## Basic Usage

### Initialize the System

First, initialize the encrypted storage system to generate your key pair:

```bash
./ipfs-storage init
```

This creates `~/.ipfs-encrypted-storage/keys.json` with your Ed25519 key pair.

### Upload a File

Upload and encrypt a file:

```bash
./ipfs-storage upload myfile.txt
```

This will:
1. Prompt for an encryption password
2. Encrypt the file with AES-256-GCM
3. Upload to IPFS
4. Save metadata to `myfile.meta.json`

### Download a File

Download and decrypt a file using its CID:

```bash
./ipfs-storage download QmYourCIDHere --metadata myfile.meta.json
```

### List Files

List all pinned files:

```bash
./ipfs-storage list
```

List contents of a directory:

```bash
./ipfs-storage list QmDirectoryCID
```

## Advanced Examples

### Batch Upload Directory

```bash
#!/bin/bash
# Upload all files in a directory
for file in /path/to/files/*; do
    if [ -f "$file" ]; then
        ./ipfs-storage upload "$file" --description "Batch upload $(basename "$file")"
    fi
done
```

### Automated Backup Script

```bash
#!/bin/bash
# Automated encrypted backup to IPFS

BACKUP_DIR="/home/user/important-data"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="backup_$TIMESTAMP.tar.gz"

# Create backup archive
tar -czf "$BACKUP_FILE" "$BACKUP_DIR"

# Upload with encryption
./ipfs-storage upload "$BACKUP_FILE" --description "Automated backup $TIMESTAMP"

# Clean up local file
rm "$BACKUP_FILE"

echo "Backup completed and uploaded to IPFS"
```

### Initialize the Local P2P Stub

```bash
#!/bin/bash
# This validates configuration and starts the in-memory stub only.
# Cross-node transport and file sharing are not implemented.
./ipfs-storage p2p --listen "/ip4/0.0.0.0/tcp/4001" --bootstrap=false
```

## Integration Examples

### Python Integration

```python
import subprocess
import json

def upload_file(filepath, password):
    """Upload file using CLI tool"""
    cmd = [
        './ipfs-storage', 'upload', filepath,
        '--password', password
    ]

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Upload failed: {result.stderr}")

    # Parse metadata file
    meta_file = filepath.replace('.txt', '.meta.json')
    with open(meta_file, 'r') as f:
        metadata = json.load(f)

    return metadata['cid']

def download_file(cid, metadata_file, password, output_path):
    """Download file using CLI tool"""
    cmd = [
        './ipfs-storage', 'download', cid,
        '--metadata', metadata_file,
        '--password', password,
        '--output', output_path
    ]

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Download failed: {result.stderr}")

# Usage
cid = upload_file('important.txt', 'my-secret-password')
download_file(cid, 'important.meta.json', 'my-secret-password', 'restored.txt')
```

### Containers

This repository does not currently ship or validate Docker or Kubernetes deployment manifests.

## Security Best Practices

### Password Management

1. **Use Strong Passwords**: Minimum 12 characters with mixed case, numbers, and symbols
2. **Password Manager**: Store passwords securely, never in plain text
3. **Regular Rotation**: Change passwords periodically

### Key Management

1. **Backup Keys**: Keep secure backups of your key pair
2. **Key Rotation**: Rotate keys periodically for high-security use cases
3. **Access Control**: Limit access to key files (chmod 600)

### Network Security

1. **Firewall**: Restrict IPFS ports to trusted networks
2. **VPN**: Use VPN for untrusted networks
3. **Private Swarm**: If needed, configure this directly on the external IPFS daemon

## Performance Optimization

### IPFS Configuration

```bash
# Optimize IPFS for your use case
ipfs config --json Swarm.ConnMgr.LowWater 50
ipfs config --json Swarm.ConnMgr.HighWater 200
ipfs config --json Datastore.StorageMax "100GB"
```

### Storage Optimization

1. **Compression**: Compress files before encryption
2. **Deduplication**: Use IPFS's built-in deduplication
3. **Garbage Collection**: Regular cleanup of unpinned content

### Network Optimization

External IPFS infrastructure only; the local P2P stub manages no peers or bandwidth.

1. **IPFS Peer Connections**: Maintain connections to reliable peers
2. **Bandwidth Limits**: Configure appropriate bandwidth limits
3. **Caching**: Use IPFS cluster for improved performance

## Troubleshooting

### Common Issues

**"Failed to connect to IPFS daemon"**
- Ensure IPFS daemon is running: `ipfs daemon`
- Check API endpoint: `curl localhost:5001/api/v0/id`

**"Decryption failed"**
- Verify password is correct
- Check metadata file integrity
- Ensure key pair is valid

**External IPFS daemon reports "No peers connected"**
- Check network connectivity
- Verify bootstrap peers are reachable
- Try a manual IPFS peer connection; the repository's local P2P stub does not dial peers

### Debug Mode

Enable verbose logging:

```bash
./ipfs-storage --verbose upload myfile.txt
```

Check IPFS logs:

```bash
ipfs log tail
```

## Monitoring

### Health Checks

```bash
# Check IPFS connectivity
curl "localhost:5001/api/v0/id"

# List content known to the configured IPFS client
./ipfs-storage list

# Verify encryption
./ipfs-storage download <cid> --metadata <meta> --output test.txt
```

### Metrics

The REST API exposes a basic JSON metrics endpoint. Most counters are placeholders; Prometheus export is not implemented.
