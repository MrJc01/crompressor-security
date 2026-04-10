#!/bin/bash
# Suite 09: MQTT IoT Telemetry
BIN="$(cd "$(dirname "$0")/../../test_suites/bin" && pwd)"
REPORTS="$(cd "$(dirname "$0")/../reports" && pwd)"

# Mock MQTT Broker: aceitar CONNECT e devolver CONNACK
python3 -c "
import socket
s=socket.socket()
s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('127.0.0.1',8080))
s.listen(1)
while True:
    c,_=s.accept()
    data=c.recv(256)
    # MQTT CONNACK fixo
    c.sendall(b'\x20\x02\x00\x00')
    c.close()
" &
PID_MQTT=$!
sleep 0.3

$BIN/proxy_out &
PID2=$!
sleep 0.3
$BIN/proxy_in &
PID3=$!
sleep 0.5

# MQTT CONNECT packet (hardcoded minimal)
RESP=$(printf '\x10\x0d\x00\x04MQTT\x04\x02\x00\x3c\x00\x01A' | nc -w 2 127.0.0.1 5432 2>/dev/null | wc -c)

if [ "$RESP" -gt "2" ] 2>/dev/null; then
    echo "09_iot_mqtt_broker: PASS (CONNACK recebido via tunnel)" > "$REPORTS/09_status.log"
    echo "PASS"
else
    echo "09_iot_mqtt_broker: FAIL" > "$REPORTS/09_status.log"
    echo "FAIL"
fi

kill $PID_MQTT $PID2 $PID3 2>/dev/null
