# Sistema de Erros Aprimorado - IPFS Encrypted Storage

Esta documentação descreve o sistema avançado de tratamento de erros implementado no IPFS Encrypted Storage.

**Versão:** 1.0.0+
**Localização:** `src/errors/errors.go`

## 📋 Visão Geral

O sistema EnhancedError fornece tratamento contextual de erros com sugestões de recuperação, substituindo o tratamento básico de erros do Go.

### Características Principais

- **Contextualização:** Erros incluem informações sobre operação, recursos afetados
- **Sugestões de Recuperação:** Orientação acionável para resolução de problemas
- **Classificação Estruturada:** Códigos de erro padronizados
- **Formatação Amigável:** Mensagens claras para usuários finais
- **Integração Completa:** Usado em toda a aplicação

## 🏗️ Estrutura do Sistema

### EnhancedError

Estrutura core do sistema de erros aprimorado.

```go
type EnhancedError struct {
    Err     error           // Erro original
    Context *ErrorContext   // Contexto adicional
    Code    ErrorCode       // Código de erro padronizado
}
```

### ErrorContext

Informações contextuais sobre o erro.

```go
type ErrorContext struct {
    Operation   string   // Operação que falhou
    Resource    string   // Recurso afetado
    UserID      string   // Usuário relacionado
    FileName    string   // Arquivo relacionado
    CID         string   // CID relacionado
    RetryCount  int      // Contagem de tentativas
    Suggestions []string // Sugestões de recuperação
}
```

### ErrorCode

Códigos de erro padronizados para categorização.

```go
type ErrorCode int

const (
    ErrCodeUnknown ErrorCode = iota
    ErrCodeNetworkFailure
    ErrCodeInvalidInput
    ErrCodeAuthentication
    ErrCodePermissionDenied
    ErrCodeResourceNotFound
    ErrCodeQuotaExceeded
    ErrCodeInternalError
)
```

## 🚀 Como Usar

### Criando EnhancedError

#### Método Básico

```go
import "ipfs-encrypted-storage/src/errors"

// Criando erro aprimorado básico
err := errors.NewEnhancedError(
    fmt.Errorf("failed to connect to IPFS"),
    errors.ErrCodeNetworkFailure,
    &errors.ErrorContext{
        Operation: "IPFS Connection",
        Suggestions: []string{
            "Check if IPFS daemon is running",
            "Verify network connectivity",
            "Try again in a few moments",
        },
    },
)
```

#### Com Contexto Completo

```go
err := errors.NewEnhancedError(
    originalErr,
    errors.ErrCodeInvalidInput,
    &errors.ErrorContext{
        Operation:   "File Upload",
        FileName:    "document.pdf",
        CID:         "QmYwAPJzv5CZsnAzt1Q6g7j5f5f5f5f5f5f5f5f5f5f5f5",
        UserID:      "user123",
        RetryCount:  2,
        Suggestions: []string{
            "Check if file exists and is readable",
            "Verify file size (max 100MB)",
            "Ensure file is not corrupted",
        },
    },
)
```

### Acessando Informações

```go
if enhancedErr, ok := err.(*errors.EnhancedError); ok {
    // Código do erro
    fmt.Printf("Error Code: %v\n", enhancedErr.Code)

    // Mensagem formatada
    fmt.Printf("Error: %s\n", enhancedErr.Error())

    // Contexto
    if enhancedErr.Context != nil {
        fmt.Printf("Operation: %s\n", enhancedErr.Context.Operation)
        fmt.Printf("File: %s\n", enhancedErr.Context.FileName)
        fmt.Printf("CID: %s\n", enhancedErr.Context.CID)
    }

    // Sugestões
    suggestions := enhancedErr.Suggestions()
    if len(suggestions) > 0 {
        fmt.Println("Suggestions:")
        for i, suggestion := range suggestions {
            fmt.Printf("  %d. %s\n", i+1, suggestion)
        }
    }
}
```

## 🛠️ ErrorHandler

Utilitário centralizado para processamento de erros.

### Funcionalidades

- **Classificação Automática:** Determina códigos de erro baseados no conteúdo
- **Geração de Sugestões:** Sugestões contextuais baseadas no tipo de erro
- **Processamento Centralizado:** Lógica consistente de tratamento de erros

### Exemplo de Uso

```go
handler := &errors.ErrorHandler{MaxRetries: 3}

func processOperation() error {
    // Operação que pode falhar
    err := riskyOperation()

    if err != nil {
        // Processar erro através do handler
        enhancedErr := handler.HandleError(err, &errors.ErrorContext{
            Operation:  "File Processing",
            FileName:   "data.txt",
            RetryCount: attempt,
        })

        return enhancedErr
    }

    return nil
}
```

