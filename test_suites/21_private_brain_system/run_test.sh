#!/bin/bash
# Suite 21: Private Brain System
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"
# Site PHP com dados sensíveis (porta 8080 - PRIVADA)
export CROM_TENANT_SEED="CROM-SEC-TENANT-ALPHA-2026"
python3 -c "
import http.server, json
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type','application/json')
        self.end_headers()
        self.wfile.write(json.dumps({
            'db_users': [{'name':'admin','password':'super_secret_2026'}],
            'api_key': 'sk-crom-PRIVATE-0xDEADBEEF'
        }).encode())
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1',8080),H).serve_forever()
" &
PID_PRIV=$!
sleep 0.3

echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID_OUT=$!
sleep 0.3
echo "$CROM_TENANT_SEED" | $BIN/proxy_in &
PID_IN=$!
sleep 0.5

# TESTE A: Acesso AUTORIZADO (via Cérebro Alpha porta 5432)
RESP_AUTH=$(curl -s --max-time 3 http://127.0.0.1:5432/ 2>/dev/null || echo "")

# TESTE B: Hacker sem seed (porta P2P 9999 direta)
RESP_HACK=$(echo "GET / HTTP/1.1" | nc -w 2 127.0.0.1 9999 2>/dev/null || echo "")

PASS=true

if [[ -z "$RESP_AUTH" ]]; then
    PASS=false
fi

# Se o hacker conseguiu dados sensíveis sem seed = FAIL
if [[ -n "$RESP_HACK" ]] && echo "$RESP_HACK" | grep -qiE "secret|password|admin|api_key"; then
    PASS=false
fi

if [ "$PASS" = true ]; then
    echo "21_private_brain_system: PASS (acesso autorizado OK, hacker bloqueado)" > "$REPORTS/21_status.log"
    echo "PASS"
else
    echo "21_private_brain_system: FAIL" > "$REPORTS/21_status.log"
    echo "FAIL"
fi

kill $PID_PRIV $PID_OUT $PID_IN 2>/dev/null
