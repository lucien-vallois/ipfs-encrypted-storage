# Basic Usage Examples

This document provides basic usage examples for IPFS Encrypted Storage.

**Versão:** 1.0.0+ (Comandos Refatorados e API REST)

## Arquitetura Atual

O sistema agora possui:
- **CLI Modular:** Comandos separados em arquivos individuais
- **API REST:** Acesso programático via HTTP
- **Sistema de Erros Aprimorado:** Com sugestões contextuais
- **Validação Robusta:** Entrada validada com feedback detalhado

## File Operations

### Encrypting and Uploading Files

```bash
# Upload com criptografia (senha será solicitada)
./ipfs-encrypted-storage upload document.pdf

# Upload sem criptografia
./ipfs-encrypted-storage upload public.txt --encrypt=false

# Upload com arquivo de metadados personalizado
./ipfs-encrypted-storage upload secret.dat --output custom.meta.json

# Upload com descrição
./ipfs-encrypted-storage upload photo.jpg --description "Foto da família de férias"

# A validação de entrada já faz parte do upload
./ipfs-encrypted-storage upload large-file.zip
```

### Downloading and Decrypting Files

```bash
# Download arquivo criptografado (usa arquivo de metadados para descriptografia)
./ipfs-encrypted-storage download QmCID123 --metadata document.meta.json

# Download para local específico
./ipfs-encrypted-storage download QmCID123 --metadata document.meta.json --output restored.pdf

# Download arquivo não criptografado
./ipfs-encrypted-storage download QmCID456 --output public.txt

# A assinatura dos metadados é verificada durante a descriptografia
./ipfs-encrypted-storage download QmCID123 --metadata document.meta.json
```

### Managing Files

```bash
# Listar todos os arquivos fixados
./ipfs-encrypted-storage list

# Listar conteúdo de um diretório
./ipfs-encrypted-storage list QmDirectoryCID

# Remover fixação de arquivo (remove do armazenamento local)
./ipfs-encrypted-storage delete QmCID123

# Verificar integridade de arquivo armazenado
./ipfs-encrypted-storage verify QmCID123 --metadata document.meta.json
```

## API REST

### Iniciando o Servidor API

```bash
# Iniciar servidor API (porta 8080 por padrão)
export IPFS_API_KEY="substitua-por-uma-chave-forte"
./ipfs-encrypted-storage api

# Iniciar servidor API em porta específica
./ipfs-encrypted-storage api --port 9090

# Verificar saúde da API
curl http://localhost:8080/api/v1/health
```

### Upload via API REST

```bash
# Upload de arquivo via API (exemplo - funcionalidade pendente)
curl -X POST \
  -H "X-API-Key: $IPFS_API_KEY" \
  -F "file=@document.pdf" \
  -F "password=my-secure-password" \
  http://localhost:8080/api/v1/files

# Resposta esperada:
# {
#   "cid": "Qm...",
#   "filename": "document.pdf",
#   "size": 1048576,
#   "uploaded_at": 1704067200,
#   "encrypted": true
# }
```

### Download via API REST

```bash
# Download de arquivo via API (exemplo - funcionalidade pendente)
curl -H "X-API-Key: $IPFS_API_KEY" \
  "http://localhost:8080/api/v1/files/QmCID123?password=my-password&metadata=document.meta.json" \
  --output downloaded.pdf
```

### Listar Arquivos via API

```bash
# Listar arquivos armazenados
curl -H "X-API-Key: $IPFS_API_KEY" \
  http://localhost:8080/api/v1/files

# Resposta:
# {
#   "files": [],
#   "count": 0
# }
```

### Monitoramento via API

```bash
# Verificar saúde detalhada
curl http://localhost:8080/api/v1/health/deep

# Obter métricas do sistema
curl -H "X-API-Key: $IPFS_API_KEY" \
  http://localhost:8080/api/v1/metrics

# Consultar a resposta estática de peers do stub
curl -H "X-API-Key: $IPFS_API_KEY" \
  http://localhost:8080/api/v1/p2p/peers
```

## Stub P2P Local

### Inicializando o Stub

