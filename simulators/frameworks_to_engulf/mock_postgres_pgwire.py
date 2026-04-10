import socket
import logging
import sys

logging.basicConfig(level=logging.INFO)

# PGWire Protocol Simulator extremamente basilar
def start_postgres_mock(port=5432):
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(('127.0.0.1', port))
    server.listen(5)
    logging.info(f"[POSTGRES-MOCK] Fake SGDB rodando na porta {port}")

    while True:
        try:
            client, addr = server.accept()
            # PostgreSQL manda Handshake de Inicializacao Binaria
            data = client.recv(1024)
            if data:
                logging.info(f"PGWire Handshake Inbound: {data[:10]}")
                # String de AuthenticationOk Hex dummy (Caracteres hex nulos)
                client.sendall(b'R\x00\x00\x00\x08\x00\x00\x00\x00')
                client.close()
        except KeyboardInterrupt:
            sys.exit(0)
        except Exception as e:
            pass

if __name__ == '__main__':
    start_postgres_mock()
