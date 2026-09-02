# API REST - IPFS Encrypted Storage

Esta documentação descreve a API REST do IPFS Encrypted Storage para acesso programático ao sistema.

**Status:** Parcialmente implementada; consulte o status de cada endpoint
**Versão:** 1.0.0-alpha
**Framework:** Servidor HTTP simples (migração para Gin pendente)

## 📋 Visão Geral

A API REST fornece endpoints para operações de armazenamento criptografado, permitindo integração com aplicações externas, automação e desenvolvimento de interfaces web.

### Características Principais

- **Autenticação:** Baseada em API Key
- **Formatos:** JSON para requests/responses
- **CORS:** Suportado para aplicações web
- **Logging:** Middleware de logging integrado
- **Health Checks:** Endpoints de monitoramento

## 🔐 Autenticação

### API Key Authentication

Todos os endpoints (exceto health checks) requerem autenticação via header `X-API-Key`.

```bash
export IPFS_API_KEY="substitua-por-uma-chave-forte"
curl -H "X-API-Key: $IPFS_API_KEY" http://localhost:8080/api/v1/files
```

**Nota:** `IPFS_API_KEY` é obrigatória. Sem ela, o servidor recusa endpoints protegidos.

## 📡 Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### Health & Monitoring

#### GET `/api/v1/health`
Verifica saúde geral do sistema.

**Resposta:**
```json
{
  "status": "healthy",
  "ipfs_connected": true,
  "p2p_peers": 0,
  "uptime": 3600.5,
  "version": "1.0.0",
  "timestamp": 1704067200
}
```

#### GET `/api/v1/health/deep`
Health check detalhado com verificações específicas.

**Resposta:**
```json
{
  "status": "healthy",
  "checks": {
    "ipfs": true,
    "filesystem": true
  },
  "timestamp": 1704067200
}
```

#### GET `/api/v1/health/ready`
Verifica se o sistema está pronto para receber tráfego.

**Resposta:**
```json
{
  "status": "ready"
}
```

#### GET `/api/v1/health/live`
Verifica se o sistema está vivo (liveness probe).

**Resposta:**
```json
{
  "status": "alive"
}
```

### File Operations

#### GET `/api/v1/files`
Lista arquivos armazenados.

**Status:** ⚠️ Resposta estática; operações reais de arquivo ainda estão pendentes

**Resposta:**
```json
{
  "files": [],
  "count": 0
}
```

#### POST `/api/v1/files`
Faz upload de arquivo criptografado.

**Status:** ❌ Não Implementado

**Request:**
```json
{
  "file": "multipart/form-data file",
  "password": "string (optional)",
  "metadata": "object (optional)"
}
```

**Resposta Esperada:**
```json
{
  "cid": "Qm...",
  "filename": "document.pdf",
  "size": 1048576,
  "uploaded_at": 1704067200,
  "encrypted": true,
  "encryption_metadata": {
    "salt": "...",
    "signature": "...",
    "public_key": "..."
  }
}
```

#### GET `/api/v1/files/{cid}`
Faz download de arquivo por CID.

**Status:** ❌ Não Implementado

**Parâmetros:**
- `cid`: Content Identifier (path parameter)
- `password`: Senha para descriptografia (query parameter)
- `metadata`: Caminho do arquivo de metadados (query parameter)

**Resposta:** File stream ou JSON error

#### DELETE `/api/v1/files/{cid}`
Remove arquivo do IPFS (unpin).

**Status:** ❌ Não Implementado

**Parâmetros:**
- `cid`: Content Identifier (path parameter)

**Resposta:**
```json
{
  "success": true,
  "message": "File unpinned successfully"
}
```

### Endpoints do Stub P2P

#### POST `/api/v1/p2p/connect`
Endpoint reservado: a implementação atual retorna HTTP 501 e não realiza conexão.

**Status:** ❌ Não Implementado

**Request:**
```json
{
  "peer_address": "string",
  "timeout": 30
}
```

**Contrato planejado:**
```json
{
  "success": true,
  "peer_id": "12D3KooW...",
  "latency": 45.2
}
```

#### GET `/api/v1/p2p/peers`
Retorna uma lista estática vazia; não consulta um nó P2P.

**Status:** ⚠️ Resposta estática do stub (sem nó P2P real)

**Resposta:**
```json
{
  "peers": [],
  "count": 0
}
```

#### GET `/api/v1/p2p/status`
Retorna metadados estáticos do stub; não representa transporte nem conexões reais.

**Status:** ⚠️ Resposta estática do stub (sem nó P2P real)

**Resposta:**
```json
{
  "node_id": "simple-server",
  "listening_addresses": [],
  "connected_peers": 0
}
```

### System Monitoring

#### GET `/api/v1/metrics`
Retorna métricas do sistema.

**Status:** ✅ Implementado

**Resposta:**
```json
{
  "files_uploaded": 0,
  "total_bytes_uploaded": 0,
  "active_connections": 0,
  "memory_usage": 0.0,
  "cpu_usage": 0.0,
  "uptime": 3600.5,
  "request_count": 0,
  "error_count": 0
}
```

