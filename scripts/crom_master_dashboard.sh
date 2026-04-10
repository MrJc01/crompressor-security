#!/bin/bash
# =========================================================================
# CROM-SEC MASTER DASHBOARD (Central TUI de Comando)
# =========================================================================

# Cores Unix Universais
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Requer privilégios
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}[!] Atenção: O Dashboard TUI deve gerenciar Serviços do OS.${NC}"
    echo -e "${YELLOW}Inicie preferencialmente com: sudo bash $0${NC}"
fi

clear_screen() {
    clear
}

header() {
    echo -e "${BLUE}================================================================${NC}"
    echo -e "${CYAN}             📡  CROM-SEC: PAINEL MESTRE (TUI)  📡             ${NC}"
    echo -e "${BLUE}================================================================${NC}"
}

pause() {
    echo -e "\n${NC}Pressione [ENTER] para retornar ao Radar Mestre..."
    read -r
}

# ----------------- Funcionalidades Táticas -----------------

opt_radar_tcp() {
    clear_screen
    echo -e "${CYAN}>>> RADAR DE REDE TÁTICO (Cérebros Conectados) <<<${NC}"
    echo -e "Analisando descritores vivos (TCP) da engine CROM...\n"
    
    echo -e "${YELLOW}Socket | Estado | IP Local:Porta | IP Alienigena:Porta${NC}"
    ss -tunp | grep -E "crom_omega" || echo -e "${RED}[x] Nenhum tráfego vivo da Mixnet encontrado neste momento.${NC}"
    
    echo -e "\n${CYAN}>>> RESUMO DO CPU / RAM (Cérebros)<<<${NC}"
    ps aux | grep crom_omega | grep -v grep | awk '{print "PID: "$2 " | CPU: "$3"% | RAM: "$4"% | Path: "$11}' || echo "Nenhum Daemon Vivo."
    
    pause
}

opt_listar_cerebros() {
    clear_screen
    echo -e "${CYAN}>>> INVENTÁRIO DE CÉREBROS (Nós Engolidos) <<<${NC}"
    
    CEREBROS=$(ls /etc/systemd/system/crom-omega-*.service 2>/dev/null)
    
    if [ -z "$CEREBROS" ]; then
        echo -e "${RED}[x] Nenhum sistema legado foi engolido ainda.${NC}"
    else
        echo -e "${GREEN}Processadores CROM Detectados no Systemd:${NC}"
        for c in $CEREBROS; do
            NAME=$(basename $c)
            STATUS=$(systemctl is-active $NAME 2>/dev/null || echo "Desconhecido")
            PORTA=$(echo $NAME | grep -o -E '[0-9]+')
            
            if [ "$STATUS" == "active" ]; then
                ST_COLOR="${GREEN}VIVO${NC}"
            else
                ST_COLOR="${RED}MORTO${NC}"
            fi
            
            echo -e " 🧠 ${MAGENTA}$NAME${NC} | Alvo Legado: ${CYAN}Porta $PORTA${NC} | Status: [$ST_COLOR]"
            # Extraindo Hash Seed para debug:
            SEED_IN=$(grep -i "TENANT_SEED=" $c | cut -d "=" -f 3)
            echo -e "    Seed Cripto: $SEED_IN\n"
        done
    fi
    pause
}

opt_operacao_editar() {
    clear_screen
    echo -e "${YELLOW}>>> MESA CIRÚRGICA (Editar ou Parar Cérebro) <<<${NC}"
    CEREBROS=$(ls /etc/systemd/system/crom-omega-*.service 2>/dev/null)
    if [ -z "$CEREBROS" ]; then
        echo -e "${RED}Nada para editar.${NC}"
        pause
        return
    fi
    
    echo "Serviços disponíveis:"
    i=1
    declare -A MAPA
    for c in $CEREBROS; do
        NAME=$(basename $c)
        echo -e " [${GREEN}$i${NC}] $NAME"
        MAPA[$i]=$NAME
        i=$((i+1))
    done
    
    echo -e "\nEscolha o número do Cérebro para editar (ou ENTER para cancelar):"
    read -r CHOICE
    
    if [ -n "${MAPA[$CHOICE]}" ]; then
        ALVO="${MAPA[$CHOICE]}"
        echo -e "\nO que deseja fazer com ${CYAN}$ALVO${NC}?"
        echo " 1) Editar Env Vars (NANO) - Trokar Crypt Key/Portas"
        echo " 2) Reiniciar Serviço (Crash Recovery)"
        echo " 3) Parar Serviço (Drop All Connections)"
        read -r ACAO
        
        case $ACAO in
            1) 
                nano /etc/systemd/system/$ALVO
                echo -e "${YELLOW}[!] Aplicando Daemon-Reload no Kernel...${NC}"
                systemctl daemon-reload
                systemctl restart $ALVO
                echo -e "${GREEN}Mutação Genética Salva! Cérebro reiniciado.${NC}"
                ;;
            2) 
                systemctl restart $ALVO
                echo -e "${GREEN}Reiniciado com sucesso.${NC}"
                ;;
            3)
                systemctl stop $ALVO
                echo -e "${RED}Lobotomia concluída. Cérebro desligado.${NC}"
                ;;
        esac
    fi
    pause
}

opt_sonar_logs() {
    clear_screen
    echo -e "${MAGENTA}>>> SONAR FORENSE (Live Output) <<<${NC}"
    echo "Digite a porta do Cérebro que deseja escutar (ex: 8080):"
    read -r PORT
    if [ -n "$PORT" ]; then
        echo -e "${GREEN}[!] Modo Tail-Follow ativo (Aperte CTRL+C para sair do sonar)...${NC}"
        journalctl -fu "crom-omega-$PORT.service"
    fi
}

opt_deploy_novo() {
    clear_screen
    echo -e "${GREEN}>>> FORJA GENÉTICA (Novo Englobamento) <<<${NC}"
    
    SCRIPT_PATH="$(cd "$(dirname "$0")" && pwd)/deploy_omega_server.sh"
    
    if [ -x "$SCRIPT_PATH" ]; then
        bash "$SCRIPT_PATH"
    else
        echo -e "${RED}Erro: Não foi possível localizar ou executar $SCRIPT_PATH${NC}"
    fi
    pause
}

# ----------------- LOOP PRINCIPAL -----------------

while true; do
    clear_screen
    header
    echo -e " ${YELLOW}1.${NC} 📡 Radar de Transmissão (Ver Sockets & Métricas)"
    echo -e " ${YELLOW}2.${NC} 📋 Inventário de Cérebros (Listar Nodes Engolidos)"
    echo -e " ${YELLOW}3.${NC} 🛠️  Mesa de Cirurgia (Editar Configs / Restart)"
    echo -e " ${YELLOW}4.${NC} 🔬 Sonar Forense (Real-Time Live Logs)"
    echo ""
    echo -e " ${CYAN}5.${NC} 🔨 Adicionar Novo Sistema P2P (Executar Instalador DevOps)"
    echo -e " ${RED}0.${NC} Desplugar da Matrix (Sair)"
    echo -e "${BLUE}================================================================${NC}"
    echo -n "Comando Tático > "
    read -r OPTION

    case $OPTION in
        1) opt_radar_tcp ;;
        2) opt_listar_cerebros ;;
        3) opt_operacao_editar ;;
        4) opt_sonar_logs ;;
        5) opt_deploy_novo ;;
        0) clear_screen; echo -e "${CYAN}Conexão Terminada. Fique seguro.${NC}"; break ;;
        *) echo -e "${RED}Comando não reconhecido pela interface.${NC}"; sleep 1 ;;
    esac
done
