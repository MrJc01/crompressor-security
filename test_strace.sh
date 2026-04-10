echo "CROM-SEC-TENANT-ALPHA-2026" | ./test_suites/bin/proxy_out > /dev/null 2>&1 &
PID=$!
sleep 1
strace -p $PID -e read -s 100 2>&1 | head -n 5 
kill -9 $PID