## 📊 Códigos de Status HTTP

| Código | Descrição |
|--------|-----------|
| 200 | OK - Operação bem-sucedida |
| 201 | Created - Recurso criado |
| 400 | Bad Request - Dados inválidos |
| 401 | Unauthorized - Autenticação necessária |
| 403 | Forbidden - Acesso negado |
| 404 | Not Found - Recurso não encontrado |
| 409 | Conflict - Conflito de recursos |
| 422 | Unprocessable Entity - Dados não processáveis |
| 500 | Internal Server Error - Erro interno |
| 503 | Service Unavailable - Serviço indisponível |

## 🛠️ Tratamento de Erros

### Estrutura de Erro

Todos os erros seguem um formato consistente:

```json
{
  "error": "Descrição do erro",
  "code": "ERROR_CODE",
  "timestamp": 1704067200
}
```

### Códigos de Erro

- `MISSING_FILE`: Arquivo não fornecido no upload
- `FILE_TOO_LARGE`: Arquivo excede limite de tamanho
- `MISSING_PASSWORD`: Senha obrigatória não fornecida
- `READ_ERROR`: Falha na leitura do arquivo
- `ENCRYPTION_ERROR`: Erro durante criptografia
- `UPLOAD_ERROR`: Falha no upload para IPFS
- `INVALID_INPUT`: Dados de entrada inválidos
- `AUTHENTICATION_FAILED`: Falha na autenticação

## 🔧 Middleware

### Authentication Middleware
- Valida API key em todos os endpoints protegidos
- Permite health checks sem autenticação

### CORS Middleware
- Suporta requisições cross-origin
- Headers configuráveis

### Logging Middleware
- Log estruturado de todas as requisições
- Métricas de performance (latência, status)

## 📝 Exemplos de Uso

### Verificar Saúde do Sistema

```bash
curl http://localhost:8080/api/v1/health
```

### Listar Arquivos (com autenticação)

```bash
curl -H "X-API-Key: $IPFS_API_KEY" \
     http://localhost:8080/api/v1/files
```

### Consultar a Resposta Estática do Stub P2P

```bash
curl -H "X-API-Key: $IPFS_API_KEY" \
     http://localhost:8080/api/v1/p2p/status
```

### Obter Métricas

```bash
curl -H "X-API-Key: $IPFS_API_KEY" \
     http://localhost:8080/api/v1/metrics
```

## 🚀 Inicialização da API

### Via Comando CLI

```bash
# Iniciar servidor API
./ipfs-encrypted-storage api

# Ou via comando específico
./ipfs-encrypted-storage api
```

### Configuração

O servidor inicia por padrão em `localhost:8080`. Para alterar:

```go
// Em server_simple.go, modificar:
return server.Start("localhost", 8080) // Altere host/port
```

## 🔄 Próximas Implementações

### Fase 1: Funcionalidades Core
- [ ] Upload de arquivos funcionais
- [ ] Download de arquivos funcionais
- [ ] Delete/unpin de arquivos

### Fase 2: Framework Gin
- [ ] Migração para Gin framework
- [ ] Middleware avançado
- [ ] Validação de entrada robusta

### Fase 3: Recursos Avançados
- [ ] Autenticação JWT
- [ ] Rate limiting
- [ ] Documentação OpenAPI/Swagger
- [ ] Webhooks e callbacks

### Fase 4: Produção
- [ ] HTTPS/TLS
- [ ] Gerenciamento de chaves seguro
- [ ] Logging avançado
- [ ] Métricas Prometheus

## 🔒 Segurança

### Considerações Atuais
- API key obrigatória via variável de ambiente
- Sem rate limiting implementado
- Sem validação de entrada avançada
- CORS permissivo

### Melhorias Planejadas
- Sistema de gerenciamento de API keys
- Rate limiting por IP/usuário
- Validação de entrada com sanitização
- CORS configurável por domínio

## 📈 Monitoramento

### Métricas Disponíveis
- Uptime do processo é calculado em tempo de execução
- Os demais campos atualmente retornam valores estáticos e não constituem telemetria operacional

### Health Checks
- `/health`: Status geral
- `/health/deep`: Verificações detalhadas
- `/health/ready`: Readiness probe
- `/health/live`: Liveness probe

## 🧪 Testes

### Testes Recomendados

```bash
# Testar health checks
curl http://localhost:8080/api/v1/health

# Testar autenticação
curl http://localhost:8080/api/v1/files  # Deve falhar sem API key

# Testar com autenticação
curl -H "X-API-Key: $IPFS_API_KEY" http://localhost:8080/api/v1/files
```

### Testes de Integração
- Verificar conectividade IPFS
- Verificar as respostas estáticas do stub P2P
- Validar criptografia/descriptografia
- Testar limites e validações

---

**Nota:** Esta documentação reflete o estado atual da API parcialmente implementada.
