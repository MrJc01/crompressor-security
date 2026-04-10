#!/bin/bash
# Script de build hardened (Gen-6)
# Remove símbolos de debug (-s -w) e oculta caminhos absolutos (-trimpath)
# para prevenir Binary Intelligence Extraction (RT-01/RT-03)

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN_DIR="$ROOT"

echo "[BUILD] Compilando Omega (proxy_out) com Hardening Gen-6..."
go build -ldflags="-s -w" -trimpath -o "$BIN_DIR/proxy_universal_out" "$ROOT/simulators/dropin_tcp/proxy_universal_out.go"

echo "[BUILD] Compilando Alpha (proxy_in) com Hardening Gen-6..."
go build -ldflags="-s -w" -trimpath -o "$BIN_DIR/proxy_universal_in" "$ROOT/simulators/dropin_tcp/proxy_universal_in.go"

echo "[BUILD] Concluído! Verifique com 'file' se estão stripped."
