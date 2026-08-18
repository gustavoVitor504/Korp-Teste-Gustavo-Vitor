# Controle de Notas Fiscais — Frontend Angular

Frontend Angular (standalone components, Angular 18) para consumir os
microsserviços em Go: **Serviço de Estoque** (produtos) e **Serviço de
Faturamento** (notas fiscais).

## Rodando o projeto

```bash
npm install
npm start
```

A aplicação sobe em `http://localhost:4200`. Os dois backends Go precisam
estar rodando:

- Serviço de Estoque → `http://localhost:8081`
- Serviço de Faturamento → `http://localhost:8082`

## Configuração da API

As URLs base ficam em `src/environments/environment.ts`:

```ts
export const environment = {
  production: false,
  estoqueApiUrl: 'http://localhost:8081/api',
  faturamentoApiUrl: 'http://localhost:8082/api',
};
```

## Fluxo de negócio implementado

1. **Cadastro de produto** → `POST {estoqueApiUrl}/produtos`
2. **Cadastro de nota fiscal** → `POST {faturamentoApiUrl}/notas-fiscais`
   — nasce sempre com status `Aberta`. **Não debita estoque ainda.**
3. **Impressão da nota** → `POST {faturamentoApiUrl}/notas-fiscais/{id}/imprimir`
   — só aceito se a nota estiver `Aberta`. O Faturamento chama o Estoque
   pra debitar o saldo, e só então fecha a nota (`Fechada`). Erros
   possíveis:
   - `409 saldo_insuficiente` — algum item não tem saldo suficiente
   - `409 nota_nao_aberta` — a nota já foi impressa ou está cancelada
   - `503 estoque_indisponivel` — o Serviço de Estoque está fora do ar

O botão "Imprimir" na tela de Notas Fiscais só aparece pra notas `Aberta`,
mostra "Imprimindo…" enquanto a chamada está em andamento, e a lista é
recarregada ao final pra refletir o novo status e o saldo atualizado dos
produtos.

## Estrutura

```
src/app/
  models/            # Produto, NotaFiscal, StatusNotaFiscal, erros de negócio
  services/           # ProdutoService (Estoque), NotaFiscalService (Faturamento)
  pages/
    produto-form/      # Cadastro + listagem de produtos
    nota-fiscal-form/   # Cadastro de nota (sempre Aberta) + ação de imprimir
    impressao/          # Relatório de produtos e notas, com botão Imprimir (browser)
  app.routes.ts        # /produtos, /notas-fiscais, /impressao
```

