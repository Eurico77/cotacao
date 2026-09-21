# Cotação

Desafio em Go com dois programas que se comunicam por HTTP para obter a cotação do dólar (USD-BRL):

- **server**: expõe `GET /cotacao` na porta `8080`, busca a cotação em `https://economia.awesomeapi.com.br/json/last/USD-BRL`, grava o resultado em um banco SQLite (`cotacao.db`) e devolve o JSON ao cliente.
- **client**: chama `http://localhost:8080/cotacao` e salva o valor de compra (`bid`) no arquivo `cotacao.txt`, no formato `Dólar: 5.117`.

## Requisitos

- Go 1.26.2 ou superior
- Um compilador C (gcc ou clang) com `CGO_ENABLED=1`, exigido pelo driver `github.com/mattn/go-sqlite3`. Sem CGO o projeto compila normalmente, mas o server encerra na inicialização
  - macOS: `xcode-select --install`
  - Debian/Ubuntu: `sudo apt install build-essential`
- `sqlite3` (CLI), opcional, apenas para consultar o banco
- Acesso à internet (o server consulta a AwesomeAPI)

## Estrutura

```
cmd/
  server/   main do servidor (e onde fica o cotacao.db)
  client/   main do cliente (e onde é gerado o cotacao.txt)
utils/
  server.go  servidor HTTP, chamada à API externa e gravação no SQLite
  client.go  requisição ao server e escrita do cotacao.txt
scripts/
  create_table.sh  cria o cotacao.db e a tabela cotacao
```

## Como rodar

Instale as dependências a partir da raiz do projeto:

```sh
go mod download
```

### 1. Server

Os caminhos do `cotacao.db` e do `cotacao.txt` são relativos ao diretório atual, então rode cada programa de dentro da sua pasta.

```sh
cd cmd/server
go run server.go
```

Na inicialização o server cria o `cotacao.db` e a tabela `cotacao`, caso ainda não existam (`CREATE TABLE IF NOT EXISTS`, então os dados de execuções anteriores são preservados). O banco não é versionado (`*db` está no `.gitignore`). Depois disso, o servidor fica ouvindo em `http://localhost:8080` e loga cada requisição com o tempo de resposta.

Se preferir criar o banco sem subir o server, use `./scripts/create_table.sh` (requer o `sqlite3` CLI).

### 2. Client

Em outro terminal, com o server no ar:

```sh
cd cmd/client
go run client.go
```

O valor do `bid` aparece no log e é gravado em `cmd/client/cotacao.txt`:

```text
Dólar: 5.117
```

### Testando o endpoint direto

```sh
curl http://localhost:8080/cotacao
```

```json
{
  "USDBRL": {
    "code": "USD",
    "codein": "BRL",
    "name": "Dólar Americano/Real Brasileiro",
    "high": "5.1442",
    "low": "5.1048",
    "varBid": "-0.0272",
    "pctChange": "-0.528748",
    "bid": "5.117",
    "ask": "5.1173",
    "timestamp": "1790000675",
    "create_date": "2026-09-21 11:24:35"
  }
}
```

### Consultando as cotações salvas

```sh
sqlite3 cmd/server/cotacao.db "SELECT bid, create_date FROM cotacao;"
```

## Timeouts

| Etapa                       | Limite |
| --------------------------- | ------ |
| Server → AwesomeAPI         | 200ms  |
| Server → gravação no SQLite | 10ms   |
| Client → Server             | 300ms  |

Os limites são curtos de propósito (fazem parte do desafio), então estourar um deles em conexões lentas é esperado. A primeira requisição depois de subir o server é a mais lenta, porque ainda abre a conexão TLS com a AwesomeAPI (~80ms contra ~20ms nas seguintes). Na maioria dos casos basta executar o client de novo.

## Problemas comuns

O client só mostra o sintoma; a causa costuma estar no log do server. Olhe sempre os dois terminais.

| Mensagem no client | Log do server | Causa | Solução |
| --- | --- | --- | --- |
| `dial tcp [::1]:8080: connect: connection refused` | — | O server não está rodando. | Suba o server (passo 1). Se ele encerrar logo após iniciar, veja "Server encerra ao iniciar" abaixo. |
| `context deadline exceeded` | pode não logar nada | O client desistiu após 300ms sem resposta do server. | Execute o client de novo. |
| `invalid character 'G' looking for beginning of value` | `Get "https://economia.awesomeapi.com.br/json/last/USD-BRL": context deadline exceeded` | A AwesomeAPI não respondeu em 200ms. O server devolve `500` com o erro em texto puro, e o client falha ao interpretá-lo como JSON. | Execute o client de novo e confira o acesso à internet. |
| `invalid character 'c' looking for beginning of value` | `context deadline exceeded` (sozinho, sem URL) | A gravação no SQLite passou de 10ms. O server devolve `500` com o erro em texto puro, e a cotação **não** é salva. | Execute o client de novo. |

### Server encerra ao iniciar

O server loga o motivo e volta para o prompt:

- **`listen tcp :8080: bind: address already in use`**: a porta `8080` já está em uso. Descubra quem ocupa a porta e encerre o processo:

  ```sh
  lsof -i :8080
  kill <PID>
  ```

- **`Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub`**: o server foi compilado sem CGO. O build **não falha**; o erro aparece ao criar a tabela na inicialização. Instale um compilador C e rode com `CGO_ENABLED=1 go run server.go`.
