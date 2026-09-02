# API Programática

Este guia descreve apenas as APIs executáveis no estado atual do repositório. Os imports abaixo são destinados a código dentro deste módulo Go, cujo caminho é `ipfs-encrypted-storage`.

## Limites atuais

- Um daemon IPFS externo precisa ser fornecido para operações de rede e armazenamento.
- O pacote `p2p` é um stub local em memória. Ele não abre transporte, não descobre peers e não implementa DHT ou PubSub entre processos.
- A API HTTP é parcial; os endpoints ainda não implementados retornam HTTP 501.
- O pacote `zkp` é demonstrativo e usa parâmetros pequenos. Ele não é um sistema de autorização para produção, e prova de intervalo segura não está implementada.

## Criptografia

`EncryptWithMetadata` aplica a política de senha, deriva a chave, cifra com AES-GCM e assina os metadados. O mesmo password e a chave pública correspondente são necessários na leitura.

```go
package main

import (
    "bytes"
    "log"
    "os"

    "ipfs-encrypted-storage/src/encryption"
)

func main() {
    password := os.Getenv("ENCRYPTION_PASSWORD")
    if password == "" {
        log.Fatal("ENCRYPTION_PASSWORD is required")
    }

    plaintext := []byte("conteúdo sensível")
    publicKey, privateKey, err := encryption.GenerateKeyPair()
    if err != nil {
        log.Fatal(err)
    }

    ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
    if err != nil {
        log.Fatal(err)
    }

    recovered, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
    if err != nil {
        log.Fatal(err)
    }
    if !bytes.Equal(recovered, plaintext) {
        log.Fatal("round trip mismatch")
    }
}
```

Para uso de baixo nível, `DeriveKey` também retorna erro:

```go
salt, err := encryption.GenerateSalt()
if err != nil {
    return err
}
key, err := encryption.DeriveKey(password, salt, encryption.DefaultKeyDerivationConfig())
if err != nil {
    return err
}
```

## Cliente IPFS

O construtor valida a URL e retorna `(*IPFSClient, error)`. Disponibilidade real deve ser comprovada com `HealthCheck`.

```go
client, err := ipfs.NewIPFSClient("localhost:5001")
if err != nil {
    return err
}
defer client.Close()

if err := client.HealthCheck(); err != nil {
    return err
}

cid, err := client.AddFile(ciphertext, "encrypted.dat")
if err != nil {
    return err
}

downloaded, err := client.GetFile(cid)
if err != nil {
    return err
}
```

Os testes que precisam de um daemon real são opt-in por `IPFS_URL`. Quando a variável está definida, endpoint inválido ou indisponível faz o teste falhar.

## Stub P2P local

As assinaturas e publicações existem apenas dentro da mesma instância em memória.

```go
node, err := p2p.NewP2PNode("/ip4/127.0.0.1/tcp/0")
if err != nil {
    return err
}
defer node.Close()

subscription, err := node.SubscribeToTopic("events", func(peerID string, message []byte) error {
    fmt.Printf("mensagem local de %s: %s\n", peerID, message)
    return nil
})
if err != nil {
    return err
}
defer subscription.Cancel()

if err := node.PublishToTopic("events", []byte("ready")); err != nil {
    return err
}

if err := node.StoreValue("key", []byte("value")); err != nil {
    return err
}
value, err := node.GetValue("key")
```

`Bootstrap` e `ConnectToPeer` apenas validam multiaddresses e IDs. Eles não discam o endereço.

## Demonstração Schnorr

O transcript Schnorr deriva o desafio com Fiat-Shamir e pode vinculá-lo a um contexto. As funções de conveniência para metadados de acesso usam parâmetros intencionalmente pequenos, apenas para testes e aprendizado. O chamador deve usar `VerifyAccessProofFor` com recurso, usuário e permissões vindos de uma fonte confiável; os campos enviados na própria prova não são autoridade.

Uma prova de intervalo real não está disponível. Use uma biblioteca criptográfica auditada antes de tratar esse tipo de alegação como controle de acesso.

## API HTTP

O servidor exige `IPFS_API_KEY`. A rota de health é pública; as demais esperam o header `X-API-Key`.

```bash
export IPFS_API_KEY='use-a-secret-from-your-secret-manager'
go run ./src api --host localhost --port 8080

curl -H "X-API-Key: $IPFS_API_KEY" http://localhost:8080/api/v1/files
```

Consulte [api-rest.md](api-rest.md) para distinguir respostas reais, estáticas e HTTP 501.

## Contratos verificáveis

Os testes são a referência executável para comportamento e assinaturas atuais:

```bash
go test ./... -count=1
go test -race -coverpkg=./... ./...
go vet ./...
```

Veja especialmente `tests/encryption_test.go`, `tests/ipfs_client_test.go`, `tests/p2p_test.go`, `tests/zkp_test.go` e `src/api/server_simple_test.go`.
