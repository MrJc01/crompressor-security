import http.server
import json
import logging

logging.basicConfig(level=logging.INFO)

class NodeHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        logging.info("Recebi um mock GET Express!")
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'status': 'engolido_com_sucesso'}).encode('utf-8'))

if __name__ == '__main__':
    logging.info("[NODE-MOCK] Fake Express Server rodando em 8085")
    http.server.HTTPServer(('127.0.0.1', 8085), NodeHandler).serve_forever()
