# Simulador Educacional WebAssembly (WASM)

Este diretório contém a prova de conceito Web demonstrando a funcionalidade nativa de interceptação do CROM-SEC no lado do cliente.

Em uma arquitetura de produção real do Crompressor, o motor em Go do "Cérebro Alpha" é compilado para binário `.wasm` utilizando as *flags* `GOOS=js GOARCH=wasm`. Esse motor reside na memória do navegador do usuário e sobrescreve nativamente as chamadas do objeto primário `window.fetch` e `XMLHttpRequest`.

### O Que Esse Laboratório Demonstra

Usamos um simulador JS/Frontend para provar a técnica de **Isolamento Silencioso no Browser**:

1.  A Aplicação Web finge enviar tráfego normal. Ex: React acessa a `/api/secret` de seu próprio servidor fictício.
2.  A camada injetada `cerebro_alpha_engine.js` detecta a requisição, **sequestra-a na memória local (RAM)** e criptografa usando a técnica LSH+HMAC (XOR Cyclic na Seed do inquilino).
3.  O túnel emite o lixo ininteligível (*payload sujo*) na rede física usando um Edge/Relay P2P Socket (O nosso mock na porta `9999`).
4.  Qualquer hacker ou proxy de rede (Zscaler, Firewall da Empresa, Wi-Fi do Hotel) bisbilhotando a tab "Network" vai conseguir capturar apenas pacotes sem sentido, opacos a analises estáticas.
5.  A resposta é devolvida de forma similar e processada inteiramente no Cérebro Injetado.

### Como Executar a Demonstração Funcional?

*   Abra o Terminal e rode na máquina hospedeira o nosso falso servidor de borda Edge/Nuvem: `python3 mock_websocket_omega.py`. Ele irá escutar na porta `9999` e será o responsável por processar o tráfico falso do browser.
*   Em seguida, dê um clique-duplo diretamente no arquivo `index.html` pelo seu navegador de preferência, ou através do VS Code usando "Live Server", e clique no botão na UI. Observe a consola.

---
**[🧭 Voltar ao Índice Principal](../../INDICE.md)**
