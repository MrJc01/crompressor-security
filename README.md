# Crompressor Security Lab (CROM-SEC) 👽🛡️

Este repositório é o laboratório de cibersegurança isolado (Red Team / SRE) dedicado a testar e expandir os limites da arquitetura P2P e Compressão Semântica do **Crompressor** oficial.

O objetivo primário deste braço é provar a soberania e a inquebrabilidade do protocolo através de uma adaptação chamada **Alien Drop-In Proxy L4/L7**.

## A Tese do "Proxy Alienígena"

Ao invés de obrigar desenvolvedores em PHP, C++, Node ou Postgres a reescreverem suas pipelines de Sockets e APIs para utilizar o SDK de criptografia segura, a infraestrutura deste repositório sequestra qualquer conexão TCP crua, passando-a nativamente pelas engrenagens do CROM-WASM.

O trajeto original `[App] -> [API TCP]` é convertido de forma transparente para:
`[App Local] -> [Cérebro Proxy Ingress] ===(Malha Alienígna P2P)=== [Cérebro Proxy Egress] -> [App Destino]`

**Por que "Alienígena"?** Porque diferentemente do IPSEC puro ou Transport Layer Security clássico onde as transações evidenciam tamanhos explícitos e deixam traços da camada 7, a mutação dos *Codebooks* empurra para a rede blocos esquisitos do espaço Euclidiano. Sem ter acesso à Semântica Local e à Seed Mutante do cliente, um Hacker injetado no meio do Swarm não coleta NADA. Literalmente zero ruído contextualizável.

## Como as pastas estão organizadas:

*   **`/docs/`**: A biblioteca de estudos acadêmicos e teóricos detalhados explicando Drop-Ins, Volumes Docker em FUSE e o Isolamento Horizontal de Inquilinos via Hashes.
*   **`/simulators/`**: O centro nervoso real em Golang. Mocks táticos do Crompressor SDK que escrevem Sockets direto em `os.TempDir` e usam as pipelines do projeto real.
*   **`/simulators/pentest/`**: Nossas armas de assédio militar contra nosso próprio software. (Sniffers MITM, DoS TCP Cannons injetando Magic Bytes forjados para simular Poisoning).
*   **`/test_suites/`**: O sistema CI/CD agnóstico do laboratório de QA corporativo. Executa os testes automatizados matando os daemons em background no final e fornecendo laudos auditáveis.

## Executando as Suítes de Teste Automatizadas

Esta infra compila os bins no padrão *Ahead-of-Time* (AOT) cortando 99% da CPU inútil do Golang JIT num laboratório e atira sem piedade em todos os fluxos.

Para auditar o motor agressivamente, vá no diretório Master de testes:
```bash
chmod +x test_suites/master_audit.sh
./test_suites/master_audit.sh
```
O console exibirá os testes aprovados baseados nos carimbos de falha contra o canhão e contra os espiões, emitindo os resultados no log unificado na sua tela e dentro do core `/reports/`.
