# Como Engolir: Servidores PHP Históricos e MySQL Legado

Muitos sistemas corporativos globais ainda rodam em velhas stacks "LAMP" (Linux, Apache/Nginx, MySQL, PHP). Essa stack é notória por estar sempre exposta na internet na porta 80 e 443, sendo o alvo primordial de *Scanners de Vulnerabilidade*, zero-days em plugins WordPress/Laravel, e tentativas de brute-force no MySQL porta 3306.

Neste guia, mostramos como injetar o sistema P2P CROM como uma blindagem impenetrável na frente de todo o servidor sem tocar nos seus arquivos PHP originais.

## A Arquitetura do Drop-in

### Antes (Vulnerável e Exposto)

*   **Ponto 1:** NGINX / Apache escutando em `0.0.0.0:80` fornecendo o PHP C-FPM. (Mundo todo acessa).
*   **Ponto 2:** Banco de Dados MySQL escutando em `0.0.0.0:3306`. (Aberto para Brute Forcers).

### Depois (Engolido)

*   **Ponto 1:** O Servidor NGINX é editado no Linux para ouvir `listen 127.0.0.1:8080;` apenas. (Bloqueia totalmente a Internet).
*   **Ponto 2:** O Banco de Dados é passado para o plano de fundo ou encapsulado pelo Backend diretamente localmente.
*   **Ponto 3:** Subimos o **Cérebro Omega CROM** na nuvem da empresa, que "escuta a internet" (porta 80) e encaminha o *plaintext* apenas para `127.0.0.1:8080` (onde o NGINX está seguro e oculto da internet real).

## Executando (Passo-a-Passo de Implementação SecOps)

### 1. Re-confinamento do NGINX
Edite seu `/etc/nginx/sites-available/default` (ou similar):
```nginx
server {
    # MUDE ACESSO ABERTO: listen 80;
    # PARA:
    listen 127.0.0.1:8080;

    root /var/www/html;
    index index.php index.html;
    server_name _;

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
    }
}
```
*Reinicie o NGINX.* A partir daqui sua empresa estaria tecnicamente "*offline*" e segura até contra varredura militar.

### 2. Implantando o Mestre Omega na Porta Extrovertida (Drop-In)
Na mesma máquina onde seu NGINX acabou de ser isolado do mundo, rode (ou transforme em SystemD service):

```bash
# Sobe o proxy_out na porta original da internet e reflete entropia na 8080
export SWARM_LISTEN="0.0.0.0:80"        # Porta exposta fisicamente ao router / internet
export BACKEND_TARGET="127.0.0.1:8080"  # O cofre privado local do Nginx
export TENANT_SEED="SECRET_DA_SUA_EMPRESA" 

./proxy_universal_out
```

### 3. Acesso Seguro Pelos Funcionários / Browser Alpha
Apenas clientes (usuários) que passarem fisicamente pela sua versão compilada do **browser WASM CROM**, ou que tenham instalado um Desktop Worker CROM ouvindo sua seed, poderão acessar sua Intranet PHP.
*   Cenário Desktop: Funcionário abre o `proxy_universal_in` e acessa os relatórios em `localhost:5432`.
*   Cenário Cliente (Mágica): O browser injetou `crompressor-client.wasm`. O site age e funciona magicamente, como demonstrado em nossa prova de conceito Web.

## Benefícios Imediatos à Empresa
1. **Fim de Banner Grabbing:** Um ataque rodando ferramentas não obterá NENHUMA resposta identificando "Servidor Apache/2.4 PHP". O Cérebro Omega fará **Silent Drop** em 1 milissegundo de qualquer malformação que não tenha os HMACs XOR.
2. **Defesa DDOS Intrínseca:** Inundar o proxy de bytes crus fará as goroutines matarem a conexão TCP internamente sem gastar processamento L7, resguardando o FPM Fila e Banco SQL local.

---
**[🧭 Voltar ao Índice Principal](../../../INDICE.md)**
