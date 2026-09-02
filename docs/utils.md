# Utilitários - IPFS Encrypted Storage

Esta documentação descreve os utilitários centralizados implementados no sistema IPFS Encrypted Storage.

**Versão:** 1.0.0+
**Localização:** `src/utils/`

## 📋 Visão Geral

Os utilitários fornecem funcionalidades comuns e reutilizáveis, promovendo consistência e reduzindo duplicação de código.

## 🔧 SafeJSONConverter (`jsonutils.go`)

Utilitário para conversões seguras de tipos JSON com suporte a múltiplos formatos.

### Funcionalidades

- **Conversões Seguras:** Tratamento robusto de tipos JSON interface{}
- **Múltiplos Formatos:** Suporte a base64, hex, arrays, strings diretas
- **Tratamento de Erros:** Mensagens de erro descritivas
- **Validação de Tipos:** Verificação rigorosa de compatibilidade

### Métodos Disponíveis

#### `Bytes(v interface{}, fieldName string) ([]byte, error)`

Converte interface{} para []byte com múltiplos formatos suportados.

**Formatos Suportados:**
- `[]byte` direto
- `[]interface{}` (converte números float64 para bytes)
- `string` com decodificação automática:
  - Base64 (padrão)
  - Hexadecimal
  - Bytes raw (fallback)

**Exemplo:**
```go
converter := &utils.SafeJSONConverter{}

// De base64
data, err := converter.Bytes("SGVsbG8gV29ybGQ=", "data") // "Hello World"

// De array de números
data, err := converter.Bytes([]interface{}{72, 101, 108, 108, 111}, "data") // "Hello"

// De hex
data, err := converter.Bytes("48656c6c6f", "data") // "Hello"
```

#### `Uint32(v interface{}, fieldName string) (uint32, error)`

Converte interface{} para uint32.

**Tipos Suportados:**
- `uint32` direto
- `float64` (JSON numbers)
- `int` / `int64`
- `string` (parsing numérico)

#### `Uint8(v interface{}, fieldName string) (uint8, error)`

Converte interface{} para uint8 com mesma lógica do Uint32.

#### `String(v interface{}, fieldName string) (string, error)`

Converte interface{} para string.

#### `Bool(v interface{}, fieldName string) (bool, error)`

Converte interface{} para bool com suporte a múltiplos formatos.

**Valores Verdadeiros:** `true`, `"true"`, `"yes"`, `"1"`, `"on"`
**Valores Falsos:** `false`, `"false"`, `"no"`, `"0"`, `"off"`

### Exemplo de Uso Completo

```go
package main

import (
    "fmt"
    "log"
    "ipfs-encrypted-storage/src/utils"
)

func processMetadata(metadata map[string]interface{}) error {
    converter := &utils.SafeJSONConverter{}

    // Extrair salt (bytes)
    salt, err := converter.Bytes(metadata["salt"], "salt")
    if err != nil {
        return fmt.Errorf("failed to parse salt: %w", err)
    }

    // Extrair chave pública
    publicKeyBytes, err := converter.Bytes(metadata["public_key"], "public_key")
    if err != nil {
        return fmt.Errorf("failed to parse public_key: %w", err)
    }

    // Extrair versão
    version, err := converter.Uint32(metadata["version"], "version")
    if err != nil {
        return fmt.Errorf("failed to parse version: %w", err)
    }

    // Extrair flag de criptografia
    encrypted, err := converter.Bool(metadata["encrypted"], "encrypted")
    if err != nil {
        return fmt.Errorf("failed to parse encrypted flag: %w", err)
    }

    fmt.Printf("Salt: %x\n", salt)
    fmt.Printf("Public Key: %x\n", publicKeyBytes)
    fmt.Printf("Version: %d\n", version)
    fmt.Printf("Encrypted: %t\n", encrypted)

    return nil
}
```

## ✅ Sistema de Validação (`validation.go`)

Sistema abrangente de validação de entrada com tratamento de erros aprimorado.

### Configuração de Validação

```go
type ValidationConfig struct {
    MaxFileSize       int64
    MinPasswordLength int
    RequireMixedCase  bool
    RequireNumbers    bool
    RequireSymbols    bool
    CommonPasswords   []string
}

func DefaultValidationConfig() *ValidationConfig {
    return &ValidationConfig{
        MaxFileSize:       100 * 1024 * 1024, // 100MB
        MinPasswordLength: 8,
        RequireMixedCase:  true,
        RequireNumbers:    true,
        RequireSymbols:    false,
        CommonPasswords: []string{
            "password", "123456", "qwerty", /* ... */
        },
    }
}
```