## 📊 Classificação de Erros

### Regras de Classificação Automática

O sistema classifica erros automaticamente baseado em padrões de texto:

```go
func (eh *ErrorHandler) classifyError(err error) ErrorCode {
    errStr := strings.ToLower(err.Error())

    switch {
    case strings.Contains(errStr, "connection refused") ||
         strings.Contains(errStr, "network") ||
         strings.Contains(errStr, "timeout"):
        return ErrCodeNetworkFailure

    case strings.Contains(errStr, "invalid") ||
         strings.Contains(errStr, "malformed") ||
         strings.Contains(errStr, "bad format"):
        return ErrCodeInvalidInput

    case strings.Contains(errStr, "unauthorized") ||
         strings.Contains(errStr, "authentication") ||
         strings.Contains(errStr, "password"):
        return ErrCodeAuthentication

    case strings.Contains(errStr, "permission") ||
         strings.Contains(errStr, "access denied"):
        return ErrCodePermissionDenied

    case strings.Contains(errStr, "not found") ||
         strings.Contains(errStr, "does not exist"):
        return ErrCodeResourceNotFound

    case strings.Contains(errStr, "quota") ||
         strings.Contains(errStr, "limit exceeded") ||
         strings.Contains(errStr, "too large"):
        return ErrCodeQuotaExceeded

    default:
        return ErrCodeUnknown
    }
}
```

## 💡 Sugestões Contextuais

### Sugestões por Tipo de Erro

#### NetworkFailure
```go
suggestions := []string{
    "Check your internet connection",
    "Verify IPFS daemon is running",
    "Try again in a few moments",
    "Check firewall settings",
}
```

#### InvalidInput
```go
suggestions := []string{
    "Verify input parameters are correct",
    "Check file format and size limits",
    "Ensure all required fields are provided",
}
```

#### Authentication
```go
suggestions := []string{
    "Verify your credentials",
    "Check password strength requirements",
    "Ensure key pair is valid",
}
```

### Sugestões Contextuais por Operação

```go
func (eh *ErrorHandler) generateSuggestions(code ErrorCode, context *ErrorContext) []string {
    baseSuggestions := []string{}

    // Sugestões específicas por operação
    switch code {
    case ErrCodeNetworkFailure:
        if context != nil && context.Operation == "IPFS Upload" {
            baseSuggestions = append(baseSuggestions,
                "Ensure IPFS daemon is running: 'ipfs daemon'",
                "Check IPFS API endpoint configuration",
                "Verify network connectivity to IPFS node")
        }
    }

    // Adicionar sugestões de retry se aplicável
    if context != nil && context.RetryCount < eh.MaxRetries {
        baseSuggestions = append(baseSuggestions,
            fmt.Sprintf("Retry the operation (attempt %d/%d)", context.RetryCount+1, eh.MaxRetries))
    }

    return baseSuggestions
}
```

## 🎨 Formatação de Mensagens

### FormatErrorMessage

Cria mensagens de erro amigáveis para usuários.

```go
func FormatErrorMessage(err *EnhancedError) string {
    var msg strings.Builder

    msg.WriteString("Error: ")
    msg.WriteString(err.Error())
    msg.WriteString("\n\n")

    if suggestions := err.Suggestions(); len(suggestions) > 0 {
        msg.WriteString("Suggestions:\n")
        for i, suggestion := range suggestions {
            msg.WriteString(fmt.Sprintf("   %d. %s\n", i+1, suggestion))
        }
        msg.WriteString("\n")
    }

    msg.WriteString("For more help, check the logs or documentation.")

    return msg.String()
}
```

### Exemplo de Saída Formatada

```
Error: failed to upload file: connection refused (operation: IPFS upload) (file: document.pdf)

Suggestions:
   1. Ensure IPFS daemon is running: 'ipfs daemon'
   2. Check IPFS API endpoint configuration
   3. Verify network connectivity to IPFS node
   4. Retry the operation (attempt 1/3)

For more help, check the logs or documentation.
```

## 🔄 Integração com CLI

### Display de Erro Amigável

```go
func displayError(err error) {
    if enhancedErr, ok := err.(*errors.EnhancedError); ok {
        fmt.Printf("❌ %s\n", enhancedErr.Error())

        if suggestions := enhancedErr.Suggestions(); len(suggestions) > 0 {
            fmt.Println("\n💡 Suggestions:")
            for i, suggestion := range suggestions {
                fmt.Printf("   %d. %s\n", i+1, suggestion)
            }
        }
    } else {
        fmt.Printf("❌ Error: %s\n", err.Error())
    }
}
```

