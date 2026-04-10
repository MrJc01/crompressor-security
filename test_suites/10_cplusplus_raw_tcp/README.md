# 🛡️ Suíte de Batalha CROM: 10_cplusplus_raw_tcp

> **Status:** Parte Automática da Auditoria (As 23 Torres)
> **Módulo:** Pentesting P2P

## 🔬 O que este teste prova?
Uma porta C++ pura. Nada de HTTP. Testa se o CROM destrói strings nulas (`\0`) que programadores C++ usam como EOF.

---

### Execução e Log
Para executar este teste isoladamente, você pode rodar o executador dele na raiz:
```bash
./run_test.sh
```

*(Lembre-se de compilar o binário `proxy_out` e `proxy_in` para `test_suites/bin/` antes de disparar unitariamente caso modifique a engine na unha)*

### 🧭 O Mapa Completo
Esta torre é apenas uma das dezenas simuladas pela arquitetura. Para compreender como ela se integra com ataques gRPC, Mempool MEV, Onion, Jitter e Mitm:

🔙 **[VOLTAR AO GUIA DAS 23 TORRES (Visualizar Toda a Auditoria)](../GUIA_DAS_23_TORRES.md)**