### Validação de Arquivos

#### `ValidateFile(filePath string, config *ValidationConfig) *errors.EnhancedError`

Realiza validação abrangente de arquivos.

**Verificações:**
- ✅ Existência do arquivo
- ✅ Permissões de leitura
- ✅ Arquivo vs diretório
- ✅ Tamanho máximo (100MB padrão)
- ✅ Nome do arquivo (caracteres válidos, comprimento)

**Exemplo:**
```go
config := utils.DefaultValidationConfig()
err := utils.ValidateFile("/path/to/document.pdf", config)
if err != nil {
    // Erro com sugestões de correção
    fmt.Println(err.Error())
    for _, suggestion := range err.Suggestions() {
        fmt.Printf("💡 %s\n", suggestion)
    }
    return err
}
```

### Validação de Senhas

#### `ValidatePassword(password string, config *ValidationConfig) *errors.EnhancedError`

Validação robusta de senhas com múltiplas camadas.

**Verificações:**
- ✅ Comprimento mínimo
- ✅ Senhas comuns (blacklist)
- ✅ Requisitos de caracteres:
  - Maiúsculas e minúsculas (se `RequireMixedCase`)
  - Números (se `RequireNumbers`)
  - Símbolos (se `RequireSymbols`)
- ✅ Entropia de senha (mínimo 50 bits)

**Cálculo de Entropia:**
```go
// Entropia = comprimento × log₂(tamanho_charset)
// Exemplo: senha "Hello123" (8 chars, charset ~52) = ~45.5 bits
entropy := calculatePasswordEntropy(password)
```

**Exemplo:**
```go
password := "MySecurePass123"
err := utils.ValidatePassword(password, config)
if err != nil {
    fmt.Printf("❌ Password invalid: %s\n", err.Error())
    fmt.Println("Suggestions:")
    for _, suggestion := range err.Suggestions() {
        fmt.Printf("  • %s\n", suggestion)
    }
}
```

### Validação de CID

#### `ValidateCID(cid string) *errors.EnhancedError`

Validação de Content Identifiers do IPFS.

**Verificações:**
- ✅ Presença do CID
- ✅ Prefixos válidos: `Qm`, `bafy`, `bafz`, `baga`
- ✅ Comprimento mínimo

**Exemplo:**
```go
cid := "QmYwAPJzv5CZsnAzt1Q6g7j5f5f5f5f5f5f5f5f5f5f5f5"
err := utils.ValidateCID(cid)
if err != nil {
    fmt.Printf("❌ Invalid CID: %s\n", err.Error())
}
```

### Validação de Endereços Peer

#### `ValidatePeerAddress(peerAddr string) *errors.EnhancedError`

Validação de multiendereços libp2p.

**Verificações:**
- ✅ Presença do endereço
- ✅ Prefixo válido (`/ip4/` ou `/ip6/`)
- ✅ Estrutura completa (protocolo/IP/porta/peerID)

**Formato Esperado:**
```
/ip4/192.168.1.100/tcp/4001/p2p/12D3KooWAbcDef123...
```

**Exemplo:**
```go
peerAddr := "/ip4/192.168.1.100/tcp/4001/p2p/peer123"
err := utils.ValidatePeerAddress(peerAddr)
```

### Validação de Endpoints IPFS

#### `ValidateIPFSEndpoint(endpoint string) *errors.EnhancedError`

Validação de URLs de endpoints IPFS.

**Verificações:**
- ✅ Presença do endpoint
- ✅ Formato de URL válido
- ✅ Protocolo suportado (http/https)
- ✅ Hostname presente

**Exemplos Válidos:**
- `localhost:5001`
- `http://localhost:5001`
- `https://ipfs.infura.io:5001`

## 🔄 Sistema de Retry (`retry.go`)

Utilitário para lógica de retry com backoff exponencial.

### Funcionalidades

- **Backoff Exponencial:** Aumento progressivo do intervalo
- **Jitter:** Randomização para evitar thundering herd
- **Configuração Flexível:** Personalização de parâmetros
- **Contexto:** Suporte a context.Context para cancelamento

### Estrutura de Configuração

```go
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
    Jitter       bool
}
```

### Exemplo de Uso

