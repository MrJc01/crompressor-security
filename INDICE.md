# 📖 Índice Remissivo e Guia de Navegação (CROM-SEC)

Bem-vindo ao sumário executivo do **Crompressor Security Hub**.  
Este repositório deixou de ser um simples apanhado de PoCs e se tornou um **Livro Aberto de Arquitetura de Defesa e Proxying Avançado**.

Este documento é a sua bússola.

---

## 🧭 A Trilha do Arquiteto (Por onde eu começo?)

Se você é novo no conceito da arquitetura CROM (ou está tentando entender como "engolimos" sistemas inteiros), recomendamos veementemente ler **na ordem exata abaixo**:

1. **[O GUIA DIDÁTICO CROM (Para Novos Devs)](docs/O_GUIA_DIDATICO_CROM.md)**: Se você não sabe absolutamente nada do que é esse projeto e quer aprender a tecnologia em linguagem coloquial, comece rigorosamente por aqui.
2. **[Manual Mestre de Operação Prática (SysAdmin)](docs/MANUAL_DE_OPERACAO.md)**: Comece por aqui se só quer aprender a instalar cérebros, plugar o dashboard CLI e usar no dia 1 sem teoria inútil.
3. **[Relatório de Análise Final (A Teoria Base)](docs/00_REPORT_ANALISE_FINAL.md)**: A introdução filosófica e teórica ao Cérebro e ao Virtual File System (VFS).
4. **[Guia Master: Drop-In Proxy e Segredos do Englobamento](docs/how_to_engulf/GUIDE_MASTER.md)**: Como, tecnicamente, injetamos a nuvem P2P nativa e bloqueamos o mundo de acessar portas expostas.
5. **[O Relatório Pentest Completo (v2)](docs/05_RELATORIO_PENTEST_COMPLETO.md)**: A prova de fogo e auditoria de 23 falhas de guerra superadas pela criptografia XOR P2P Orgânica.

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

Este é o local das nossas **23 Torres de Testes Automatizadas**. Onde rodamos nosso arsenal na unha.

> **📍 [LER O GUIA DAS 23 TORRES: O QUE CADA TESTE FAZ](test_suites/GUIA_DAS_23_TORRES.md)**

O arquivo raiz `master_audit.sh` executa todas as torres abaixo simultaneamente disparando cenários de guerra:

*   **[01_routing_nominal](test_suites/01_routing_nominal/README.md)**: Verifica se A consegue empacotar Bytes, enviar para B e sair limpo do outro lado da Nuvem.
*   **[02_pentest_mitm](test_suites/02_pentest_mitm/README.md)**: O Sniffer Alienígena da pasta tenta ler pedaços literais da memória. Falha graças ao XOR HMAC Mutante.
*   **[03_pentest_dos_cannon](test_suites/03_pentest_dos_cannon/README.md)**: O Canhão dispara 500 conexões maliciosas por milissegundo no Omega. Testa a estabilidade do CPU.
*   **[04_websocket_chat](test_suites/04_websocket_chat/README.md)**: Testa o "Full-Duplex" - bytes não empacotados em HTTP, mas sim uma tubulação TCP pura aberta eternamente rodando chat real-time.
*   **[05_php_fpm_cgi](test_suites/05_php_fpm_cgi/README.md)**: FastCGI. Testa o envio de arrays `$_POST` massivos sendo reconstruídos corretamente antes do Apache.
*   **[06_python_grpc](test_suites/06_python_grpc/README.md)**: Bufferização em Python. Streams binários HTTP/2 passando no CROM sem corromper pacotes.
*   **[07_postgres_pgwire](test_suites/07_postgres_pgwire/README.md)**: Testa um Backend DB cru engolido (`psql` login). Valida pacotes de Autenticação do PostgreSQL que não possuem Header HTTP.
*   **[08_redis_resp](test_suites/08_redis_resp/README.md)**: Engole o Redis Cache. Tráfego hiper-rápido operando em milissegundos sem adicionar latência perceptível (Overhead de Crypto = Zero).
*   **[09_iot_mqtt_broker](test_suites/09_iot_mqtt_broker/README.md)**: Internet das coisas. Valida como dispositivos pequenos conseguem persistir ping-pong por trás do escudo.
*   **[10_cplusplus_raw_tcp](test_suites/10_cplusplus_raw_tcp/README.md)**: Uma porta C++ pura. Nada de HTTP. Testa se o CROM destrói strings nulas (`\0`) que programadores C++ usam como EOF.
*   **[11_large_payload_chunking](test_suites/11_large_payload_chunking/README.md)**: Envia um arquivo brutal de 50MB no terminal para forçar o GoLang Chunk size a quebrar e fatiar pacotes.
*   **[12_high_concurrency](test_suites/12_high_concurrency/README.md)**: Não é o canhão, mas sim milhares de clientes legítimos (Alphas validos com senha) usando a tubulação simultaneamente `Goroutines`.
*   **[13_sybil_swarm_attack](test_suites/13_sybil_swarm_attack/README.md)**: Um enxame de mil Scanners usando chaves erradas por um Cópia mal feita do software para ver se dá Memory Leak por log excessivo.
*   **[14_silent_drop_validation](test_suites/14_silent_drop_validation/README.md)**: O principal tesouro. Garante O(1) Time. Se eu chutar o servidor, o Firewall do GO derruba sem gastar CPU em resposta de `Erro HTTP 403`.
*   **[15_split_brain_recovery](test_suites/15_split_brain_recovery/README.md)**: Derruba o Servidor CROM da porta e religa. O Alpha do cliente precisa saber reestabelecer o TCP Handshake sozinho.
*   **[16_nodejs_express_rest](test_suites/16_nodejs_express_rest/README.md)**: Testa envios JSON massivos com Headers variados (onde a Compressão LLM nativa do Gen-3 prova valor esmagando os atributos longos atômicos).
*   **[17_java_spring_boot_xml](test_suites/17_java_spring_boot_xml/README.md)**: O terror do mundo das empresas. Envia SOAP/XML via CROM. O Parser deve conseguir transitar intacto.
*   **[18_dns_hijack_spoofing](test_suites/18_dns_hijack_spoofing/README.md)**: Tentativa de passar Cabeçalhos de HOST falsificados dizendo estar em IP da intranet local.
*   **[19_payload_forgery](test_suites/19_payload_forgery/README.md)**: O Alien Sniffer intercepta o tráfego da rede, altera o último byte de uma senha Criptografada e tenta enviar pro Omega. O HMAC rejeita o pacote pela Checksum.
*   **[20_vfs_fd_exhaust](test_suites/20_vfs_fd_exhaust/README.md)**: Esgotamento de `File Descriptors`. Força o Linux ao talo para tentar *Crashar* o serviço de borda.
*   **[21_private_brain_system](test_suites/21_private_brain_system/README.md)**: Consolidou a tese do sistema privado, que nem sequer usa o framework clássico da empresa, fechando tráfego num loopback escuro invisível.
*   **[22_onion_multi_hop_route](test_suites/22_onion_multi_hop_route/README.md)**: Testa o repasse cebola cego. Um novo binário `proxy_onion_relay` repassa Sockets Alpha de um computador para o outro validando Dark Routing.
*   **[23_jitter_cover_traffic](test_suites/23_jitter_cover_traffic/README.md)**: O teste do Motor da Goroutine Anti-NSA. Garante que rajadas de tráfego de Lixo Hexadecimal (`JITT Magic Headers`) sejam identificadas na nuvem e Dropadas do canal principal sem interromper uma transação Web3 válida.
