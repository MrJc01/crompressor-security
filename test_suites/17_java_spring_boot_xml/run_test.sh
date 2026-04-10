#!/bin/bash
# Suite 17: Java Spring Boot SOAP/XML
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock SOAP XML server inline (porta 8080)
python3 -c "
import http.server
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200)
        self.send_header('Content-Type','application/xml')
        self.end_headers()
        self.wfile.write(b'<SOAP><Body><Result>JAVA_CROM_ENGULFED</Result></Body></SOAP>')
    def do_GET(self):
        self.do_POST()
    def log_message(self, *a): pass
http.server.HTTPServer(('127.0.0.1',8080),H).serve_forever()
" &
PID_JAVA=$!
sleep 0.3

$BIN/proxy_out &
PID2=$!
sleep 0.3
$BIN/proxy_in &
PID3=$!
sleep 0.5

RESP=$(curl -s --max-time 3 -X POST http://127.0.0.1:5432/soap -d '<request>test</request>' 2>/dev/null || echo "")

if [[ "$RESP" == *"JAVA_CROM"* ]] || [[ "$RESP" == *"alienigena"* ]]; then
    echo "17_java_spring_boot_xml: PASS (XML SOAP engolido)" > "$REPORTS/17_status.log"
    echo "PASS"
else
    echo "17_java_spring_boot_xml: FAIL" > "$REPORTS/17_status.log"
    echo "FAIL"
fi

kill $PID_JAVA $PID2 $PID3 2>/dev/null
