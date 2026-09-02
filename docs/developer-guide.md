# Guia Técnico do Desenvolvedor - IPFS Encrypted Storage

Este guia fornece informações técnicas detalhadas para desenvolvedores que trabalham no projeto IPFS Encrypted Storage.

**Versão:** 1.0.0+
**Última Atualização:** Dezembro 2025
**Go Version:** 1.21+

## 📋 Visão Geral do Projeto

### Arquitetura Atual

```
src/
├── main.go              # Entry point refatorado (25 linhas)
├── cmd/                 # Comandos CLI modulares
│   ├── root.go         # Comando raiz e configuração
│   ├── upload.go       # Upload de arquivos
│   ├── download.go     # Download de arquivos
│   ├── list.go         # Listagem de arquivos
│   ├── delete.go       # Exclusão de arquivos
│   ├── p2p.go          # Operações do stub P2P local
│   ├── init.go         # Inicialização
│   ├── verify.go       # Verificação de arquivos
│   └── api.go          # Servidor API
├── api/                 # API REST (parcialmente implementada)
│   ├── server_simple.go # Servidor HTTP básico
│   └── models/         # Estruturas de request/response
├── utils/               # Utilitários centralizados
│   ├── jsonutils.go    # Conversões JSON seguras
│   ├── validation.go   # Validação robusta de entrada
│   └── retry.go        # Lógica de retry
├── errors/              # Sistema de erros aprimorado
│   └── errors.go       # EnhancedError com contexto
├── handlers/            # Parsers de mensagens do stub P2P
├── encryption/          # Módulo de criptografia
├── ipfs/               # Cliente IPFS
├── p2p/                # Stub local em memória (sem transporte)
├── config/             # Gerenciamento de configuração
├── did/                # Identidade decentralizada
├── zkp/                # Demonstração Schnorr experimental
├── contentaddressable/ # Content addressing
└── integrity/          # Verificação de integridade
```

## 🏗️ Arquitetura e Design

### Princípios de Design

1. **Separação de Responsabilidades:** Cada módulo tem uma responsabilidade clara
2. **Injeção de Dependências:** Interfaces para facilitar testes e mocking
3. **Tratamento de Erros Contextual:** EnhancedError com sugestões acionáveis
4. **Validação Robusta:** Entrada validada em todas as camadas
5. **Configuração Centralizada:** Sistema flexível de configuração

### Padrões Implementados

#### Command Pattern (CLI)
```go
// cmd/root.go
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ipfs-encrypted-storage",
        Short: "IPFS Encrypted Storage CLI",
    }

    cmd.AddCommand(
        NewUploadCmd(),
        NewDownloadCmd(),
        // ... outros comandos
    )

    return cmd
}
```

#### Factory Pattern (Componentes)
```go
// ipfs/ipfs_client.go
func NewIPFSClient(url string) (*IPFSClient, error) {
    client := &IPFSClient{
        url:    url,
        client: &http.Client{Timeout: 30 * time.Second},
    }

    // Validação e inicialização
    if err := client.HealthCheck(); err != nil {
        return nil, errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
            &errors.ErrorContext{
                Operation: "IPFS Client Initialization",
                Suggestions: []string{
                    "Ensure IPFS daemon is running",
                    "Check the provided URL",
                },
            })
    }

    return client, nil
}
```

#### Strategy Pattern (Validação)
```go
// utils/validation.go
func ValidateFile(filePath string, config *ValidationConfig) *errors.EnhancedError {
    // Estratégia de validação configurável
    if config == nil {
        config = DefaultValidationConfig()
    }

    // Múltiplas validações em sequência
    validations := []func() *errors.EnhancedError{
        func() *errors.EnhancedError { return validateFileExists(filePath) },
        func() *errors.EnhancedError { return validateFileSize(filePath, config) },
        func() *errors.EnhancedError { return validateFileName(filePath) },
    }

    for _, validation := range validations {
        if err := validation(); err != nil {
            return err
        }
    }

    return nil
}
```

## 🔧 Desenvolvimento

