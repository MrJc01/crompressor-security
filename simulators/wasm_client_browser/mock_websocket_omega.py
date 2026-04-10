import http.server
import json
import codecs

# O Cérebro Omega de Borda - O recebedor P2P em nuvem
class OmegaHandler(http.server.BaseHTTPRequestHandler):
    def do_OPTIONS(self):
        # CORS permissions for browser JS
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'POST, GET, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'X-Requested-With, Content-Type')
        self.end_headers()

    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        encrypted_hex_payload = self.rfile.read(content_length).decode('utf-8')
        
        print("\n==============================================")
        print("[OMEGA-EDGE] Pacote Criptografado Recebido do Browser (WASM)!")
        print(f"Payload Sujo: {encrypted_hex_payload[:30]}...")
        
        # Na vida real: Omega decifra o hex com o HMAC, lê "GET /api/secret",
        # repassa pro PHP em localhost. 
        # Aqui para a POC Web Visual, entregamos logo os dados de resposta do banco de dados fakes 
        # fingindo que foram "descriptografados e lidos pelo servidor php real"
        
        print("[OMEGA-EDGE] Descriptografando e pegando dados DB localhost:8080...")
        fake_db_response = {
            "usuario": "admin_crom",
            "financas": "R$ 4.532.110,00",
            "hash_secreto": "SHA256:d8a2...32fc",
            "mensagem": "Você está lendo um dado do Banco de Dados que fluiu intocável por P2P."
        }
        
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(fake_db_response).encode('utf-8'))
        print("[OMEGA-EDGE] DB Payload reenviado ao Cérebro Alpha (no cliente).")
        print("==============================================\n")

    def log_message(self, *a): pass

if __name__ == '__main__':
    print("[MOCK] Cérebro Omega Cloud Edge rodando na porta 9999...")
    print("Aguardando fetchs simulados encriptados do Browser (index.html)")
    http.server.HTTPServer(('127.0.0.1', 9999), OmegaHandler).serve_forever()
