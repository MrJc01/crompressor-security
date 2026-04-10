import socket
import time

conns = []
for i in range(15):
    try:
        s = socket.socket()
        s.connect(('127.0.0.1', 5432))
        conns.append(s)
        print(f"Connected {i+1}")
    except Exception as e:
        print(f"Failed at {i+1}: {e}")

time.sleep(2)
for i, s in enumerate(conns):
    try:
        s.sendall(b"GET / HTTP/1.1\r\n\r\n")
        d = s.recv(1024)
        print(f"Conn {i+1} got response len: {len(d)}")
    except Exception as e:
        print(f"Conn {i+1} error: {e}")
