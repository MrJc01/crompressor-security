# 🛡️ Diretiva de Defesa 06: Múltiplas Mitigações Anti-Root L7 e L4

O CROM-SEC é impermeável a ataques ingênuos L4, mas este guia traça como mitigamos ataques avançados de *Threat Modeling* na ponta (KPA, Dumps, Hooking).

### 1. Esgotamento de Key-Plaintext (KPA) no Tráfego P2P
**O Vector:** XOR Puro é frágil quando dados brutos de aplicação HTTP viajam nele.
**A Solução (Implantada):** `crypto/aes` no formato **GCM (Galois/Counter Mode)**.
Com AES-256 e o Nonce atrelado ao Header CROM, se o Payload da rede sofrer alteração ou spoofing sem a chave de *TenantSeed*, ele fracassa não apenas no Header Magic, mas na diretiva Criptográfica do Selo (Seal Checksum Authentication do GCM Block). Além do dado ilegível, garante *Integridade* irrefutável de Autoria L4.

### 2. TLS Root CA Evasão (Spoofing de Assinatura de Chaves)
**A Vulnerabilidade:** Se os Administradores de TI ou invasores instalam falsas Autoridades de Certificação Locais (CAs Root), eles interceptam a verificação de Licença do Alpha e copiam a Seed de Criptografia original via Man-in-the-Middle HTTPS em trânsito com Fiddler ou Burpsuite.
**A Mitigação:** Instalação Explicita de `Certificate Pinning`. No construtor de Ping *Auth* do GoLang Alpha, você chumba como constante o algorítmo de Chave e a impressão PÚBLICA SHA256 Exata do seu Servidor Matriz em Nuvem e valida no evento `DialTLS`. Mesmo a máquina Root tendo certificado válido mascarado, a API recusa conectar dizendo "Hash divergente, Keygen detectado".

### 3. Blindagem Pós-Compilatória Anti-Reversing
**A Vulnerabilidade:** GoLang não retira os nomes de funções e tabelas do executável sem parâmetros apropriados (como `print_json`). O invasor sobe no disassembler *Ghidra* ou *IDA* e lê os Offsets perfeitos achando onde capturar de volta a "Seed" ou desativar o "Silence Drop" apagando uma linha HEX do binário modificado (`jz` pra `jmp`).
**A Mitigação Tática (Garble + LDFlags):**
Instalação do *Obfuscador Go Garble*:
`go install mvdan.cc/garble@latest`

Comando de Build Militar de Despacho (Produção Real):
```bash
garble -tiny -literals build -ldflags="-s -w -buildid=" -trimpath -o proxy_alpha main_in.go
garble -tiny -literals build -ldflags="-s -w -buildid=" -trimpath -o proxy_omega main_out.go
```
O Resultado: Nada de referências da biblioteca CROM, sem variáveis de package global com nomes decifráveis e string lixo para dificultar desmonte em Sandbox e debuggers.

### 4. Namespaces Violados por Exec/nsenter
**A Vulnerabilidade:** O Invasor ignora o pacote Alpha inteiramente e tenta pular pro ambiente host (via root) entrando no Namespace de rede fechada do Docker via comando de `docker exec` (Shell Administrativo).
**A Mitigação:** Utilização do script atrelado a este repositório (`docker_anti_intrusion_daemon.sh`) como PID1, que tranca todas tentativas de Shell Subsequentes via Terminais em emulação (`/dev/pts/`) aniquilando a memória Lúdica/RamDisk caso seja ativado de forma indesejada.
