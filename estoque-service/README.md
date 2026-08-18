# Serviço de Estoque

Microsserviço responsável pelo cadastro de produtos e pelo controle de
saldo em estoque. É o dono exclusivo dos dados de produto — nenhum outro
serviço acessa essa tabela diretamente; toda interação passa pela API HTTP
deste serviço.

Faz parte do sistema **Korp - Teste Técnico: Notas Fiscais**, junto com o
[`faturamento-service`](../faturamento-service) e o frontend Angular.

## Stack

- **Go** (`net/http` puro, sem framework HTTP externo)
- **GORM** como ORM, com **PostgreSQL**
- Banco de dados próprio (`estoque_db`), sem chaves estrangeiras
  compartilhadas com outros serviços — isolamento de dados é intencional

## Como rodar

```bash
go mod tidy
go run .
```

O serviço sobe em `http://localhost:8081`.

### Variáveis de ambiente

Copie `.env.example` para `.env` e preencha:

| Variável | Descrição | Padrão |
|---|---|---|
| `DB_HOST` | Host do Postgres | `localhost` |
| `DB_USER` | Usuário do banco | — |
| `DB_PASSWORD` | Senha do banco | — |

O nome do banco (`estoque_db`) e a porta do serviço (`8081`) estão fixos no
código — ajuste em `config/` e `main.go` se precisar mudar.

## Endpoints

### Públicos (usados pelo frontend)

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/produtos` | Lista todos os produtos cadastrados |
| `POST` | `/api/produtos` | Cadastra um novo produto |

**Exemplo — criar produto**

```http
POST /api/produtos
Content-Type: application/json

{
  "codigo": "PRF-M8-001",
  "descricao": "Parafuso sextavado M8",
  "saldo": 120
}
```

```json
{
  "id": 1,
  "codigo": "PRF-M8-001",
  "descricao": "Parafuso sextavado M8",
  "saldo": 120
}
```

### Internos (chamados pelo `faturamento-service`, não pelo frontend)

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/produtos/consultar` | Consulta produtos por ID, sem alterar saldo (só leitura) |
| `POST` | `/api/produtos/baixar-saldo` | Debita saldo de um ou mais produtos, de forma transacional |
| `POST` | `/api/produtos/repor-saldo` | Devolve saldo debitado (compensação de saga, quando o Faturamento não consegue concluir a criação/impressão da nota) |

**Exemplo — baixar saldo**

```http
POST /api/produtos/baixar-saldo
Content-Type: application/json

{
  "itens": [
    { "produtoId": 1, "quantidade": 5 },
    { "produtoId": 2, "quantidade": 2 }
  ]
}
```

Resposta de sucesso (`200`):
```json
{
  "itens": [
    { "id": 1, "descricao": "Parafuso sextavado M8", "saldo": 115 },
    { "id": 2, "descricao": "Porca de parafuso", "saldo": 348 }
  ]
}
```

## Tratamento de erros

Erros de negócio são tipados (`ErrSaldoInsuficiente`, `ErrProdutoNaoEncontrado`,
`ErrQuantidadeInvalida`) e mapeados explicitamente para status HTTP, com corpo
estruturado quando faz sentido pro cliente tomar uma decisão a partir dele:

| Situação | Status | Corpo |
|---|---|---|
| Saldo insuficiente pra baixa | `409 Conflict` | `{ "erro": "saldo_insuficiente", "produtoId", "produto", "disponivel", "solicitado" }` |
| Produto não encontrado | `400 Bad Request` | texto simples |
| Quantidade inválida (≤ 0) | `400 Bad Request` | texto simples |
| JSON malformado | `400 Bad Request` | texto simples |
| Erro inesperado (logado no console do serviço) | `500 Internal Server Error` | texto simples |

## Concorrência

`BaixarSaldo` e `ReporSaldo` rodam dentro de uma transação com
`SELECT ... FOR UPDATE` (lock de linha) sobre o produto sendo alterado. Isso
garante que duas notas fiscais criadas/impressas ao mesmo tempo não consigam
vender além do saldo real disponível — a segunda requisição espera a primeira
transação terminar antes de ler o saldo.

## Estrutura

```
estoque-service/
├── main.go                    # rotas, CORS
├── config/                    # conexão com o banco
└── internal/
    └── produto/
        ├── model.go            # struct Produto
        ├── errors.go           # erros de negócio tipados
        ├── repository.go       # acesso a dados (GORM)
        ├── service.go          # regras de negócio + transações
        └── handler.go          # HTTP handlers + serialização
```