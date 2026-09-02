# IPFS Encrypted Storage

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0+-blue.svg?style=flat-square)](CHANGELOG.md)
[![Build Status](https://github.com/lucien-vallois/ipfs-encrypted-storage/actions/workflows/ci.yml/badge.svg)](https://github.com/lucien-vallois/ipfs-encrypted-storage/actions/workflows/ci.yml)

**Sistema de armazenamento de arquivos sobre IPFS com criptografia ponta-a-ponta, assinaturas criptográficas e um stub P2P local para desenvolvimento.**

## 🚀 Novidades na v1.0.0+

### ✅ Melhorias Implementadas
- **Arquitetura Refatorada**: CLI modular com comandos separados
- **Sistema de Erros Aprimorado**: EnhancedError com sugestões contextuais
- **Validação Robusta**: Entrada validada com feedback detalhado
- **Utilitários Centralizados**: JSON seguro, validação e retry
- **API REST Básica**: Acesso programático (parcialmente implementada)

### 🔧 Status Atual
- **CLI**: Comandos modulares implementados; o comando P2P inicializa apenas um stub local
- **API REST**: Parcial; vários endpoints de escrita ainda retornam HTTP 501
- **Testes**: Suíte local executável com cobertura dos pacotes de produção
- **Documentação**: Alinhada ao escopo atual e às limitações conhecidas

## ✨ Funcionalidades

### Criptografia e Segurança
- **AES-256-GCM**: Criptografia autenticada
- **Ed25519 Signatures**: Assinaturas para integridade e autenticidade
- **Argon2 KDF**: Derivação de chaves segura e configurável
- **Demonstração Schnorr**: API experimental; não usar como autorização de produção
- **Gestão de Chaves**: Sobrescrita best-effort de um buffer; cópias criadas pelo runtime podem permanecer

### Armazenamento Decentralizado
- **Integração IPFS**: Armazenamento com endereçamento de conteúdo
- **P2P experimental**: Stub local com IDs libp2p e validação de multiaddresses
- **DHT/PubSub em rede**: Planejados; não implementados no stub atual
- **DID Integration**: Identidade decentralizada (W3C DID)

### Interface e Usabilidade
- **CLI Modular**: Interface de linha de comando estruturada
- **API REST**: Acesso programático com autenticação
- **Validação Inteligente**: Entrada validada com sugestões
- **Sistema de Erros**: Mensagens contextuais e acionáveis

## 🏗️ Arquitetura Atual

O sistema implementa padrões avançados de armazenamento decentralizado:

### Estrutura Modular (v1.0.0+)
```
src/
├── cmd/                 # Comandos CLI modulares
│   ├── upload.go       # Upload de arquivos
│   ├── download.go     # Download de arquivos
│   ├── api.go          # Servidor API REST
│   └── ...             # Outros comandos
├── utils/               # Utilitários centralizados
│   ├── validation.go   # Validação robusta
│   ├── jsonutils.go    # Conversões JSON seguras
│   └── retry.go        # Lógica de retry
├── errors/              # Sistema de erros aprimorado
├── api/                 # API REST (parcial)
└── ...                  # Outros módulos
```

### Padrões Implementados
- **Command Pattern**: CLI modular e extensível
- **Enhanced Error Handling**: Contexto e sugestões acionáveis
- **Comprehensive Validation**: Entrada validada em múltiplas camadas
- **Factory Pattern**: Criação consistente de componentes

## 🔒 Modelo de Segurança

### Criptografia de Ponta-a-Ponta
- **AES-256-GCM**: Criptografia autenticada com nonces únicos
- **Argon2 KDF**: Parâmetros configuráveis para diferentes perfis
- **Ed25519**: Assinaturas para verificação de integridade
- **Demonstração Schnorr**: Fora do modelo de autorização de produção
- **Gestão de Chaves**: Sobrescrita best-effort do buffer mantido por `SecureKey`; cópias do runtime podem permanecer

### Validação e Segurança
- **Validação Multifacetada**: Arquivos, senhas, CIDs, endereços peer
- **Tratamento de Erros**: Erros estruturados com contexto e sugestões
- **Autenticação API**: Baseada em API keys (configurável)
- **Rate Limiting**: Não implementado

## ⚡ Performance

### Otimizações Implementadas
- **Benchmarks Abrangentes**: Suite de benchmarking incluída
- **Operações Otimizadas**: Criptografia/descriptografia eficientes
- **Concorrência**: Suporte a operações paralelas
- **Overhead Mínimo**: ~1KB por arquivo de metadados

### Medição de Performance

Execute `go test -bench=. ./tests` no ambiente de destino. O projeto não declara SLOs de latência, throughput, memória ou CPU; os resultados dependem do hardware e do endpoint IPFS.

## 🚀 Quick Start

### Pré-requisitos
- **Go 1.21+** (versão atual recomendada)
- **IPFS Daemon** rodando localmente ou endpoint remoto
- **Git** para controle de versão

### Instalação
```bash
# Clonar repositório
git clone https://github.com/lucien-vallois/ipfs-encrypted-storage.git
cd ipfs-encrypted-storage

# Instalar dependências
go mod download

# Verificar instalação
go version  # Deve mostrar 1.21+
```

### Build e Execução
```bash
# Build otimizado
go build -o ipfs-encrypted-storage ./src

# Ou build com flags de produção
CGO_ENABLED=0 go build -ldflags="-w -s" -o ipfs-encrypted-storage ./src

# Ver comandos disponíveis
./ipfs-encrypted-storage --help
```

### Primeiros Passos
```bash
# 1. Gerar o arquivo local de chaves
./ipfs-encrypted-storage init

# 2. Criar arquivo de teste
echo "Conteúdo confidencial para teste" > secret.txt

# 3. Upload com criptografia (senha será solicitada)
./ipfs-encrypted-storage upload secret.txt

# 4. Listar arquivos armazenados
./ipfs-encrypted-storage list

# 5. Verificar integridade
./ipfs-encrypted-storage verify <CID> --metadata secret.meta.json
```

## 🌐 API REST (Beta)

Acesse o sistema programaticamente via HTTP REST API:

```bash
# Configurar uma chave antes de iniciar o servidor
export IPFS_API_KEY="substitua-por-uma-chave-forte"

# Iniciar servidor API
./ipfs-encrypted-storage api --port 8080

# Verificar saúde
curl http://localhost:8080/api/v1/health

# Listar arquivos (com autenticação)
curl -H "X-API-Key: $IPFS_API_KEY" \
  http://localhost:8080/api/v1/files

# Status P2P
curl -H "X-API-Key: $IPFS_API_KEY" \
  http://localhost:8080/api/v1/p2p/status

# Métricas do sistema
curl -H "X-API-Key: $IPFS_API_KEY" \
  http://localhost:8080/api/v1/metrics
```

**Nota:** API parcialmente implementada. Upload/download funcionais pendentes.

## ⚠️ Tratamento de Erros

Os comandos retornam erros contextuais e sugestões quando a implementação conhece uma ação de recuperação.

### Initialize
```bash
./ipfs-storage init
```

### Upload File
```bash
./ipfs-storage upload myfile.txt
# Enter encryption password when prompted
```

### Download File
```bash
./ipfs-storage download <CID> --metadata myfile.meta.json
```

## Architecture

### Components

#### Encryption Module (`encryption.go`)
- AES-256-GCM encryption/decryption
- Argon2 key derivation
- Ed25519 digital signatures
- Best-effort overwrite helper for one in-memory key buffer
- Stream encryption support

#### IPFS Client (`ipfs_client.go`)
- IPFS API integration
- File upload/download
- Directory operations
- Pinning management
- Health checks and monitoring

#### P2P Local Stub (`p2p.go`)
- In-memory stub for CLI and tests
- Valid libp2p peer IDs and multiaddress validation
- Local storage and local-only PubSub
- Protocol handler registration
- Bootstrap address validation (no network connection)

#### Zero-Knowledge Proofs (`zkp.go`)
- Experimental Schnorr transcript for learning and tests
- Access-context binding demonstration with intentionally small parameters
- No production-safe range proof or authorization system

#### Decentralized Identity (`did.go`)
- DID document management
- Verifiable credentials
- DID resolution and registry
- W3C DID specification compliance

#### Content-Addressable Storage (`content_addressable.go`)
- Content addressing with SHA-256
- Merkle DAG construction
- File system abstraction
- Block-level operations

#### CLI Interface (`main.go`)
- Command-line operations
- Configuration management
- Interactive password input
- Local P2P stub management

#### Comprehensive Test Suite (`tests/`)
- Unit tests for all modules
- Integration tests
- Performance benchmarks
- Mock implementations for testing

### Security Flow

1. **Key Derivation**: Password → Argon2 → 256-bit key
2. **Encryption**: AES-256-GCM with random nonce
3. **Signing**: Ed25519 signature of ciphertext
4. **Storage**: IPFS content addressing
5. **Verification**: Signature validation on retrieval

## Usage Examples

### Basic Operations

```bash
# Initialize system (generates key pair)
ipfs-storage init

# Upload with encryption
ipfs-storage upload document.pdf

# Download and decrypt
ipfs-storage download QmCID123 --metadata document.meta.json

# List stored files
ipfs-storage list

# Initialize the local P2P stub
ipfs-storage p2p --bootstrap
```

### Advanced Usage

```bash
# Upload without encryption
ipfs-storage upload public.txt --encrypt=false

# Custom metadata output
ipfs-storage upload secret.dat --output custom.meta.json

# Configure the local P2P stub address
ipfs-storage p2p --listen "/ip4/0.0.0.0/tcp/4002" --bootstrap=false
```

### Programmatic Usage

```go
package main

import (
    "log"
    "os"

    "ipfs-encrypted-storage/src/encryption"
    "ipfs-encrypted-storage/src/ipfs"
)

func main() {
    password := os.Getenv("ENCRYPTION_PASSWORD")
    if password == "" {
        log.Fatal("ENCRYPTION_PASSWORD is required")
    }

    client, err := ipfs.NewIPFSClient("localhost:5001")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    plaintext := []byte("sensitive data")
    publicKey, privateKey, err := encryption.GenerateKeyPair()
    if err != nil {
        log.Fatal(err)
    }
    ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
    if err != nil {
        log.Fatal(err)
    }

    cid, err := client.AddFile(ciphertext, "encrypted.dat")
    if err != nil {
        log.Fatal(err)
    }

    downloaded, err := client.GetFile(cid)
    if err != nil {
        log.Fatal(err)
    }
    decrypted, err := encryption.DecryptWithMetadata(downloaded, metadata, password, publicKey)
    if err != nil || string(decrypted) != string(plaintext) {
        log.Fatal("encrypted round trip failed")
    }
}
```

## API Reference

### Encryption

```go
// Generate key pair
publicKey, privateKey, err := encryption.GenerateKeyPair()

// Encrypt with metadata
ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)

// Decrypt with metadata
plaintext, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
```

### IPFS Client

```go
client, err := ipfs.NewIPFSClient("localhost:5001")
if err != nil {
    return err
}
defer client.Close()

// Upload file
cid, err := client.AddFile(data, filename)

// Download file
data, err := client.GetFile(cid)

// Pin/unpin
err = client.PinFile(cid)
err = client.UnpinFile(cid)
```

### Local P2P Stub

O módulo atual é um stub em memória. Ele não abre transporte libp2p nem conecta nós distintos.

```go
node, err := p2p.NewP2PNode("/ip4/0.0.0.0/tcp/0")
if err != nil {
    return err
}
defer node.Close()

subscription, err := node.SubscribeToTopic("my-topic", func(peerID string, message []byte) error {
    fmt.Printf("local message from %s: %s\n", peerID, message)
    return nil
})
if err != nil {
    return err
}
defer subscription.Cancel()

err = node.PublishToTopic("my-topic", []byte("message"))
```

## Configuration

### Environment Variables

- `IPFS_API_KEY`: Required credential for protected REST endpoints

Use the global `--ipfs-url` and `--config` flags for the IPFS endpoint and config file path.

### Config File

```json
{
  "ipfs": {
    "url": "localhost:5001",
    "timeout": "30s"
  },
  "encryption": {
    "key_derivation": {
      "time": 1,
      "memory": 65536,
      "threads": 4
    }
  },
  "p2p": {
    "listen_addr": "/ip4/0.0.0.0/tcp/0",
    "bootstrap_peers": []
  }
}
```

## Security Considerations

### Key Management
- Keys are stored locally in `~/.ipfs-encrypted-storage/keys.json`
- The private key is base64-encoded in JSON, not encrypted; protect the file with operating-system permissions
- Automated key rotation is not implemented; replacing keys is currently a manual operation

### Password Security
- Use strong, unique passwords
- Passwords are never stored
- Argon2 provides resistance to brute force

### Network Security
- File payloads are encrypted before they are sent to IPFS
- The REST API requires the configured `IPFS_API_KEY`
- P2P transport, connection authentication, and private-swarm management are not implemented by this repository

## Performance Tuning

### IPFS Configuration
```bash
# Increase connection limits
ipfs config --json Swarm.ConnMgr.HighWater 200

# Storage optimization
ipfs config --json Datastore.StorageMax "500GB"
```

### Application Tuning
```go
// Adjust encryption parameters
config := &encryption.KeyDerivationConfig{
    Time:    2,        // More iterations = more security
    Memory:  131072,   // More memory = more security
    Threads: 8,        // More threads = faster derivation
}
```

## 📚 Documentação e Desenvolvimento

### Documentação Técnica
- **[Arquitetura](docs/architecture.md)**: Visão geral da arquitetura e componentes
- **[API REST](docs/api-rest.md)**: Estado e contratos da API programática
- **[Utilitários](docs/utils.md)**: Documentação dos utilitários centralizados
- **[Sistema de Erros](docs/error-system.md)**: Guia do EnhancedError
- **[Exemplos](docs/examples/basic-usage.md)**: Exemplos práticos de uso
- **[Guia do Desenvolvedor](docs/developer-guide.md)**: Guia técnico do projeto

### Desenvolvimento
```bash
# Configurar ambiente de desenvolvimento
go mod tidy
go install github.com/cosmtrek/air@latest  # Hot reload

# Executar testes
go test ./... -v

# Build de desenvolvimento
go build -o dev-build ./src

# Verificar cobertura de testes
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Contribuição
1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'feat: adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

### Padrões de Commit
- `feat:` novas funcionalidades
- `fix:` correções de bugs
- `docs:` mudanças na documentação
- `refactor:` refatoração de código
- `test:` adição ou correção de testes

## Troubleshooting

### Common Issues

**"connection refused"**
- Ensure IPFS daemon is running: `ipfs daemon`
- Check API port: `curl localhost:5001/api/v0/id`

**"decryption failed"**
- Verify password is correct
- Check metadata file hasn't been corrupted
- Ensure same key pair is used

**IPFS daemon reports "no peers found"**
- Check network connectivity
- Try manual bootstrap: `ipfs bootstrap add <peer>`
- Verify firewall settings

### Debug Logging
```bash
export GLOG_logtostderr=1
./ipfs-storage --verbose upload file.txt
```

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific test suite
go test ./tests/encryption_test.go -v

# Run benchmarks
go test -bench=. -benchmem ./tests

# Run performance validation
go test -bench=BenchmarkEncryptionThroughput ./tests -benchtime=10s

# Run integration tests
go test -tags=integration ./tests
```

### Test Coverage

The project includes comprehensive testing:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow validation
- **Performance Benchmarks**: Throughput and latency measurements
- **Concurrent Operations**: Multi-goroutine testing
- **Mock Implementations**: Testing without external dependencies

### Benchmark Results

Benchmark results depend on the host and are not release gates. Run `go test -bench=. -benchmem ./tests` and compare results only on the same controlled environment.

## 📈 Roadmap 2026

### Q1 2026: API REST Completa
- ✅ CLI modular implementado
- 🔄 Migração para Gin framework
- 🔄 Upload/download funcionais via API
- 🔄 Autenticação JWT
- 🔄 Documentação OpenAPI/Swagger

### Q2 2026: Escalabilidade
- 🔄 Suporte a múltiplos nós IPFS
- 🔄 Load balancing inteligente
- 🔄 Cache distribuído
- 🔄 Otimizações de performance

### Q3 2026: Recursos Avançados
- 🔄 Interface web responsiva
- 🔄 SDKs para Python, JavaScript, Java
- 🔄 Integração com blockchains
- 🔄 Suporte a NFT/storage

### Q4 2026: Produção
- 🔄 Auditoria de segurança profissional
- 🔄 Compliance (GDPR, CCPA, LGPD)
- 🔄 Multi-cloud deployment
- 🔄 SLA e monitoring avançado

## 📋 Changelog

### v1.0.0+ (Dezembro 2025)
**Refatoração arquitetural**
- ✅ Arquitetura CLI modular implementada
- ✅ Sistema EnhancedError com contexto e sugestões
- ✅ Validação robusta de entrada em todas as camadas
- ✅ Utilitários centralizados (JSON, validação, retry)
- ✅ API REST básica parcialmente implementada
- ✅ Documentação alinhada ao escopo atual
- ✅ Padrões de código e desenvolvimento estabelecidos

### v0.9.x (Antes)
- Sistema monolítico funcional
- Funcionalidades básicas de criptografia
- Integração básica com IPFS
- CLI simples

## 🤝 Contribuição

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças seguindo convenções
4. Adicione testes abrangentes
5. Submeta um Pull Request detalhado

### Padrões de Commit
- `feat:` novas funcionalidades
- `fix:` correções de bugs
- `docs:` mudanças na documentação
- `refactor:` refatoração de código
- `test:` adição ou correção de testes

## 📄 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo LICENSE para detalhes.

## Development

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./tests
```

### Building
```bash
# Using make
make all

# Manual build
go build -o ipfs-storage ./src
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Lint (requires golangci-lint)
make lint
```
