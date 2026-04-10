# Roteamento Nominal (Drop-In Mocker)

Esta suite de testes tem o foco em provar que o ambiente de roteamento CROM-SEC funciona exatamente como descrito. Esmagando conexões e remontando na outra ponta.

**O Fluxo Executado:**
`Curl -> Ingress Proxy (5432) -> Egress Proxy (9999) -> Dummy App (8080)`