## 🌐 Integração com API REST

### Resposta de Erro JSON

```go
func sendErrorResponse(c *gin.Context, err error) {
    var statusCode int
    var errorCode string
    var message string
    var suggestions []string

    if enhancedErr, ok := err.(*errors.EnhancedError); ok {
        message = enhancedErr.Error()
        suggestions = enhancedErr.Suggestions()

        // Mapear códigos para status HTTP
        switch enhancedErr.Code {
        case errors.ErrCodeInvalidInput:
            statusCode = http.StatusBadRequest
            errorCode = "INVALID_INPUT"
        case errors.ErrCodeAuthentication:
            statusCode = http.StatusUnauthorized
            errorCode = "AUTHENTICATION_FAILED"
        case errors.ErrCodeResourceNotFound:
            statusCode = http.StatusNotFound
            errorCode = "RESOURCE_NOT_FOUND"
        case errors.ErrCodeNetworkFailure:
            statusCode = http.StatusBadGateway
            errorCode = "NETWORK_ERROR"
        default:
            statusCode = http.StatusInternalServerError
            errorCode = "INTERNAL_ERROR"
        }
    } else {
        statusCode = http.StatusInternalServerError
        errorCode = "UNKNOWN_ERROR"
        message = err.Error()
    }

    response := gin.H{
        "error": message,
        "code": errorCode,
    }

    if len(suggestions) > 0 {
        response["suggestions"] = suggestions
    }

    c.JSON(statusCode, response)
}
```

### Exemplo de Resposta de Erro

```json
{
  "error": "failed to upload file: connection refused (operation: IPFS upload) (file: document.pdf)",
  "code": "NETWORK_ERROR",
  "suggestions": [
    "Ensure IPFS daemon is running: 'ipfs daemon'",
    "Check IPFS API endpoint configuration",
    "Verify network connectivity to IPFS node",
    "Retry the operation (attempt 1/3)"
  ]
}
```

## ⚡ Validação Pré-Operação

### ValidateOperation

Validação prévia de parâmetros antes da execução.

```go
func ValidateOperation(operation string, params map[string]interface{}) *EnhancedError {
    switch operation {
    case "encrypt":
        if password, ok := params["password"].(string); ok {
            if len(password) < 8 {
                return NewEnhancedError(
                    fmt.Errorf("password too short"),
                    ErrCodeInvalidInput,
                    &ErrorContext{
                        Operation: operation,
                        Suggestions: []string{
                            "Use a password with at least 8 characters",
                            "Include uppercase, lowercase, and digits",
                        },
                    },
                )
            }
        }

    case "upload":
        if fileSize, ok := params["file_size"].(int64); ok && fileSize > 100*1024*1024 {
            return NewEnhancedError(
                fmt.Errorf("file too large: %d bytes", fileSize),
                ErrCodeQuotaExceeded,
                &ErrorContext{
                    Operation: operation,
                    Suggestions: []string{
                        "Split large files into smaller chunks",
                        "Compress the file before uploading",
                        "Maximum file size is 100MB",
                    },
                },
            )
        }
    }

    return nil
}
```

## 🔍 RetryableError

Determina se um erro é recuperável e deve ser retentado.

```go
func RetryableError(err *EnhancedError) bool {
    switch err.Code {
    case ErrCodeNetworkFailure:
        return true
    case ErrCodeResourceNotFound:
        // Alguns erros 404 podem ser temporários
        return strings.Contains(err.Err.Error(), "temporary")
    default:
        return false
    }
}
```

## 📝 Logging

### LogError

Logging estruturado de erros.

```go
func LogError(err *EnhancedError, logger interface{}) {
    // Logging básico para desenvolvimento
    fmt.Printf("ERROR [%s]: %s\n", err.Code, err.Error())

    if suggestions := err.Suggestions(); len(suggestions) > 0 {
        fmt.Println("Suggestions:")
        for _, suggestion := range suggestions {
            fmt.Printf("  - %s\n", suggestion)
        }
    }
}
```

## 🧪 Testes

### Teste de EnhancedError

```go
func TestEnhancedError(t *testing.T) {
    // Criar erro aprimorado
    err := errors.NewEnhancedError(
        fmt.Errorf("test error"),
        errors.ErrCodeInvalidInput,
        &errors.ErrorContext{
            Operation: "test",
            Suggestions: []string{"try again"},
        },
    )

    // Verificar estrutura
    assert.Equal(t, errors.ErrCodeInvalidInput, err.Code)
    assert.Contains(t, err.Error(), "test error")
    assert.Contains(t, err.Error(), "(operation: test)")

    // Verificar sugestões
    suggestions := err.Suggestions()
    assert.Len(t, suggestions, 1)
    assert.Equal(t, "try again", suggestions[0])
}
```

