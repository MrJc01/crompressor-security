# Sistema Privado Local (Private Brain System)

## O Conceito

Imagine rodar um site PHP completo com banco de dados SQL **dentro** do Cérebro Crompressor. 
Mesmo que o servidor esteja na sua máquina local, **você não consegue acessar os dados crus**.

O Cérebro age como uma "caixa-forte computacional":
- O PHP e o SQL rodam **escondidos** na porta `127.0.0.1:8080` (acessível SOMENTE localmente)
- O Cérebro Omega (`proxy_out`) é o **único** que fala com o PHP/SQL
- O Cérebro Alpha (`proxy_in`) é a **única porta de entrada** autorizada
- Qualquer acesso direto à porta P2P (9999) sem a Seed do tenant é **silenciosamente destruído**

## Vetores Testados

1. **Acesso Autorizado (via Cérebro Alpha):** O usuário legítimo envia requests via porta 5432. O Alpha encripta, o Omega descriptografa, o PHP responde, e a resposta volta encriptada. ✅
2. **Bypass Direto (localhost:8080):** Em produção, esta porta seria bloqueada por `iptables`. O teste demonstra porque o firewall é necessário.
3. **Hacker Sem Seed (porta P2P 9999):** Um atacante que encontra a porta P2P aberta e tenta injetar HTTP cru. O Silent Drop garante que ele recebe **zero bytes** de volta.

## Implicação Prática

Com `iptables` configurado para bloquear a porta 8080 externamente e permitir apenas `127.0.0.1`, o sistema se torna um **cofre digital funcional**. O site PHP opera normalmente para clientes autorizados via Cérebro Alpha, mas ninguém — nem mesmo o administrador do servidor — consegue ler os dados do banco sem possuir a Seed criptográfica correta.

---
**[🧭 Voltar ao Índice Principal](../../INDICE.md)**
