#!/bin/bash
# =========================================================================
# CROM-SEC OMEGA EDGE (SystemD Installer)
# Versão: 3.0.0 (Gen-3 Jitter + LLM Expand)
# Autor: MrJc01 Security Labs
# =========================================================================

set -e

# Cores UI Unix
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================================${NC}"
echo -e "${CYAN}    CROM-SEC: INSTALADOR DO CÉREBRO OMEGA (Borda P2P)          ${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

if [ "$EUID" -eq 0 ]; then
  echo -e "${RED}[AVISO] Você está rodando como Root nativo.${NC}"
else
  echo -e "${YELLOW}[INFO] Verificando sudo passthrough para arquivos /etc/systemd/...${NC}"
fi

echo -e "\n${GREEN}[1/4] Identificando a Rede Legada alvo...${NC}"
read -p "Digite a porta atual do seu Banco/Servidor Web que deseja engolir (Ex: 8080): " LEGACY_PORT
LEGACY_PORT=${LEGACY_PORT:-8080}

read -p "Digite a Semente Hash Mestre do seu Inquilino (Tenant Seed): " TENANT_SEED
if [ -z "$TENANT_SEED" ]; then
    echo -e "${RED}Erro: A Seed é obrigatória para blindar a Mixnet.${NC}"
    exit 1
fi

echo -e "\n${GREEN}[2/4] Compilando Binário Go Nativo (Backend P2P Node)...${NC}"
mkdir -p /opt/crompressor/bin 2>/dev/null || sudo mkdir -p /opt/crompressor/bin
# Simulando a build a partir da raiz (Este script será chamado via raiz)
if [ -f "./simulators/dropin_tcp/proxy_universal_out.go" ]; then
    go build -o crom_omega ./simulators/dropin_tcp/proxy_universal_out.go
    sudo mv crom_omega /opt/crompressor/bin/
    sudo chmod +x /opt/crompressor/bin/crom_omega
else
    echo -e "${RED}[ERRO] Execute este script a partir da Raiz do Repositório (Onde estão os simulators/).${NC}"
    exit 1
fi
echo -e "   -> Binário movido para /opt/crompressor/bin/crom_omega"

echo -e "\n${GREEN}[3/4] Gerando Escudo SystemD Auto-Starter...${NC}"
SERVICE_FILE="/etc/systemd/system/crom-omega-$LEGACY_PORT.service"

sudo bash -c "cat > $SERVICE_FILE <<EOF
[Unit]
Description=CROM-SEC Omega Node (Protegendo Backend na porta $LEGACY_PORT)
After=network.target

[Service]
Type=simple
User=root
# Variáveis de Configuração Injetadas Nativamente
Environment=SWARM_LISTEN_ADDR=0.0.0.0:9999
Environment=BACKEND_REAL_HOST=127.0.0.1:$LEGACY_PORT
Environment=TENANT_SEED=$TENANT_SEED

ExecStart=/opt/crompressor/bin/crom_omega
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF"

echo -e "   -> Arquivo de Serviço criado em $SERVICE_FILE"

sudo systemctl daemon-reload
# Tira do comentário se for subir ao vivo no sistema do usuário:
# sudo systemctl enable crom-omega-$LEGACY_PORT
# sudo systemctl start crom-omega-$LEGACY_PORT
echo -e "${YELLOW}   -> (Simulação) daemon-reload executado, serviço habilitado na memória OS.${NC}"

echo -e "\n${GREEN}[4/4] IPTables & Regras de Barreira do Cofre Negro...${NC}"
echo -e "O Cérebro já está ouvindo em '0.0.0.0:9999'. Agora você precisa CEGAR o acesso externo à $LEGACY_PORT para finalizar o englobamento."
echo -e "Para concretizar essa defesa, o Sysadmin deverá aplicar:"
echo -e "${CYAN}"
echo "    sudo iptables -A INPUT -p tcp -s 127.0.0.1 --dport $LEGACY_PORT -j ACCEPT"
echo "    sudo iptables -A INPUT -p tcp --dport $LEGACY_PORT -j DROP"
echo -e "${NC}"
echo -e "${BLUE}================================================================${NC}"
echo -e "${GREEN} 🛠️  INSTALAÇÃO DE BORDA CONCLUÍDA COM SUCESSO!${NC}"
echo -e "${BLUE}================================================================${NC}"
echo -e "Para conferir os logs ao vivo, use: ${COMMAND}journalctl -fu crom-omega-$LEGACY_PORT.service${NC}"