```bash
# Inicializar e validar os endereços de bootstrap
./ipfs-encrypted-storage p2p --bootstrap

# Configurar um endereço local
./ipfs-encrypted-storage p2p --listen "/ip4/0.0.0.0/tcp/4001"

# Inicializar sem validar bootstrap
./ipfs-encrypted-storage p2p --bootstrap=false

# O comando apenas inicializa o stub local; conexão de rede ainda não existe
```

## Batch Operations

### Processing Multiple Files

```bash
#!/bin/bash
# Encrypt and upload multiple files

FILES=("doc1.pdf" "doc2.pdf" "image.jpg")
PASSWORD="my-secure-password"

for file in "${FILES[@]}"; do
    echo "Processing $file..."
    ./ipfs-storage upload "$file" --password "$PASSWORD"
done
```

### Backup Script

```bash
#!/bin/bash
# Automated encrypted backup

BACKUP_DIR="/home/user/documents"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="backup_$TIMESTAMP.tar.gz"

# Create archive
tar -czf "$BACKUP_FILE" "$BACKUP_DIR"

# Upload encrypted
CID=$(./ipfs-storage upload "$BACKUP_FILE" --description "Backup $TIMESTAMP" | grep "CID:" | cut -d' ' -f2)

# Clean up
rm "$BACKUP_FILE"

echo "Backup completed. CID: $CID"
```

## Sistema de Erros Aprimorado

O sistema agora fornece erros contextuais com sugestões acionáveis.

### Exemplo de Tratamento de Erro Aprimorado

```bash
# Tentativa de upload com erro (arquivo muito grande)
./ipfs-encrypted-storage upload large-file.zip
# Saída:
# ❌ File too large: 150000000 bytes (max: 104857600 bytes) (operation: file_validation)
#
# 💡 Suggestions:
#    1. Maximum file size is 100MB
#    2. Consider splitting large files
#    3. Compress the file if possible
```

```bash
# Tentativa de download com erro de rede
./ipfs-encrypted-storage download QmCID123 --metadata doc.meta.json
# Saída:
# ❌ failed to connect to IPFS: connection refused (operation: IPFS download) (CID: QmCID123)
#
# 💡 Suggestions:
#    1. Ensure IPFS daemon is running: 'ipfs daemon'
#    2. Check IPFS API endpoint configuration
#    3. Verify network connectivity to IPFS node
#    4. Retry the operation (attempt 1/3)
```

### Validação de Entrada

```bash
# Validação de senha fraca
./ipfs-encrypted-storage upload secret.txt
# Password: password123
# ❌ password entropy too low (32.4 bits) (operation: password_validation)
#
# 💡 Suggestions:
#    1. Use longer passwords for better security
#    2. Incorporate more character variety
#    3. Avoid predictable patterns
```

```bash
# Validação de CID inválido
./ipfs-encrypted-storage download invalid-cid --metadata doc.meta.json
# ❌ invalid CID format (operation: cid_validation) (resource: invalid-cid)
#
# 💡 Suggestions:
#    1. CID should start with valid prefix (Qm, bafy, bafz, baga, etc.)
#    2. Verify the CID was copied correctly
#    3. Ensure the content is properly pinned
```

## Error Handling

### Checking Operations

```bash
# Verificar se IPFS está rodando
curl -s localhost:5001/api/v0/id

# Testar workflow de criptografia/descriptografia
echo "test data" > test.txt
CID=$(./ipfs-encrypted-storage upload test.txt | grep "CID:" | cut -d' ' -f2)
./ipfs-encrypted-storage download "$CID" --metadata test.meta.json --output test_restored.txt
diff test.txt test_restored.txt && echo "Success: files match" || echo "Error: files differ"
```

### Troubleshooting

```bash
# Habilitar logging verboso
./ipfs-encrypted-storage --verbose upload file.txt

# Verificar logs do IPFS
ipfs log tail

# Verificar integridade do arquivo
./ipfs-encrypted-storage download <CID> --metadata <file>.meta.json --output verify.txt

# Verificar saúde da API
curl http://localhost:8080/api/v1/health/deep

# O transporte P2P real ainda não está implementado
echo "P2P: stub local"
```

## Novos Utilitários (v1.0.0+)

### Validação Robusta

A validação de arquivo, senha, CID e configuração é executada pelos comandos que consomem cada entrada. Não há subcomandos separados de validação.

### Conversões JSON Seguras

O sistema agora suporta múltiplos formatos de dados JSON:

