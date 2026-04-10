# 📖 Índice Remissivo e Guia de Navegação (CROM-SEC)

Bem-vindo ao sumário executivo do **Crompressor Security Hub**.  
Este repositório deixou de ser um simples apanhado de PoCs e se tornou um **Livro Aberto de Arquitetura de Defesa e Proxying Avançado**.

Este documento é a sua bússola.

---

## 🧭 A Trilha do Arquiteto (Por onde eu começo?)

Se você é novo no conceito da arquitetura CROM (ou está tentando entender como "engolimos" sistemas inteiros), recomendamos veementemente ler **na ordem exata abaixo**:

1. **[Manual Mestre de Operação Prática](docs/MANUAL_DE_OPERACAO.md)**: Comece por aqui se só quer aprender a instalar cérebros, plugar o dashboard CLI e usar a infraestrutura no dia 1 sem teoria inútil.
2. **[Relatório de Análise Final (A Teoria Base)](docs/00_REPORT_ANALISE_FINAL.md)**: A introdução filosófica e teórica ao Cérebro e ao Virtual File System (VFS).
3. **[Guia Master: Drop-In Proxy e Segredos do Englobamento](docs/how_to_engulf/GUIDE_MASTER.md)**: Como, tecnicamente, injetamos a nuvem P2P nativa e bloqueamos o mundo de acessar suas portas expostas.
4. **[O Relatório Pentest Completo (v2)](docs/05_RELATORIO_PENTEST_COMPLETO.md)**: A prova de fogo e auditoria de 23 falhas de guerra superadas pela criptografia XOR P2P Orgânica.

*(Sente-se familiarizado com a teoria acima? Prossiga para os Laboratórios Práticos abaixo.)*

---

## 📚 Mapeamento da Documentação (`/docs`)

Aqui estão armazenadas as teses e estudos aprofundados sobre os quatro pilares da tecnologia que garante que nosso Proxy não é apenas um "tunel SSH moderno".

*   `00_` [O Motor Base e o LSH](docs/00_REPORT_ANALISE_FINAL.md)
*   `01_` [A Matemática do Drop-In Proxy](docs/01_DROP_IN_PROXY_TCP.md)
*   `02_` [Segurança de Dados e VFS em RAM](docs/02_DADOS_DOCKER_VFS.md)
*   `03_` [A Tática de Mutação dos Cérebros no Cliente](docs/03_MUTACAO_DE_CEREBROS_CLIENTES.md)
*   `04_` [Estratégias de Roteamento Alien](docs/04_ROTEAMENTO_MULTI_CAMADAS.md)
*   `05_` [O Relatório Pentest Hacker Oficial Mestre](docs/05_RELATORIO_PENTEST_COMPLETO.md)
*   `06_` [Retrospectiva & Escalabilidade Futura (Mixnets, Jitter, AI)](docs/06_RETROSPECTIVA_E_FUTURO.md)

---

## 🛠️ O Hub de Englobamento Prático (Cenários Reais)

Mostramos a Devs e SREs corporativos como aplicar a **Engenharia SecOps CROM** e plugar nosso Cérebro Omega nos sistemas mais frágeis da era moderna.

### 🚀 Instaladores e Ferramentas Práticas
*   📌 **[Painel Interativo de Terminal (TUI)](scripts/crom_master_dashboard.sh)**: A Central de Comando CROM. Administre conexões vivas, veja logs e edite túneis direto num menu de bash rápido (`./scripts/crom_master_dashboard.sh`).
*   📌 **[Instalador Nativo SystemD do Cérebro](scripts/deploy_omega_server.sh)**: Automatize a blindagem completa do seu Linux.
*   📌 **[Laboratórios Acionáveis Visuais](docs/how_to_engulf/runnable_labs/)**: Não só leia, execute. Simulações em tela com Pings.

### 🌐 Guias Stacks e Web3
*   📌 **[Cenário A: PHP Limpo + MySQL Antigo](docs/how_to_engulf/examples/01_PHP_e_MySQL_Legado.md)**
*   📌 **[Cenário B: Node.js, Banco de React SPA e Redis](docs/how_to_engulf/examples/02_Node_React_Redis.md)**
*   📌 **[Cenário C: Java Spring Boot O(1) e Arquivos SOAP Monstruosos](docs/how_to_engulf/examples/03_Java_SpringBoot_Corporativo.md)**
*   📌 **[Cenário Web3: Blindando DApps, IPFS, Geth e Redes DeFi p/ Block de MEVs](docs/how_to_engulf/examples/04_Web3_Blockchain_DApps.md)**

---

## 🧪 Laboratórios e Simuladores (`/simulators`)

Nós recusamos a explicação puramente teórica. O diretório de simuladores é onde compilamos pequenos trechos isolados da nossa engine principal de Go para provar que a teoria para de pé sob estresse físico absurdo.

*   **[O Motor Cérebro WASM (Injetado via Browser)](simulators/wasm_client_browser/README.md):** Uma joia da coroa. Veja na Prática (com Interface Gráfica Glassmorphism) como injetamos via JavaScript uma mutação no motor `window.fetch`, criptografando tudo no browser local das empresas!  
*   **O Universal Drop-in TCP (`simulators/dropin_tcp/`)**: Código nativo `proxy_in` e `proxy_out` onde aplicamos as WaitGroups Full-Duplex e a HMAC Crypto Seed-Driven.
*   **Armamento Hacker (`simulators/pentest/`)**: Aqui moram nossos inimigos, o `alien_sniffer` que fita a Mixnet e tenta vazar plaintexts, e o temido `tcp_cannon` que bombardeia com conexões sujas a cada 20 milissegundos tentando extrair dados ou causar *Crash*.

---

## ⚔️ Auditório de Tormenta (`/test_suites`)

Este é o local das nossas **21 Torres de Testes Automatizadas**. 
Onde chamamos o arquivo raiz `master_audit.sh` para disparar os cenários de guerra:

*   **11 a 20**: Testamos sobrecargas via Chunking, *Sybil DDoS Swarm*, falência de sistema (Split Brain), e falsificações na Camada Osi.
*   **[Suite 21 Especial: O Painel do Sistema Privado Local](test_suites/21_private_brain_system/README.md)**: A prova definitiva do **Silent Drop**. Testamos se um hacker com a porta crua mas sem a Seed é devidamente ignorado e aniquilado pelo Daemon sem respostas verbosas (Impedindo o *Banner Grabbing*).
