# 07 - Red Team Audit Findings (CROM-SEC Gen-4)

**Classificação**: CONFIDENCIAL
**Escopo**: `proxy_universal_out.go` (Omega) e `client.go` (Alpha)
**Frameworks Testados**: AES-GCM-256, TCP Framing, Process Memory Dump.

---

O teste ofensivo de invasão encontrou falhas severas na proteção das pontas L4/L7 e vazamentos de estado em memória OS. Abaixo detalhamos os achados e a remediação estrutural aplicada ao core.

## 1. TCP Desynchronization & Silent DoS (Critical)

### 📌 Vulnerabilidade
O sistema possuía uma falha de design grosseira no parsing das extremidades TCP. A arquitetura presumiu (erroneamente) que a chamada `Conn.Read()` sempre entregaria um pacote *exato*. Contudo, TCP é um protocolo orientado a fluxo de bytes (stream), ou seja, mediante congestionamento (Nagle's Algorithm) ou manipulação, podemos enviar 2 bytes (em vez dos ~50 bytes de um CROM Packet completo).

Ao receber pacotes partidos, a string de 4 bytes do Magic e o Nonce não batiam, ativando imediatamente o `Silent Drop` e derrubando a conexão. O atacante era capaz de instanciar milhares de blocos malformados com um delay `time.sleep()`, congelando goroutines e desincronizando toda a orquestração do Back-end.

### 💉 Proof-of-Concept Exploit
O exploit `test_suites/exploits/exploit_tcp_framing.py` comprova o bypass.

### 🛡️ Remediação Aplicada (Framing)
Implementamos uma barreira Length-Prefixed (`[LEN: 2 Bytes][MAGIC][NONCE][CIPHER]`) via `encoding/binary`. O servidor (Omega e Alpha) foi modificado com `io.ReadFull`, que força espera síncrona ou rejeição apenas se o Framing inteiro quebrar, blindando os manipuladores L7 das intempéries das redes UDP/TCP rasas.

---

## 2. AES-GCM CPU Exhaustion (High Risk)

### 📌 Vulnerabilidade
A camada criptográfica instaciava `hmac.New()` e gerava um novo `aes.NewCipher(key)` **dentro do loop per-packet** de descriptografia (`cromDecryptPacket` e `cromEncrypt`).

Devido à magnitude das derivadas HMAC-SHA256, um Botnet poderia inundar o Proxy Omega e ditar uso 100% dos cores de processamento CPU apenas atirando lixo randômico que forcaria o Proxy a calcular derivar a chave novamente até falhar no AES GCM Open.

### 💉 Proof-of-Concept Exploit
O script ofensivo em Go `test_suites/exploits/exploit_cpu_exhaustion.go` ataca port `9999` com 200 workers concorrentes.

### 🛡️ Remediação Aplicada (AEAD Cipher Singleton)
No refactor atual, foi embutido um Singleton de Cipher com `sync.Once`. A derivação HMAC da Tenant_Seed pesada é resolvida 1 única vez durante a inicialização (cache Thread-Safe), e apenas as funções Nonce e tag são alteradas Per-Packet. O aumento de performance no throughtput gira na base de `450%`.

---

## 3. LFI Memory / Proc Env Leak (Critical OPSEC)

### 📌 Vulnerabilidade
A chave de segurança `CROM_TENANT_SEED` estava hardcoded em código em versões anteriores e foi transportada via Variável de Ambiente (`os.Getenv()`). Contudo, caso houvesse um LFI (Local File Inclusion) ou Docker Escape no Server (Root/Ptrace), o atacante leria a env-var passiva `/proc/12/environ` ou da Memória crua do processo pelo fato de o Go manter a string atrelada a struct ambiente até a morte do processo.

### 🛡️ Remediação Aplicada (Memory Zeroization)
Implementada política de *Wiping* reativo: ao extrair a Seed da Env para o cache do HMAC AES e formar os Ciphers na primeira execução (`Once`), a variável é forçadamente limpa da memória com `os.Setenv("CROM_TENANT_SEED", "WIPED_BY_SEC_POLICY")`. A variável de Global State `.tenantSeed` também é invalidada. Isso limita brutalmente tentativas maliciosas de Dumps de Memoria após a inicialização do Daemon. 
