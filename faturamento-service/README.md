# Serviço de Faturamento

Microsserviço responsável pelo cadastro e impressão de notas fiscais. Não
tem acesso direto aos dados de produto — toda consulta e alteração de saldo
passa por chamadas HTTP ao [`estoque-service`](../estoque-service).

Faz parte do sistema **Korp - Teste Técnico: Notas Fiscais**, junto com o
`estoque-service` e o frontend Angular.

## Stack

- **Go** (`net/http` puro, sem framework HTTP externo)
- **GORM** como ORM, com **PostgreSQL**
- Banco de dados próprio (`nota_fiscal_db`), sem chave estrangeira para
  produtos — a integridade entre os dois serviços é garantida pela chamada
  HTTP ao Estoque, não pelo banco

## Como rodar

Depende do `estoque-service` estar rodando (`http://localhost:8081` por
padrão) — sem ele, criar/imprimir nota retorna erro de indisponibilidade
(ver seção de erros abaixo).

```bash
go mod tidy
go run .
```

O serviço sobe em `http://localhost:8082`.

### Variáveis de ambiente

Copie `.env.example` para `.env` e preencha:

| Variável | Descrição | Padrão |
|---|---|---|
| `DB_HOST` | Host do Postgres | `localhost` |
| `DB_USER` | Usuário do banco | — |
| `DB_PASSWORD` | Senha do banco | — |
| `ESTOQUE_SERVICE_URL` | Base URL do `estoque-service` | `http://localhost:8081/api` |

O nome do banco (`nota_fiscal_db`) e a porta do serviço (`8082`) estão
fixos no código.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/notas-fiscais` | Lista todas as notas fiscais, com seus itens |
| `POST` | `/api/notas-fiscais` | Cria uma nova nota fiscal, sempre com status `aberta` |
| `POST` | `/api/notas-fiscais/{id}/imprimir` | Fecha a nota e debita o estoque dos itens |

**Exemplo — criar nota fiscal**

```http
POST /api/notas-fiscais
Content-Type: application/json

{
  "itens": [
    { "produtoId": 1, "quantidade": 5 },
    { "produtoId": 2, "quantidade": 2 }
  ]
}
```

```json
{
  "id": 4,
  "numero": 4,
  "numeroFormatado": "000000004",
  "status": "aberta",
  "itens": [
    { "id": 8, "notaFiscalId": 4, "produtoId": 1, "descricao": "Parafuso sextavado M8", "quantidade": 5 },
    { "id": 9, "notaFiscalId": 4, "produtoId": 2, "descricao": "Porca de parafuso", "quantidade": 2 }
  ]
}
```

A criação da nota **não altera saldo** — só reserva/registra os itens.
O número (`numero`) é sequencial, calculado automaticamente a cada criação.

**Exemplo — imprimir nota**

```http
POST /api/notas-fiscais/4/imprimir
```

```json
{
  "id": 4,
  "numero": 4,
  "numeroFormatado": "000000004",
  "status": "fechada",
  "itens": [ /* ... */ ]
}
```

Só é aceito se a nota estiver `aberta`. Ao imprimir, o saldo dos produtos é
debitado no `estoque-service` e o status muda para `fechada`.

## Fluxo de negócio — por que criar e imprimir são etapas separadas

| Ação | O que acontece |
|---|---|
| Criar nota | Nasce `aberta`. Não mexe em estoque. |
| Imprimir nota | Só permitido se `aberta`. Debita saldo no Estoque, depois fecha a nota (`fechada`). |

Isso espelha o requisito de negócio: o estoque só deve ser comprometido no
momento em que a nota é efetivamente emitida/impressa, não na simples
criação/rascunho.

## Padrão saga — como a criação/impressão lida com dois bancos diferentes

Como `nota_fiscal_db` e `estoque_db` são bancos **fisicamente separados**,
não existe uma transação única do GORM cobrindo as duas operações
(criar/atualizar nota + debitar saldo). O fluxo usa uma saga simples com
compensação:

1. **Faturamento chama o Estoque** (`/produtos/baixar-saldo` na impressão,
   `/produtos/consultar` na criação, só leitura) — se o Estoque estiver
   fora do ar, a operação falha aqui e nada é gravado no lado do
   Faturamento.
2. **Faturamento grava a mudança no seu próprio banco** (transação local
   com `db.Transaction`, cobrindo só as tabelas dele).
3. **Se o passo 2 falhar** depois que o passo 1 já debitou o saldo, o
   Faturamento chama `/produtos/repor-saldo` no Estoque pra desfazer a
   baixa — evitando que o produto "suma" do estoque sem uma nota
   correspondente de verdade.

Essa lógica está em `internal/notafiscal/service.go`, no método `Imprimir`.

## Tratamento de erros

| Situação | Status | Corpo |
|---|---|---|
| Nota já foi impressa/cancelada | `409 Conflict` | `{ "erro": "nota_nao_aberta", "mensagem" }` |
| Saldo insuficiente em algum item | `409 Conflict` | `{ "erro": "saldo_insuficiente", "produtoId", "produto", "disponivel", "solicitado", "mensagem" }` |
| Estoque fora do ar / inacessível | `503 Service Unavailable` | `{ "erro": "estoque_indisponivel", "mensagem" }` |
| Nenhum item informado / quantidade inválida | `400 Bad Request` | texto simples |
| Nota não encontrada | `404 Not Found` | texto simples |
| Erro inesperado (logado no console do serviço) | `500 Internal Server Error` | texto simples |

O `503` é o cenário de "falha de microsserviço" do teste: derrube o
`estoque-service` e tente criar ou imprimir uma nota pra reproduzir.

## Estrutura

```
faturamento-service/
├── main.go                    # rotas, CORS
├── config/                    # conexão com o banco
└── internal/
    ├── estoqueclient/          # cliente HTTP pro estoque-service
    │   └── client.go            # BaixarSaldo, ReporSaldo, Consultar
    └── notafiscal/
        ├── model.go             # NotaFiscal, NotaFiscalItem
        ├── errors.go            # erros de negócio tipados
        ├── repository.go        # acesso a dados (GORM)
        ├── service.go           # regras de negócio + saga
        └── handler.go           # HTTP handlers + serialização
```