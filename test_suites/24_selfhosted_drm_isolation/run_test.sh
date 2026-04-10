#!/bin/bash
# ==============================================================================
# Suíte 24: Isolamento Anti-Pirata (Self-Hosted DRM) 
# Garante que um App local cego só aceita comandos validados pelo CROM
# ==============================================================================

set -e
ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="$ROOT_DIR/test_suites/bin"

if [ ! -f "$BIN/proxy_out" ]; then
    echo "Falta binario. Rode make ou script mestre."
    exit 1
fi

# Portas obrigatórias (Hardcoded nos binários Go compilados)
DRM_APP_PORT=8080
DRM_OMEGA_PORT=9999
DRM_ALPHA_PORT=5432

# Limpeza preventiva de processos antigos nas portas
fuser -k ${DRM_APP_PORT}/tcp 2>/dev/null || true
fuser -k ${DRM_OMEGA_PORT}/tcp 2>/dev/null || true
fuser -k ${DRM_ALPHA_PORT}/tcp 2>/dev/null || true
sleep 1

echo "============================================================"
echo " [SUITE 24] DRM SELF-HOSTED ISOLATION TEST"
echo "============================================================"

echo "[DRM] FASE 0: Subindo App Falso na porta restrita $DRM_APP_PORT..."
python3 -c "
import http.server
class SecureHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(b'{\"license\": \"VALID_CROM_ACCESS\", \"data\": \"SECRET_DRM_CONTENT\"}')
    def log_message(self, format, *args): pass
http.server.HTTPServer(('127.0.0.1', $DRM_APP_PORT), SecureHandler).serve_forever()
" &
PID_APP=$!
sleep 1

echo "[DRM] FASE 0: Subindo o Escudo Omega (Silent Drop Ativado)..."
$BIN/proxy_out > /dev/null 2>&1 &
PID_OMEGA=$!
sleep 1

# ============================================================
# FASE 1: O Hacker tentando bater direto no Omega sem Seed.
# ============================================================
echo "[DRM] FASE 1: Hacker tenta acessar API diretamente (DROP esperado)..."

HACKER_ATTEMPT=$(curl -s --max-time 2 http://127.0.0.1:$DRM_OMEGA_PORT/ 2>/dev/null || echo "SILENT_DROP")
if [ "$HACKER_ATTEMPT" = "SILENT_DROP" ] || [ -z "$HACKER_ATTEMPT" ]; then
    echo "[DRM] ✅ FASE 1 PASS: Silent Drop! Atacante recebeu 0 bytes."
else
    echo "❌ FASE 1 FAIL: O Omega vazou conteúdo: $HACKER_ATTEMPT"
    kill -9 $PID_APP $PID_OMEGA 2>/dev/null || true
    exit 1
fi

# ============================================================
# FASE 2: Cliente legítimo Alpha com Seed válida.
# ============================================================
echo "[DRM] FASE 2: Subindo Cliente Alpha com licença válida..."
$BIN/proxy_in > /dev/null 2>&1 &
PID_ALPHA=$!
sleep 2

VALID_CURL=$(curl -s --max-time 3 http://127.0.0.1:$DRM_ALPHA_PORT/ 2>/dev/null)
if echo "$VALID_CURL" | grep -q "SECRET_DRM_CONTENT"; then
    echo "[DRM] ✅ FASE 2 PASS: Titular leu API protegida com sucesso."
    echo ""
    echo "============================================================"
    echo " RESULTADO FINAL: PASS"
    echo "============================================================"
else
    echo "❌ FASE 2 FAIL: Alpha válido não atravessou."
    echo "  Response: [$VALID_CURL]"
fi

kill -9 $PID_APP $PID_OMEGA $PID_ALPHA 2>/dev/null || true
exit 0
