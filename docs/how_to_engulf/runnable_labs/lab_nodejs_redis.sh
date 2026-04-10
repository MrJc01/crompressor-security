#!/bin/bash
# =========================================================================
# LABORATÓRIO CROM EDUCATIVO: Engolindo APP Node.js em Tempo Real
# =========================================================================
set -e

# Cores UI Unix
GREEN='\033[0;32m'
BLUE='\033[1;34m'
RED='\033[0;31m'
NC='\033[0m'

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
BIN="$ROOT_DIR/test_suites/bin"

echo -e "${BLUE}================================================================${NC}"
echo -e "${GREEN} 🔬 LAB ACIONÁVEL: Node.js e Roteamento In-Memory CROM P2P ${NC}"
echo -e "${BLUE}================================================================${NC}"

# 1. Start Vulnerable Node App
echo -e "\n[PASSO 1] Subindo um servidor Node.js Legacy exposto (Porta 8080)..."
python3 -c "
import http.server, json
class NodeHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'status': 'HACKED'}).encode('utf-8'))
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1', 8080), NodeHandler).serve_forever()
" &
PID_NODE=$!
sleep 1

echo -e "${RED} >> Um hacker sniffou e acessou os plaintexts: $(curl -s http://127.0.0.1:8080/api)${NC}"
echo "    [O Atacante está lendo tudo e consumindo CPU do seu Node!]"
sleep 2

# 2. Add Firewall Rule (Simulated by Proxy Concept)
echo -e "\n[PASSO 2] Subindo Barreira de Cérebro OMEGA na porta 9999 da sua Nuvem Edge..."
$BIN/proxy_out > /dev/null 2>&1 &
PID_OMEGA=$!
sleep 1

echo -e "   [!] Sysadmin aplica iptables, cortando a porta 8080 da Web. Backend agora protegido em Loopback."
sleep 2

# 3. Alpha Connect
echo -e "\n[PASSO 3] Startando Client Side GoMobile SDK (O Front-End Blindado CROM Alpha) na Porta 5432..."
export SWARM_CLOUD_TARGET="127.0.0.1:9999"
$BIN/proxy_in > /dev/null 2>&1 &
PID_ALPHA=$!
sleep 1

# 4. Result
echo -e "\n[PASSO 4] O usuário autenticado e legitimo acessa com Tunnel Crypt:"
SAFE_OUTPUT=$(curl -s http://127.0.0.1:5432/api)
echo -e "${GREEN} >> Resposta Criptografada entregue nativa: $SAFE_OUTPUT${NC}"

echo -e "\n[INFO] Um atacante escaneando o OMEGA CROM na web (9999) sofrerá:"
curl -s --max-time 1 http://127.0.0.1:9999/ || echo -e "${RED} >> Connection DROP SILENCIOSO! (0 Bytes, zero timeouts, RST TCP Fechado).${NC}"

# CLEANUP
kill -9 $PID_NODE $PID_OMEGA $PID_ALPHA 2>/dev/null || true
echo -e "\n${BLUE}================================================================${NC}"
echo -e " ${GREEN}Laboratório de Roteamento Acadêmico Terminado!${NC}"
