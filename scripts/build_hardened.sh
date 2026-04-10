#!/bin/bash
# =============================================================================
# Script de build hardened (Gen-8)
# Remove símbolos de debug (-s -w), oculta caminhos absolutos (-trimpath),
# e valida que o binário foi produzido limpo.
# [GEN-8 RT-201/208/209/210 FIX]
# =============================================================================

set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN_DIR="$ROOT"
BIN_TEST="$ROOT/test_suites/bin"
mkdir -p "$BIN_TEST"

echo "================================================================="
echo " [BUILD] CROM-SEC Gen-8 Hardened Build"
echo "================================================================="

# Compilar Omega (proxy_out)
echo "[BUILD] Compilando Omega (proxy_out) com Hardening Gen-8..."
go build -ldflags="-s -w" -trimpath -o "$BIN_DIR/proxy_universal_out" "$ROOT/simulators/dropin_tcp/proxy_universal_out.go"
echo "  → $(ls -lh "$BIN_DIR/proxy_universal_out" | awk '{print $5}')"

# Compilar Alpha (proxy_in)
echo "[BUILD] Compilando Alpha (proxy_in) com Hardening Gen-8..."
go build -ldflags="-s -w" -trimpath -o "$BIN_DIR/proxy_universal_in" "$ROOT/simulators/dropin_tcp/proxy_universal_in.go"
echo "  → $(ls -lh "$BIN_DIR/proxy_universal_in" | awk '{print $5}')"

# Copiar para test_suites/bin
cp "$BIN_DIR/proxy_universal_out" "$BIN_TEST/proxy_out"
cp "$BIN_DIR/proxy_universal_in" "$BIN_TEST/proxy_in"
echo "[BUILD] Binários copiados para test_suites/bin/"

# Validação
echo ""
echo "[VALIDATE] Verificando strip de símbolos..."
SYMS_OUT=$(go tool nm "$BIN_DIR/proxy_universal_out" 2>/dev/null | wc -l || echo "0")
SYMS_IN=$(go tool nm "$BIN_DIR/proxy_universal_in" 2>/dev/null | wc -l || echo "0")
echo "  proxy_out: $SYMS_OUT símbolos (deve ser 0)"
echo "  proxy_in:  $SYMS_IN símbolos (deve ser 0)"

echo ""
echo "[VALIDATE] Verificando debug sections..."
DEBUG_OUT=$(readelf -S "$BIN_DIR/proxy_universal_out" 2>/dev/null | grep -c "debug" || echo "0")
DEBUG_IN=$(readelf -S "$BIN_DIR/proxy_universal_in" 2>/dev/null | grep -c "debug" || echo "0")
echo "  proxy_out: $DEBUG_OUT debug sections (deve ser 0)"
echo "  proxy_in:  $DEBUG_IN debug sections (deve ser 0)"

echo ""
echo "[VALIDATE] Verificando paths do developer..."
PATHS_OUT=$(strings "$BIN_DIR/proxy_universal_out" | grep -c "^/home" || echo "0")
PATHS_IN=$(strings "$BIN_DIR/proxy_universal_in" | grep -c "^/home" || echo "0")
echo "  proxy_out: $PATHS_OUT paths expostos (deve ser 0)"
echo "  proxy_in:  $PATHS_IN paths expostos (deve ser 0)"

echo ""
echo "[VALIDATE] Verificando label KDF..."
KDF_OUT=$(strings "$BIN_DIR/proxy_universal_out" | grep -c "CROM_AES_GCM_KEY" || echo "0")
KDF_IN=$(strings "$BIN_DIR/proxy_universal_in" | grep -c "CROM_AES_GCM_KEY" || echo "0")
echo "  proxy_out: $KDF_OUT ocorrências (deve ser 0)"
echo "  proxy_in:  $KDF_IN ocorrências (deve ser 0)"

echo ""
echo "================================================================="
echo " [BUILD] Concluído! Gen-8 Hardened."
echo "================================================================="
