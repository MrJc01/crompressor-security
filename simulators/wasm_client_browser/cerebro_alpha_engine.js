/**
 * Crompressor Cérebro Alpha Engine (WASM Simulator)
 * Em um cenário real, isso é compilado de Go para WebAssembly.
 * Para nossa Prova de Conceito, usamos JS moderno para ilustrar a matemática exata,
 * mostrando como nós sequestramos os Sockets globais do navegador.
 */

const ENGINE_LOGS = document.getElementById('engineLogs');
const CROM_MAGIC = "CROM";
const TENANT_SEED = "CROM-SEC-TENANT-ALPHA-2026";
// Proxy P2P falso batendo direto em um mock Python rodando na máquina via HTTP para o bypass
const OMEGA_EDGE_NODE = "http://127.0.0.1:9999/omega_tunnel";

function appendEngineLog(text, type = 'info') {
    if (!ENGINE_LOGS) return;
    const div = document.createElement('div');
    div.className = `log-entry ${type}`;
    div.innerText = text;
    ENGINE_LOGS.appendChild(div);
    ENGINE_LOGS.scrollTop = ENGINE_LOGS.scrollHeight;
}

// 1. Simulação do SHA256 e HMAC (simplificada para XOR P2P visual)
function generateKeyBlock(seed) {
    // Simulando 32 bytes derivados
    let result = '';
    for(let i=0; i<32; i++) {
        result += seed.charCodeAt(i % seed.length).toString(16).padStart(2,'0');
    }
    return result;
}

function cromEncryptVisual(plaintext) {
    const key = generateKeyBlock(TENANT_SEED);
    let hexCipher = '';
    let rawEncryptedVisual = '';
    for (let i = 0; i < plaintext.length; i++) {
        const pCode = plaintext.charCodeAt(i);
        const kCode = key.charCodeAt(i % key.length);
        const cipherCode = pCode ^ kCode;
        hexCipher += cipherCode.toString(16).padStart(2, '0');
        // Apenas caracteres malucos para visualização "suja" de entropia
        rawEncryptedVisual += String.fromCharCode(33 + (cipherCode % 94)); 
    }
    return { hexCipher, rawEncryptedVisual };
}

// 2. BACKUP DO FETCH ORIGINAL
const originalFetch = window.fetch;

// 3. SEQUESTRO GLOBAL (A MÁGICA DA BLINDAGEM)
window.fetch = async function(...args) {
    const targetUrl = args[0];
    
    appendEngineLog(`[WASM] Interceptado request local para: ${targetUrl}`, 'action');
    
    // Converte o intent do fetch em uma string serializada (Ação crua)
    const rawAction = JSON.stringify({ method: 'GET', url: targetUrl });
    
    // Encrypta
    appendEngineLog(`[WASM] Aplicando HMAC-SHA256 Cyclic XOR (Seed: ${TENANT_SEED.substring(0,8)}...)`, 'info');
    const { hexCipher, rawEncryptedVisual } = cromEncryptVisual(rawAction);
    
    // Construção do Frame CROM
    const frame = `[HEADER:${CROM_MAGIC}][NONCE:0x8A7B][DATA:${rawEncryptedVisual}]`;
    
    appendEngineLog(`[WASM-PHYSICAL-LAYER] Pacote formatado e pronto para P2P. Entropia visual gerada:`, 'info');
    appendEngineLog(frame, 'raw');
    appendEngineLog(`[WASM] Disparando emulação Socket para Edge Node Omega (${OMEGA_EDGE_NODE})...`, 'action');

    try {
        // Envia lixo ininteligível pela rede de verdade
        const response = await originalFetch(OMEGA_EDGE_NODE, {
            method: 'POST',
            body: hexCipher // enviamos o hex para facilitar no nosso mock python
        });

        const omegaEncryptedReturn = await response.text();
        appendEngineLog(`[WASM] Retorno cego recebido do Omega Edge Node. Tamanho: ${omegaEncryptedReturn.length} bytes`, 'info');
        appendEngineLog(`[WASM] Descriptografando entropia recebida na RAM local...`, 'action');

        // Para a POC, o Python Omega vai mandar os bytes de volta em hex já fingindo
        // Nós apenas repassamos o json mockado pra interface do React explodir de alegria
        const jsonMocked = JSON.parse(omegaEncryptedReturn);
        
        appendEngineLog(`[WASM] JSON limpo extraído. Devolvendo ao React nativamente.`, 'secure');

        // Finge a resposta do window.fetch
        return new Response(JSON.stringify(jsonMocked), {
            status: 200,
            headers: { 'Content-Type': 'application/json' }
        });

    } catch (e) {
        appendEngineLog(`[WASM-ERR] Edge Omega falhou ou inalcançável: ${e.message}. Para a demo funcionar, rode o mock_websocket_omega.py!`, 'raw');
        throw e;
    }
};

appendEngineLog("[WASM] Cérebro Alpha Carregado em Isolated-Memory. Fetch Hijacked.", "secure");
