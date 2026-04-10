#!/bin/bash
# Suite 16: Node.js Express REST API Engulfer
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Run mock on port 8080 so proxy_out can route to it
python3 -c "
import http.server, json
class NodeHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'status': 'engolido_com_sucesso'}).encode('utf-8'))
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1', 8080), NodeHandler).serve_forever()
" &
PID_NODE=$!
sleep 0.3

$BIN/proxy_out &
PID2=$!
sleep 0.3
$BIN/proxy_in &
PID3=$!
sleep 0.5

RESP=$(curl -s --max-time 3 http://127.0.0.1:5432/api/test 2>/dev/null || echo "EMPTY")

if [[ "$RESP" == *"engolido"* ]] || [[ "$RESP" == *"alienigena"* ]]; then
    echo "16_nodejs_express_rest: PASS" > "$REPORTS/16_status.log"
    echo "PASS"
else
    echo "16_nodejs_express_rest: FAIL" > "$REPORTS/16_status.log"
    echo "FAIL"
fi

kill -9 $PID_NODE $PID2 $PID3 2>/dev/null
