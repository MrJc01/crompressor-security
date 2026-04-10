# Roteamento Multi-Camadas e Malha "Alienígena" (Alien Onion-Routing)

Um dos pilares conceituais do **Crompressor Alien Proxy** é garantir não apenas que o payload seja impenetrável (Cifrado e Comprimido Semânticamente), mas que o simples tráfego seja irrastreável quanto a padrão de rede temporal e comportamental (Traffic Analysis / Packet Sniffing via ISPs e Routers Nacionais).

## O Problema da Linha Reta Criptografada

Criptografia de túnel moderno clássica (como VPN TLS ou Wireguard IPSEC) sofre de metadados:
Se você envia para a DB na nuvem `150 bytes` (uma Query) e recebe `1.4 MB` (os dados de tabela) todo dia ao meio dia. Mesmo um IPS (Intrusion Prevention System) incapaz de ler o interior cifrado sabe que um _Job grande rodou para o Cliente X ali e demorou 1320ms_. Se o IP alvo e fonte são detectados, a estrutura comportamental está exposta.

## A "Topologia Alienígena": Multi-Cérebros (Mixmets)

Criamos uma rota onde o tráfego não vai direto de `A` (Cliente) para `C` (Backend Host).
Ele mergulha e pinga num ou dois Middle-Brains (Exemplo o nodo `B`) da nossa rede de enxame de `internal/network/swarm.go`.

### Como o Nó "B" Cega os Espiões:

1. Tráfego encapsulado (VFS Semântico) entra no `Cérebro Proxy In` A. É codificado, o payload encapsulado P2P (UDP Gossip/Bitswap).
2. "A" Manda a enxurrada de dados UDP aleatoriamente para "B" (Um Cérebro Relay que não quebra a cripto de fora, ou usa um sistema Onion-Cifrada de repasse em pacotes estáticos).
3. "B" (Relay Node) armazena pequenos blocos num pool temporário minúsculo (Jitter). Ele comuta e mescla blocos gerando um tráfego estático (MixNet/Cover Traffic - Enviando 200kbps constantes para a internet e misturando os payloads de clientes A e Z).
4. O nó "Relay" redireciona isso sob uma cadência "Alienígena" - totalmente amorfa se vista via PCAPs clássicos do WireShark para o nó `C`.
5. "C", lendo pacote a pacote a partir de assinaturas latentes e blocos Hash convergentes decifra os tokens e emite a chamada local real.

### Trade-off Calculado: Latências Elevadas para Infra Crítica
Tais desvios e re-empacotamentos causam milissegundos pesados (alta latência). Isso deve estar ciente nos casos de uso.
Recomendamos o Roteamento Multi-Camadas para tráfegos como:
- Relatórios Diários Seguros (Dados Fiscais).
- Transferência de logs hiper sensíveis Docker Host.
- Comandos SysAdmin Remotos via SSH Proxifier (Baixa banda, Zero Vazamento).

---
**[🧭 Voltar ao Índice Principal](../INDICE.md)**