### Configuração do Ambiente

#### Pré-requisitos
- **Go 1.21+**
- **IPFS Daemon** rodando localmente
- **Git** para controle de versão

#### Instalação de Dependências
```bash
# Clonar repositório
git clone https://github.com/lucien-vallois/ipfs-encrypted-storage.git
cd ipfs-encrypted-storage

# Instalar dependências Go
go mod download

# Verificar instalação
go version  # Deve ser 1.21+
go mod verify
```

#### Configuração do IPFS
```bash
# Iniciar IPFS daemon
ipfs daemon

# Verificar se está rodando
curl http://localhost:5001/api/v0/id
```

### Estrutura de Diretórios

#### `src/cmd/` - Comandos CLI
- Cada comando em arquivo separado
- Uso do Cobra framework
- Flags e argumentos padronizados
- Integração com sistema de erros

#### `src/utils/` - Utilitários
- Funções puras e testáveis
- Sem dependências externas
- Reutilizáveis entre módulos

#### `src/errors/` - Sistema de Erros
- EnhancedError como tipo principal
- Context e sugestões integrados
- Classificação automática

### Convenções de Código

#### Nomenclatura
```go
// ✅ Bom
type SafeJSONConverter struct{}
func (c *SafeJSONConverter) Bytes(v interface{}, fieldName string) ([]byte, error)
func ValidatePassword(password string, config *ValidationConfig) *errors.EnhancedError

// ❌ Ruim
type jsonConverter struct{}  // Não exportado
func convertBytes(v interface{}) ([]byte, error)  // Muito genérico
func checkPass(pwd string) error  // Abreviação não clara
```

#### Tratamento de Erros
```go
// ✅ Sempre usar EnhancedError
func processFile(filePath string) error {
    if err := utils.ValidateFile(filePath, config); err != nil {
        return err // Já é EnhancedError com contexto
    }

    data, err := os.ReadFile(filePath)
    if err != nil {
        return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
            &errors.ErrorContext{
                Operation: "File Reading",
                FileName:  filePath,
                Suggestions: []string{
                    "Check if file exists",
                    "Verify read permissions",
                    "Ensure file is not locked",
                },
            })
    }

    return nil
}
```

#### Logging
```go
// ✅ Logging estruturado
import "github.com/sirupsen/logrus"

func uploadFile(filePath string) error {
    logrus.WithFields(logrus.Fields{
        "operation": "file_upload",
        "file":      filePath,
        "size":      fileSize,
    }).Info("Starting file upload")

    if err := uploadLogic(); err != nil {
        logrus.WithError(err).Error("File upload failed")
        return err
    }

    logrus.Info("File upload completed successfully")
    return nil
}
```

## 🧪 Testes

### Estratégia de Testes

#### Unitários (`*_test.go`)
- Cobertura mínima: 80%
- Mocks para dependências externas
- Testes de tabela para múltiplos cenários
- Testes de edge cases

#### Integração
- Testes com IPFS real (quando possível)
- Testes de API REST
- Testes de P2P (simulados)

#### E2E (End-to-End)
- Workflows completos
- Testes de CLI
- Validação de integração

### Exemplo de Teste Unitário

```go
// utils/validation_test.go
func TestValidatePassword(t *testing.T) {
    config := DefaultValidationConfig()

    testCases := []struct {
        name        string
        password    string
        shouldError bool
        errorCode   errors.ErrorCode
    }{
        {
            name:        "valid password",
            password:    "SecurePass123!",
            shouldError: false,
        },
        {
            name:        "too short",
            password:    "short",
            shouldError: true,
            errorCode:   errors.ErrCodeInvalidInput,
        },
        {
            name:        "common password",
            password:    "password123",
            shouldError: true,
            errorCode:   errors.ErrCodeInvalidInput,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidatePassword(tc.password, config)

            if tc.shouldError {
                assert.NotNil(t, err)
                if enhancedErr, ok := err.(*errors.EnhancedError); ok {
                    assert.Equal(t, tc.errorCode, enhancedErr.Code)
                    assert.NotEmpty(t, enhancedErr.Suggestions())
                }
            } else {
                assert.Nil(t, err)
            }
        })
    }
}
```

