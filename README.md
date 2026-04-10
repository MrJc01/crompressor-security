# 🛡️ CROM-SEC - A Suite de Auditoria Cérebro Alpha/Omega

Bem-vindo ao repositório satélite de P&D (Pesquisa e Desenvolvimento) ofensivo do Projeto Crompressor.

Se você procura arquivos soltos ou testes, **PARE AQUI**.  
Este repositório consolidou-se em um livro prático de Arquitetura Hacker para proteger aplicações legadas (Java, PHP, Nodejs) contra ataques usando um proxy invisível de englobamento (O Drop-in P2P).

### ➡️ O Ponto de Entrada Oficial: [Leia o Índice (INDICE.md)](INDICE.md)

No nosso [Índice](INDICE.md) você encontrará o caminho do mapa. É obrigatório passar por ele caso deseje acessar os Simuladores WASM UI ou os Guias Executivos para Engenheiros de Software ("Como engolir seu projeto usando CROM").

---

## 🏆 Estado da Arte (Auditoria Mestre Final)

Executamos `master_audit.sh` testando o Cérebro Injetado na defesa contra simuladores de força bruta contínua (Denial-of-Service Cannon, Sybil Ghost Swarms e Fake Plaintexts). Nossos middlewares de criptografia XOR/HMAC alcançaram resultado impecável.

| Torres de Teste | Status Final de Batalha | Módulo Crítico Validado |
| ------------- | :-------------: | :------------- |
| HTTP/RPC/REST (Nominal) | ✅ PASS (21/21) | Criptografia In-flight O(1) Memory |
| Full-Duplex (WebSocket, DBs) | ✅ PASS | Goroutines `sync.WaitGroup` bidirecional |
| Silent Drop (Blindagem Anti-Hacker) | ✅ PASS | Aniquilação sem handshake TCP vazado |

![Vitória Absoluta](https://img.shields.io/badge/VULNERABILIDADES_LOCAIS-0-brightgreen.svg) ![Arquitetura P2P](https://img.shields.io/badge/DROP--IN_PROXY-WASM_READY-blue.svg)

> **Nota para SREs e Pentesters:** Todo nosso P&D em cima da criptografia do `proxy_universal_in` e `out` em Go está detalhado na seção **O Relatório Pentest** via `INDICE.md`. Use-o com sabedoria!
