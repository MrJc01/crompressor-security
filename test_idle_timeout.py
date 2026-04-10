import time
import socket

# backend
s_backend = socket.socket()
s_backend.bind(('127.0.0.1', 8080))
s_backend.listen()

print("Waiting for connection")

import threading
def run_backend():
    conn, _ = s_backend.accept()
    print("[Backend] Accepted")
    while True:
        try:
            d = conn.recv(1024)
            if not d: break
            print("[Backend] recv", d)
        except Exception as e:
            print("[Backend] Exception:", e)
            break
    print("[Backend] Close")
    conn.close()

threading.Thread(target=run_backend, daemon=True).start()

# Let proxy_in/out start, interact via proxy_in
s = socket.socket()
s.connect(('127.0.0.1', 5432))
s.sendall(b"Hello world")
print("[Client] Sent Hello")

time.sleep(12)
try:
    s.sendall(b"Hello again")
    print("[Client] Sent Hello again!")
    d = s.recv(1024)
    print("Recv", d)
except Exception as e:
    print("[Client] EXCEPTION:", e)