### Teste de Classificação

```go
func TestErrorClassification(t *testing.T) {
    handler := &errors.ErrorHandler{}

    testCases := []struct {
        errorMsg string
        expected errors.ErrorCode
    }{
        {"connection refused", errors.ErrCodeNetworkFailure},
        {"invalid input", errors.ErrCodeInvalidInput},
        {"not found", errors.ErrCodeResourceNotFound},
        {"unauthorized", errors.ErrCodeAuthentication},
    }

    for _, tc := range testCases {
        code := handler.classifyError(fmt.Errorf(tc.errorMsg))
        assert.Equal(t, tc.expected, code, "Failed for: %s", tc.errorMsg)
    }
}
```

## 📊 Métricas de Qualidade

### Cobertura de Cenários

| Cenário de Erro | Tratamento | Status |
|----------------|------------|--------|
| Falhas de Rede | Classificação + sugestões específicas | ✅ Completo |
| Entrada Inválida | Contexto detalhado + validação | ✅ Completo |
| Autenticação | Sugestões de credenciais | ✅ Completo |
| Permissões | Context-aware suggestions | ✅ Completo |
| Recursos Não Encontrados | Sugestões de localização | ✅ Completo |
| Cotas Excedidas | Orientações de resolução | ✅ Completo |
| Erros Internos | Logging estruturado | ✅ Completo |

### Benefícios Obtidos

1. **Melhor UX:** Usuários recebem orientação clara para resolver problemas
2. **Debugging Aprimorado:** Contexto completo facilita troubleshooting
3. **Consistência:** Tratamento uniforme de erros em toda aplicação
4. **Manutenibilidade:** Centralização da lógica de tratamento de erros
5. **Monitoramento:** Códigos estruturados facilitam métricas e alertas

## 🔄 Migração de Código

### Antes (Tratamento Básico)

```go
// Código antigo
if err != nil {
    return fmt.Errorf("failed to upload file: %w", err)
}

// Uso
if err := uploadFile(file); err != nil {
    log.Printf("Upload failed: %v", err)
    return err
}
```

### Depois (EnhancedError)

```go
// Código novo
if err != nil {
    return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
        &errors.ErrorContext{
            Operation: "File Upload",
            FileName:  filePath,
            Suggestions: []string{
                "Check IPFS daemon status",
                "Verify file permissions",
                "Try uploading again",
            },
        })
}

// Uso aprimorado
if err := uploadFile(file); err != nil {
    if enhancedErr, ok := err.(*errors.EnhancedError); ok {
        displayError(enhancedErr) // Mostra erro + sugestões
    }
    return err
}
```

## 🚀 Exemplos Avançados

### Tratamento de Erro com Retry

```go
func uploadWithRetry(filePath string, maxRetries int) error {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := uploadFile(filePath)
        if err == nil {
            return nil
        }

        // Verificar se erro é retryable
        if enhancedErr, ok := err.(*errors.EnhancedError); ok {
            if !errors.RetryableError(enhancedErr) {
                return err // Não retryable
            }

            log.Printf("Attempt %d failed: %s", attempt, enhancedErr.Error())

            if attempt < maxRetries {
                // Exponential backoff
                time.Sleep(time.Duration(attempt) * time.Second)
                continue
            }
        }

        return err
    }
    return nil
}
```

### Agregação de Erros

```go
func processBatch(files []string) *errors.EnhancedError {
    var errors []string
    var processed int

    for _, file := range files {
        if err := processFile(file); err != nil {
            if enhancedErr, ok := err.(*errors.EnhancedError); ok {
                errors = append(errors, fmt.Sprintf("%s: %s", file, enhancedErr.Error()))
            } else {
                errors = append(errors, fmt.Sprintf("%s: %s", file, err.Error()))
            }
        } else {
            processed++
        }
    }

    if len(errors) > 0 {
        return errors.NewEnhancedError(
            fmt.Errorf("batch processing completed with errors: %d/%d succeeded", processed, len(files)),
            errors.ErrCodeInternalError,
            &errors.ErrorContext{
                Operation: "Batch Processing",
                Suggestions: []string{
                    "Check individual file errors above",
                    "Verify file formats and permissions",
                    "Consider processing files individually",
                },
            },
        )
    }

    return nil
}
```

---

**Nota:** O sistema EnhancedError foi implementado na versão 1.0.0+ como parte da iniciativa de melhoria de qualidade e experiência do usuário.
