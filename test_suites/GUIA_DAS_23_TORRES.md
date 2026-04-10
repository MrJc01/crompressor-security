# 🏰 O GUIA DAS 23 TORRES (Suítes de Batalha CROM)

Este repositório contém 23 simulações militares exatas que provam a impenetratibilidade da engine GoLang do CROM-SEC. Abaixo está a documentação consolidada de todas as sub-pastas e do que cada teste foi programado para validar.

## 🟢 1. O Nível Fundacional (Criptografia e Sockets)
*   **01_routing_nominal:** Verifica se A consegue empacotar Bytes, enviar para B e sair limpo do outro lado da Nuvem.
*   **02_pentest_mitm:** O Sniffer Alienígena da pasta tenta ler pedaços literais da memória. Falha graças ao XOR HMAC Mutante.
*   **03_pentest_dos_cannon:** O Canhão dispara 500 conexões maliciosas por milissegundo no Omega. Testa a estabilidade do CPU.
*   **04_websocket_chat:** Testa o "Full-Duplex" - bytes não empacotados em HTTP, mas sim uma tubulação TCP pura aberta eternamente rodando chat real-time.

## 🐘 2. Simulando Englobamentos de Sistemas Clássicos
Estas pastas testam a compatibilidade do túnel ao acoplar linguagens de mercado:
*   **05_php_fpm_cgi:** FastCGI. Testa o envio de arrays `$_POST` massivos sendo reconstruídos corretamente antes do Apache.
*   **06_python_grpc:** Bufferização em Python. Streams binários HTTP/2 passando no CROM sem corromper pacotes.
*   **07_postgres_pgwire:** Testa um Backend DB cru engolido (`psql` login). Valida pacotes de Autenticação do PostgreSQL que não possuem Header HTTP.
*   **08_redis_resp:** Engole o Redis Cache. Tráfego hiper-rápido operando em milissegundos sem adicionar latência perceptível (Overhead de Crypto = Zero).
*   **09_iot_mqtt_broker:** Internet das coisas. Valida como dispositivos pequenos conseguem persistir ping-pong por trás do escudo.
*   **10_cplusplus_raw_tcp:** Uma porta C++ pura. Nada de HTTP. Testa se o CROM destrói strings nulas (`\0`) que programadores C++ usam como EOF.

## 🚨 3. Teste de Sobrecarga e Ataques Sistêmicos
*   **11_large_payload_chunking:** Envia um arquivo brutal de 50MB no terminal para forçar o GoLang Chunk size a quebrar e fatiar pacotes.
*   **12_high_concurrency:** Não é o canhão, mas sim milhares de clientes legítimos (Alphas validos com senha) usando a tubulação simultaneamente `Goroutines`.
*   **13_sybil_swarm_attack:** Um enxame de mil Scanners usando chaves erradas por um Cópia mal feita do software para ver se dá Memory Leak por log excessivo.
*   **14_silent_drop_validation:** O principal tesouro. Garante O(1) Time. Se eu chutar o servidor, o Firewall do GO derruba sem gastar CPU em resposta de `Erro HTTP 403`.
*   **15_split_brain_recovery:** Derruba o Servidor CROM da porta e religa. O Alpha do cliente precisa saber reestabelecer o TCP Handshake sozinho.

## 🧠 4. Parsers Mágicos & Ameaças Avançadas
*   **16_nodejs_express_rest:** Testa envios JSON massivos com Headers variados (onde a Compressão LLM nativa do Gen-3 prova valor esmagando os atributos longos atômicos).
*   **17_java_spring_boot_xml:** O terror do mundo das empresas. Envia SOAP/XML via CROM. O Parser deve conseguir transitar intacto.
*   **18_dns_hijack_spoofing:** Tentativa de passar Cabeçalhos de HOST falsificados dizendo estar em IP da intranet local.
*   **19_payload_forgery:** O Alien Sniffer intercepta o tráfego da rede, altera o último byte de uma senha Criptografada e tenta enviar pro Omega. O HMAC rejeita o pacote pela Checksum.
*   **20_vfs_fd_exhaust:** Esgotamento de `File Descriptors`. Força o Linux ao talo para tentar *Crashar* o serviço de borda.

## 🔮 5. A Geração 3 Absoluta (A Glória de Produção)
*   **21_private_brain_system:** Consolidou a tese do sistema privado, que nem sequer usa o framework clássico da empresa, fechando tráfego num loopback escuro invisível.
*   **22_onion_multi_hop_route:** Testa o repasse cebola cego. Um novo binário `proxy_onion_relay` repassa Sockets Alpha de um computador para o outro validando Dark Routing.
*   **23_jitter_cover_traffic:** O teste do Motor da Goroutine Anti-NSA. Garante que rajadas de tráfego de Lixo Hexadecimal (`JITT Magic Headers`) sejam identificadas na nuvem e Dropadas do canal principal sem interromper uma transação Web3 válida.

---
### ⚙️ Como essa Documentação Prova Tudo?
Se você subir no Servidor e ler o script central `master_audit.sh`, verá que o script acessa estas 23 pastas uma por uma em tempo real e roda o compilado validando sucesso. Nenhuma ponta solta. O motor P2P atua perfeitamente desde o C++ mais básico até Mixnets P2P Onion Complexas.
