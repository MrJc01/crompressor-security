# 📖 MANUAL DE OPERAÇÃO MESTRE (CROM-SEC)
**Versão Gen-3 (Onion + Jitter + LLM)**

Este é o **Manual do Usuário Final** para SysAdmins, Desenvolvedores Web3 e DevOps. Aqui você aprenderá, na ordem exata e prática, como fechar uma porta web pública, erguer o Escudo CROM na sua empresa e monitorar os atacantes falhando em tempo real, utilizando apenas as ferramentas nativas que construímos no Terminal.

---

## Passo 1: O Englobamento Inicial (A Forja Genética)
Você tem um banco de dados **Redis**, um nó validador **Geth de Ethereum**, ou um servidor **PHP Nginx** rodando no seu Linux (Ex: Eles escutam na porta `8080`). Você quer torná-los invisíveis e acessíveis apenas via Criptografia P2P Orgânica CROM.

### Ação 1.1: Rodando a Instalação
No servidor onde seu App vulnerável se encontra, acesse a raiz deste repositório e chame o automador raiz:
```bash
sudo bash scripts/deploy_omega_server.sh
```
1. Ele pedirá a porta legada atual (Digite **`8080`**).
2. Ele pedirá sua Master Seed Criptográfica (Digite algo difícil, como **`CROM_ADMIN_PROD_999X`**).
3. O instalador gerará o binário Mestre em `/opt/crompressor/bin/` e habilitará no Sistema Operacional como Daemon.

### Ação 1.2: A Chave e a Fechadura (IPTables)
O Cérebro OMEGA do CROM já está ouvindo aberto focado em processar apenas quem possui a *Seed*. Mas você *precisa fechar a porta 8080* pro resto do mundo. Cole no seu terminal o output verde que o script te entregou:
```bash
sudo iptables -A INPUT -p tcp -s 127.0.0.1 --dport 8080 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP
```
**Vitória Inicial:** Se o seu Node.js estava sofrendo DDOS, ele acaba de ser silenciado. Todo o tráfego da internet agora ignora a porta 8080, respondendo como "Connection Refused". 

---

## Passo 2: Gerenciamento pelo Terminal (TUI)
Você não precisa saber os comandos do Linux de cabeça. O CROM possui um painel gráfico dentro do seu terminal Bash.

Chame o Painel TUI (Terminal User Interface):
```bash
sudo bash scripts/crom_master_dashboard.sh
```

### Usando o Painel CROM:
- **Opção 1 (Radar)**: Vai te mostrar a Topologia TCP viva da máquina enxergando os pacotes. Use se você sabe que a API Mobile está chamando o servidor e quer ver se a conexão completou o túnel e engavetou bytes.
- **Opção 2 (Inventário)**: Mostrará uma lista com `crom-omega-8080`, com a etiqueta verde **VIVO** e vazará qual *Seed* está amarrada nele. Útil caso um Sysadmin esqueça a senha CROM que colocou há meses atrás.
- **Opção 4 (Sonar Logs)**: Digite "8080" e seu terminal rolará ao vivo as tentativas de acesso. Se o hacker bater ali com um scanner nmap, você verá apenas o OMEGA aplicando a lei silenciosa do DROP (0 Bytes).

---

## Passo 3: O Acesso do Cliente Autorizado (Go Mobile)
A Nuvem da sua empresa sumiu da web e virou um nó restrito OMEGA. Como o seu Funcionário de TI em Casa ou o Aplicativo Mobile do Cliente final acessa seu Redis/Node.js?

Usando o **Cérebro Alpha**.

O núcleo do lado do funcionário foi externalizado para se embutir nativamente (`Gomobile SDK`), mas pode ser invocado no Linux do Computador do Funcionário também:

### A Execução do In-Flight
No PC do Funcionário (que possui a Seed Confidencial), ele deve rodar sua própria barreira apontando para sua Nuvem.
```bash
# Exportando a Nuvem CROM
export SWARM_CLOUD_TARGET="IP_DA_SUA_NUVEM_AWS:9999"  # Endereço OMEGA exposto pela Forja Genética

# Rodando o Emulador do Cérebro
./test_suites/bin/proxy_in
```
*(Observação: Certifique-se nas entranhas do Go que o `TenantSeed` no lado Alpha coincida via Build com a Seed do Servidor!)*

A tela do Alpha ficará verde sinalizando `escutando na porta local 5432`.

### Passo 3.1: Trabalhando como se estivesse fisicamente na Empresa
Agora o Funcionario abre o *DBeaver* (para Banco de Dados) ou faz as chamadas Web3 no código dele local e simplesmente ignora a existência de IPs na internet. Ele digita `localhost:5432`.

A mágica criptográfica começa:
1. Ele faz uma Query no DB falso na porta 5432.
2. O sistema orgânico **GoMobile SDK** tokeniza sintaticas pesadas (Compressão LLM), engole, empacota num XOR Seed-driven.
3. Abre uma porta na Nuvem CROM que instantaneamente aplica a técnica **Neblina/Jitter** (Ocultando Analytics).
4. O servidor aceita, expande o dado e injeta no NodeJs Legado (aquele da porta 8080 Dropada que estavamos protegendo no Passo 1).

**Tudo isso acontece em milisegundos. Completamente transparente. Invadível e blindado contra ataques "Mempool / Front-Running / MEVs", sem ter precisado editar e compilar nem um único arquivo de rotas no código Legado.**
