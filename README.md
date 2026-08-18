# Korp - Teste Técnico: Sistema de Notas Fiscais

## Arquitetura
- `estoque-service` (porta 8081) — cadastro de produtos e controle de saldo
- `faturamento-service` (porta 8082) — cadastro e impressão de notas fiscais
- `frontend` — Angular, consome os dois serviços

Bancos de dados separados por serviço (sem FK cruzando serviços);
comunicação entre eles via HTTP, com padrão saga + compensação na
criação/impressão de nota fiscal (ver `faturamento-service/internal/notafiscal/service.go`).

## Como rodar
1. `estoque-service`: copie `.env.example` para `.env`, preencha, `go run .`
2. `faturamento-service`: mesma coisa, `go run .`
3. `frontend`: `npm install && ng serve`

## Detalhamento técnico
Ciclos de vida: 1 ciclo (NgOnInit)
Framework: RxJS (usado pipe.map() para função Normalizar())

## Vídeo de apresentação
https://drive.google.com/file/d/1mG82p8NLyudUd6FdNNSBh8VuBP0Hyp1Q/view?usp=drive_link