### Executando Testes

```bash
# Todos os testes
go test ./...

# Testes com cobertura
go test -cover ./...

# Testes de um pacote específico
go test ./src/utils -v

# Testes com race detection
go test -race ./...

# Benchmark
go test -bench=. ./src/encryption
```

### Mocks e Fixtures

```go
// mocks/ipfs_client_mock.go
type IPFSClientMock struct {
    AddFileFunc    func(data []byte, filename string) (string, error)
    GetFileFunc    func(cid string) ([]byte, error)
    HealthCheckFunc func() error
}

func (m *IPFSClientMock) AddFile(data []byte, filename string) (string, error) {
    if m.AddFileFunc != nil {
        return m.AddFileFunc(data, filename)
    }
    return "QmMockCID", nil
}
```

## 📊 Monitoramento e Observabilidade

### Métricas

#### Performance
- Latência de operações IPFS
- Tempo de criptografia/descriptografia
- Throughput de upload/download

#### Uso de Recursos
- Memória utilizada
- CPU usage
- Conexões de rede ativas

#### Erros
- Taxa de erro por operação
- Tipos de erro mais comuns
- Tempo médio de resolução

### Logging

#### Níveis de Log
```go
logrus.SetLevel(logrus.InfoLevel)  // Produção
logrus.SetLevel(logrus.DebugLevel) // Desenvolvimento

// Uso contextual
logrus.WithFields(logrus.Fields{
    "user_id":    userID,
    "operation":  "file_upload",
    "file_size":  fileSize,
}).Info("File upload started")
```

#### Estrutura de Logs
```json
{
  "level": "info",
  "msg": "File upload started",
  "time": "2025-12-18T10:30:00Z",
  "fields": {
    "user_id": "user123",
    "operation": "file_upload",
    "file_size": 1048576
  }
}
```

## 🔒 Segurança

### Práticas de Desenvolvimento Seguro

#### Criptografia
- Uso de algoritmos aprovados (AES-256-GCM, Ed25519)
- Nonces únicos para cada operação
- Sobrescrita best-effort de buffers de chave; cópias do runtime não são garantidamente apagadas

#### Validação de Entrada
- Validação em múltiplas camadas
- Sanitização de dados
- Limites de tamanho apropriados

#### Autenticação e Autorização
- API keys para acesso programático
- Validação de permissões
- Controle de acesso baseado em contexto

### Code Review Checklist

#### Segurança
- [ ] Dados sensíveis não logados
- [ ] Chaves criptográficas protegidas
- [ ] Validação de entrada adequada
- [ ] Tratamento seguro de erros

#### Qualidade
- [ ] Testes cobrindo cenários críticos
- [ ] Documentação atualizada
- [ ] Código seguindo convenções
- [ ] Performance não regrediu

## 🚀 Deployment

### Build

```bash
# Build otimizado para produção
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags)" \
    -o ipfs-encrypted-storage \
    ./src

# Build com debug symbols
go build -o ipfs-encrypted-storage-debug ./src
```

### Containerização

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./src

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Configuração de Produção

```json
{
  "ipfs": {
    "url": "https://ipfs.infura.io:5001",
    "timeout": "60s"
  },
  "api": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 8080
  },
  "validation": {
    "max_file_size": 1073741824,  // 1GB
    "min_password_length": 12
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

Forneça a credencial da API pelo ambiente (`IPFS_API_KEY`), nunca pelo arquivo de configuração.

## 🔄 CI/CD

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

    - name: Run tests
      run: go test -race -coverprofile=coverage.out ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  build:
    runs-on: ubuntu-latest
    needs: test

    steps:
    - uses: actions/checkout@v3

    - name: Build
      run: go build -o ipfs-encrypted-storage ./src

    - name: Upload artifact
      uses: actions/upload-artifact@v3
      with:
        name: ipfs-encrypted-storage
        path: ipfs-encrypted-storage
```

