import socket
import logging

logging.basicConfig(level=logging.INFO)

def start_redis_mock(port=6379):
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(('127.0.0.1', port))
    server.listen(5)
    logging.info(f"[REDIS-MOCK] Fake In-Memory Cache rodando na porta {port}")

    while True:
        client, _ = server.accept()
        data = client.recv(1024)
        if data:
            cmd = data.decode('utf-8', errors='ignore').strip()
            logging.info(f"Comando recebido: {repr(cmd)}")
            
            # Responder no padrao RESP simple string
            client.sendall(b"+PONG\r\n")
        client.close()

if __name__ == '__main__':
    start_redis_mock()
