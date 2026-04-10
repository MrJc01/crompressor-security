# Como Engolir: Mixnet em Node.js com React (SPA)

Frameworks modernos como **Node.js, Express, React, Vue, e Next.js** introduzem grande carga de roteadores web complexos, e pesados tráfegos REST via API ou Server-Side Rendering (SSR). Frequentemente, são acoplados a bancos em memórias altamente efêmeros como **Redis**.

O Crompressor engole ecossistemas assíncronos NodeJS de maneira impressionante graças a manipulação nativa IO `net.Conn` goroutine-unblocking do Go.

## A Arquitetura do SPA + Omega/Alpha

### Backend API (O Nodo Escuro)
Imagine a estrutura atual do Node: você usa PM2 para clusterizar `node api.js` na porta 3000. 

No Linux, trave tudo via IPTables e Firewall para apenas `localhost`:
```bash
# Bloquear a porta 3000 de acessos diretos.
sudo iptables -A INPUT -p tcp -s localhost --dport 3000 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3000 -j DROP
```

Logo à frente da barricada do firewall IP, nós colocamos o Cérebro Omega ouvindo conexões P2P encriptadas usando a mesma máquina via portas CloudFlare/Públicas (Pois o Proxy não sofre desse iptables).

### FrontEnd SPA (React / Vue) + Cérebro WASM (Alpha Engine)
Diferente de sistemas PHP (servidor clássico), Aplicações React são APENAS JavaScript rolando na máquina do celular/computador do visitante — efetuando requisições REST `fetch('api.minha-empresa.com/login')`.

**Para rodar a magia:**
O time React de Frontend simplesmente precisa adicionar nosso *SDK de Substituição* de requisição no arquivo `index.js/main.js` principal da sua aplicação.

```javascript
/* index.js do seu App React - Importar o Cérebro Alpha SDK Browser */
import { initCromAlphaEngine } from '@crompressor/alpha-wasm';

// Configurar o Cérebro Cliente antes de carregar o React
initCromAlphaEngine({
  tenantSeed: process.env.REACT_APP_TENANT_SEED, // Exemplo
  omegaEndpoint: 'https://p2p-cloud.minha-empresa.com',
  interceptAllRequests: true
});

import { createRoot } from 'react-dom/client';
import App from './App';
const root = createRoot(document.getElementById('root'));
root.render(<App />);
```

## Efeito Final:
Todas as vezes que os componentes React fizerem requisição a "Tabelas", "Gráficos de Dashboards", ou dispararem Websockets `socket.io()`, o Cérebro WASM interceptará localmente na memória do Google Chrome.
* Ninguém consegue ver o body real do pacote usando Inspecionar Elemento "Network".
* Nenhum Firewall intrusivo de LanHause detecta os endpoints.
* A carga do Motor Go interceptador é minúscula, gastando `~4mb` de RAM do visitante. 

---
**[🧭 Voltar ao Índice Principal](../../../INDICE.md)**