## 📚 Documentação

### Manutenção da Documentação

#### Arquivos de Documentação
- `docs/architecture.md` - Arquitetura do sistema
- `docs/api-rest.md` - Documentação da API REST
- `docs/utils.md` - Utilitários do sistema
- `docs/error-system.md` - Sistema de erros
- `docs/examples/` - Exemplos de uso
- `docs/developer-guide.md` - Este arquivo

#### Atualização Automática
```bash
# Gerar documentação da API (futuro)
swag init -g src/api/server_simple.go

# Verificar links quebrados
markdown-link-check docs/*.md docs/**/*.md
```

## 🐛 Debugging

### Ferramentas de Debug

#### Delve (Go Debugger)
```bash
# Instalar Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug aplicação
dlv debug ./src

# Attach a processo rodando
dlv attach $(pidof ipfs-encrypted-storage)
```

#### Profiling
```go
import _ "net/http/pprof"

// Acessar profiles em http://localhost:6060/debug/pprof/
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Troubleshooting Comum

#### Problemas de Conectividade IPFS
```bash
# Verificar se IPFS está rodando
curl http://localhost:5001/api/v0/id

# Verificar peers conectados
ipfs swarm peers

# Verificar configuração
ipfs config show
```

#### Problemas de Memória
```go
// Adicionar profiling de memória
import "runtime/pprof"

// Dump heap profile
f, _ := os.Create("heap.prof")
pprof.WriteHeapProfile(f)
f.Close()
```

#### Problemas de Performance
```go
// Benchmarking
go test -bench=. -benchmem ./src/encryption

// CPU profiling
go tool pprof cpu.prof
```

## 🤝 Contribuição

### Processo de Contribuição

1. **Fork** o repositório
2. **Criar branch** para feature/bugfix
3. **Implementar** mudanças com testes
4. **Executar** todos os testes
5. **Criar PR** com descrição detalhada
6. **Code review** e aprovação

### Padrões de Commit

```bash
# Formato: <tipo>(<escopo>): <descrição>

feat(validation): add CID format validation
fix(api): handle malformed JSON requests
docs(api): update REST API documentation
test(utils): add comprehensive validation tests
refactor(cmd): split monolithic upload command
```

### Pull Request Template

```markdown
## Descrição
[Descrição clara e concisa das mudanças]

## Tipo de Mudança
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Checklist
- [ ] Testes adicionados/atualizados
- [ ] Documentação atualizada
- [ ] Código segue convenções do projeto
- [ ] Linting passa
- [ ] Testes de integração passam

## Testes Realizados
[Descrever testes realizados]

## Notas Adicionais
[Qualquer informação adicional relevante]
```

## 📈 Roadmap de Desenvolvimento

### Próximas Implementações (2026)

#### Q1 2026: API REST Completa
- Migração para Gin framework
- Upload/download funcionais
- Autenticação JWT
- Documentação OpenAPI

#### Q2 2026: Escalabilidade
- Suporte a múltiplos nós IPFS
- Load balancing
- Cache distribuído
- Otimização de performance

#### Q3 2026: Recursos Avançados
- Interface web
- SDKs para múltiplas linguagens
- Integração com blockchains
- Suporte a NFT/storage

#### Q4 2026: Produção
- Auditoria de segurança profissional
- Compliance (GDPR, etc.)
- Multi-cloud deployment
- SLA e monitoring avançado

---

## 📞 Suporte

### Canais de Comunicação
- **Issues**: GitHub Issues para bugs e features
- **Discussions**: GitHub Discussions para questões gerais
- **Discord/Slack**: Para comunicação em tempo real

### Recursos Adicionais
- **Documentação**: `/docs` directory
- **Exemplos**: `/docs/examples` directory
- **Testes**: `/tests` directory
- **API**: `/docs/api-rest.md`

---

**Nota:** Este guia é mantido atualizado com as evoluções do projeto. Para dúvidas específicas, consulte a documentação relevante ou abra uma issue no repositório.