```bash
# Arquivos de metadados suportam:
# - Base64: "SGVsbG8=" -> "Hello"
# - Hex: "48656c6c6f" -> "Hello"
# - Arrays: [72, 101, 108, 108, 111] -> "Hello"
# - Strings diretas
```

### Operações com Retry

O cliente possui utilitários internos de retry; a CLI atual não expõe flags para ajustar tentativas por comando.

## Configuration

### Opções Globais e Ambiente

```bash
# Definir endpoint do daemon IPFS e arquivo de configuração
./ipfs-encrypted-storage --ipfs-url "localhost:5001" --config "/path/to/config.json" list

# Apenas o servidor REST lê esta variável
export IPFS_API_KEY="load-from-your-secret-manager"
```

### Configuration File

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
  "validation": {
    "max_file_size": 104857600,
    "min_password_length": 8,
    "require_mixed_case": true,
    "require_numbers": true,
    "require_symbols": false,
    "common_passwords": [
      "password", "123456", "qwerty"
    ]
  },
  "api": {
    "enabled": true,
    "host": "localhost",
    "port": 8080
  },
  "p2p": {
    "listen_addr": "/ip4/0.0.0.0/tcp/0",
    "bootstrap_peers": []
  },
  "error_handling": {
    "max_retries": 3,
    "retry_delay": "1s",
    "enhanced_errors": true
  }
}
```

## Scripts Avançados

### Backup Automatizado com Validação

```bash
#!/bin/bash
# backup_advanced.sh - Backup com validação e verificação

set -e

BACKUP_DIR="$1"
if [ -z "$BACKUP_DIR" ]; then
    echo "Usage: $0 <backup_directory>"
    exit 1
fi

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="backup_$TIMESTAMP"

echo "🗂️  Criando backup de: $BACKUP_DIR"
echo "📅 Timestamp: $TIMESTAMP"

# Criar arquivo tar
tar -czf "${BACKUP_NAME}.tar.gz" "$BACKUP_DIR"

# O upload executa sua própria validação de entrada
echo "📤 Fazendo upload..."
CID=$(./ipfs-encrypted-storage upload "${BACKUP_NAME}.tar.gz" --description "Backup $TIMESTAMP de $BACKUP_DIR" | grep "CID:" | cut -d' ' -f2)

# Verificar upload
echo "🔍 Verificando upload..."
./ipfs-encrypted-storage verify "$CID"

# Limpar arquivo local
rm "${BACKUP_NAME}.tar.gz"

echo "✅ Backup concluído!"
echo "📋 CID: $CID"
echo "💾 Salve este CID para recuperação futura"
```

### Monitoramento de Saúde do Sistema

```bash
#!/bin/bash
# health_check.sh - Verificação abrangente da saúde do sistema

echo "🔍 Verificando saúde do sistema IPFS Encrypted Storage"
echo "=================================================="

# Verificar IPFS
echo "1. Verificando IPFS..."
if curl -s localhost:5001/api/v0/id > /dev/null; then
    echo "✅ IPFS está rodando"
else
    echo "❌ IPFS não está acessível"
    exit 1
fi

# Verificar API REST (se estiver rodando)
echo "2. Verificando API REST..."
if curl -s http://localhost:8080/api/v1/health > /dev/null; then
    echo "✅ API REST está rodando"

    # Verificar saúde detalhada
    HEALTH=$(curl -s http://localhost:8080/api/v1/health)
    IPFS_CONNECTED=$(echo "$HEALTH" | grep -o '"ipfs_connected":true' || true)

    if [ -n "$IPFS_CONNECTED" ]; then
        echo "✅ API conectada ao IPFS"
    else
        echo "⚠️  API não conectada ao IPFS"
    fi
else
    echo "⚠️  API REST não está rodando (inicie com: ./ipfs-encrypted-storage api)"
fi

# Registrar o estado P2P atual
echo "3. P2P: stub local; transporte de rede pendente"

# Verificar arquivos de configuração
echo "4. Verificando configuração..."
if [ -f "config.json" ]; then
    echo "✅ Arquivo de configuração encontrado"
else
    echo "⚠️  Arquivo config.json não encontrado (usando padrões)"
fi

echo ""
echo "🎉 Verificação de saúde concluída!"
```
