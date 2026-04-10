#!/bin/bash
# Suite 05: PHP FastCGI interception
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock PHP: python http server como se fosse php -S
python3 -c "
import http.server, json
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type','text/html')
        self.end_headers()
        self.wfile.write(b'<?php echo CROM_PHP_ENGULFED; ?>')
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1',8080),H).serve_forever()
" &
PID_PHP=$!
sleep 0.3

echo "$CROM_TENANT_SEED" | $BIN/proxy_out &
PID2=$!
sleep 0.3
echo "$CROM_TENANT_SEED" | $BIN/proxy_in &
PID3=$!
sleep 0.5

RESP=$(curl -s --max-time 3 http://127.0.0.1:5432/ 2>/dev/null || echo "EMPTY")

if [[ "$RESP" == *"CROM_PHP"* ]] || [[ "$RESP" == *"alienigena"* ]]; then
    echo "05_php_fpm_cgi: PASS (PHP engolido com sucesso)" > "$REPORTS/05_status.log"
    echo "PASS"
else
    echo "05_php_fpm_cgi: FAIL" > "$REPORTS/05_status.log"
    echo "FAIL"
fi

kill $PID_PHP $PID2 $PID3 2>/dev/null