```go
import "ipfs-encrypted-storage/src/utils"

// Configuração de retry
config := &utils.RetryConfig{
    MaxAttempts:  5,
    InitialDelay: time.Second,
    MaxDelay:     time.Minute,
    Multiplier:   2.0,
    Jitter:       true,
}

// Função com retry
result, err := utils.WithRetry(ctx, config, func() (interface{}, error) {
    // Operação que pode falhar
    return ipfsClient.AddFile(data, filename)
})

if err != nil {
    log.Printf("Operation failed after %d attempts: %v", config.MaxAttempts, err)
}
```

## 🎯 Integração com EnhancedError

Todos os utilitários de validação integram-se com o sistema EnhancedError:

```go
// Validação retorna EnhancedError com contexto
err := utils.ValidateFile("document.pdf", config)
if err != nil {
    // Contexto da operação
    fmt.Printf("Operation: %s\n", err.Context.Operation)

    // Sugestões de recuperação
    for _, suggestion := range err.Suggestions() {
        fmt.Printf("💡 %s\n", suggestion)
    }

    // Código de erro estruturado
    fmt.Printf("Error Code: %v\n", err.Code)
}
```

## 📊 Cobertura de Validação

| Tipo de Entrada | Validações Implementadas | Status |
|----------------|------------------------|--------|
| Arquivos | Existência, tamanho, nome, permissões | ✅ Completo |
| Senhas | Comprimento, complexidade, entropia, blacklist | ✅ Completo |
| CID | Formato, prefixos, comprimento | ✅ Completo |
| Peer Address | Formato multiaddr, estrutura | ✅ Completo |
| IPFS Endpoint | URL format, protocolo, hostname | ✅ Completo |
| JSON Types | Conversões seguras múltiplos formatos | ✅ Completo |

## 🔧 Configuração Padrão

```go
// Configuração padrão recomendada
validationConfig := utils.DefaultValidationConfig()

// Personalização se necessário
validationConfig.MaxFileSize = 50 * 1024 * 1024 // 50MB
validationConfig.MinPasswordLength = 12         // Senhas mais longas
```

## 🧪 Testes e Qualidade

### Testes Implementados
- ✅ Testes unitários para todas as funções
- ✅ Testes de edge cases
- ✅ Testes de integração com EnhancedError
- ✅ Validação de mensagens de erro

### Exemplos de Teste

```go
func TestValidatePassword(t *testing.T) {
    config := DefaultValidationConfig()

    // Teste senha válida
    err := ValidatePassword("SecurePass123!", config)
    assert.Nil(t, err)

    // Teste senha curta
    err = ValidatePassword("short", config)
    assert.NotNil(t, err)
    assert.Contains(t, err.Error(), "too short")
    assert.Contains(t, err.Suggestions()[0], "at least 8 characters")
}
```

## 🚀 Boas Práticas

### 1. Sempre Use Configuração
```go
// ✅ Bom
config := utils.DefaultValidationConfig()
err := utils.ValidateFile(filePath, config)

// ❌ Ruim - sem configuração
err := utils.ValidateFile(filePath, nil)
```

### 2. Trate Erros EnhancedError
```go
// ✅ Bom
if err := utils.ValidateFile(filePath, config); err != nil {
    log.Printf("Validation failed: %s", err.Error())
    for _, suggestion := range err.Suggestions() {
        log.Printf("Suggestion: %s", suggestion)
    }
    return err
}

// ❌ Ruim - tratamento básico
if err := someValidation(filePath); err != nil {
    return err
}
```

### 3. Use Context para Retry
```go
// ✅ Bom
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := utils.WithRetry(ctx, config, operation)

// ❌ Ruim - sem timeout
result, err := utils.WithRetry(context.Background(), config, operation)
```

## 🔄 Migração de Código Legado

### Antes (Código Antigo)
```go
// Código antigo com validação manual
if len(password) < 8 {
    return fmt.Errorf("password too short")
}
```

### Depois (Novo Sistema)
```go
// Novo sistema com EnhancedError
config := utils.DefaultValidationConfig()
if err := utils.ValidatePassword(password, config); err != nil {
    return err // Já inclui contexto e sugestões
}
```

## 📈 Benefícios Obtidos

1. **Consistência:** Validações uniformes em toda aplicação
2. **Manutenibilidade:** Centralização da lógica de validação
3. **Usabilidade:** Mensagens de erro com sugestões acionáveis
4. **Robustez:** Tratamento abrangente de edge cases
5. **Testabilidade:** Funções puras fáceis de testar
6. **Reutilização:** Utilitários compartilhados entre módulos

---

**Nota:** Estes utilitários foram implementados como parte da refatoração v1.0.0+ para melhorar a qualidade e manutenibilidade do código.
