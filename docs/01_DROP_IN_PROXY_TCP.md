# Visão Geral: Drop-in Proxy TCP com Crompressor

Este documento mapeia como o design flexível do Crompressor pode ser adaptado para se tornar um Proxy L4/L7 perfeitamente transparente. 
A missão é: **Engolir o tráfego gerado por sistemas já implantados, sem obrigar os desenvolvedores a modificar nenhuma linha de código da aplicação-fonte.**

## O Paradigma "Drop-In"

Um sistema Drop-In, na prática de conectividade moderna orientada ao "Zero Trust", é uma caixinha preta que encapsula ou sequestra uma conexão vulnerável ou pesada, otimiza ela matematicamente/criptograficamente e repassa para outra ponta idêntica (Proxy In/Proxy Out).

### Esquema de Conexão:

**Em vez de:**  
`[App Cliente Web/Mobile] ---------- (Internet Pública TLS) ---------> [Backend API / SQL Database]`

**Nossa Topologia CROM:**  
`[App Cliente] ---> [Crompressor_Proxy_Local (In)] ===== (MALHA ALIENÍGENA) =====> [Crompressor_Proxy_Server (Out)] ---> [Backend API]`

1. **Proxy In (Ingress):** Abre uma porta de escuta local, idêntica ao da base de dados verdadeira ex: `localhost:5432` (Postgres). 
   - A aplicação web conecta pensando estar falando via sockets TCP puros com a DB local.
   - O Proxy engole os bytes (HTTP Rest ou PostgreSQL Binary Protocol).
   - O Payload original sofre parsing pelo `Compiler Stream` nativo do Crompressor.  O contexto vira identificadores de espaço latente CROM. Esmagado e codificado via blocos Delta / XOR. Tudo envelopado num pacote UDP Gossip super seguro com Hash convergente (PostQuantum/AESGCM).
   
2. **Tunneling (A Malha):** A nuvem de Cérebros (P2P Gossip Network) envia e comuta esses pacotes esmagados através da web. Invasores veem apenas ruídos estatísticos em protocolos Swarm.

3. **Proxy Out (Egress):** Recebendo o estilhaço neural o Cérebro Servidor usa o seu próprio Codebook para desfazer o Delta XOR e reinjetar no loop localhost os Bytes exatos e bit a bit do Protocolo original do Postgres/Web na porta verdadeira da API, que ignora todo o percurso e devolve a resposta instantânea.

> [!TIP]
> A latência induzida pela compressão semântica (se feita com tensores nativos no CROM-WASM) tipicamente "paga" o seu tempo e velocidade, visto que mandaremos pacotes TCP muito menores pela internet pública, mitigando a saturação global e gargalos de Backhaul (link).
