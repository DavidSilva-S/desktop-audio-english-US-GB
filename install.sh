#!/bin/sh
# ---------------------------------------------------------------------------
# Instalador de <SEU_PROGRAMA>
#
# Uso:
#   curl -fsSL https://SEU-DOMINIO/install.sh | sh
#
# Este script:
#   1. Detecta SO e arquitetura
#   2. Baixa o binário correto do último release no GitHub
#   3. Instala em ~/.local/bin (ou /usr/local/bin se rodar como root)
#   4. Verifica se o diretório de instalação está no PATH
# ---------------------------------------------------------------------------

set -eu

# ------------------------- CONFIGURAÇÃO -------------------------
# Ajuste estes valores para o seu projeto:
GITHUB_REPO="SEU_USUARIO/SEU_REPO"   # ex: "david/uk-tts-player"
BINARY_NAME="seuprograma"            # nome final do executável instalado
# Nome do padrão de arquivo do release, sem extensão. Deve bater com o que
# o seu processo de build/release gera (ex: via goreleaser).
# Placeholders disponíveis: {version} {os} {arch} {ext}
ASSET_PATTERN="${BINARY_NAME}_{version}_{os}_{arch}{ext}"
# ------------------------------------------------------------------

BOLD="$(printf '\033[1m')"
GREEN="$(printf '\033[32m')"
RED="$(printf '\033[31m')"
YELLOW="$(printf '\033[33m')"
RESET="$(printf '\033[0m')"

info()  { printf "%s[info]%s %s\n" "$BOLD" "$RESET" "$1"; }
ok()    { printf "%s[ok]%s %s\n" "$GREEN" "$RESET" "$1"; }
warn()  { printf "%s[aviso]%s %s\n" "$YELLOW" "$RESET" "$1"; }
error() { printf "%s[erro]%s %s\n" "$RED" "$RESET" "$1" >&2; exit 1; }

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || error "Comando obrigatório '$1' não encontrado. Instale-o e tente novamente."
}

# ------------------------- PRÉ-CHECAGENS -------------------------
require_cmd curl
require_cmd uname
require_cmd tar

if command -v sha256sum >/dev/null 2>&1; then
    SHASUM_CMD="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
    SHASUM_CMD="shasum -a 256"
else
    warn "Nenhum utilitário sha256 encontrado — verificação de checksum será pulada."
    SHASUM_CMD=""
fi

# ------------------------- DETECÇÃO DE PLATAFORMA -------------------------
detect_os() {
    case "$(uname -s)" in
        Linux)  echo "linux" ;;
        Darwin) echo "darwin" ;;
        *)      error "Sistema operacional não suportado: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        armv7l)         echo "arm" ;;
        *)              error "Arquitetura não suportada: $(uname -m)" ;;
    esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
EXT="tar.gz"

info "Sistema detectado: ${OS}/${ARCH}"

# ------------------------- VERSÃO -------------------------
# Usa a versão passada como variável de ambiente (VERSION=v1.2.3 ... sh)
# ou busca a última release publicada no GitHub.
if [ "${VERSION:-}" = "" ]; then
    info "Buscando a última versão publicada em ${GITHUB_REPO}..."
    LATEST_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    VERSION="$(curl -fsSL "$LATEST_URL" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
    [ -n "$VERSION" ] || error "Não foi possível determinar a última versão. Verifique se o repositório tem releases publicados."
fi
info "Versão selecionada: ${VERSION}"

# ------------------------- MONTA URL DE DOWNLOAD -------------------------
ASSET_NAME=$(echo "$ASSET_PATTERN" \
    | sed "s/{version}/${VERSION}/g; s/{os}/${OS}/g; s/{arch}/${ARCH}/g; s/{ext}/.${EXT}/g")

DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ASSET_NAME}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

info "Baixando ${ASSET_NAME}..."
if ! curl -fsSL -o "${TMP_DIR}/${ASSET_NAME}" "$DOWNLOAD_URL"; then
    error "Falha ao baixar ${DOWNLOAD_URL}
Verifique se o release '${VERSION}' possui um artefato com esse nome para ${OS}/${ARCH}."
fi

# ------------------------- CHECKSUM (opcional) -------------------------
if [ -n "$SHASUM_CMD" ]; then
    CHECKSUMS_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/checksums.txt"
    if curl -fsSL -o "${TMP_DIR}/checksums.txt" "$CHECKSUMS_URL" 2>/dev/null; then
        info "Verificando checksum..."
        EXPECTED="$(grep "$ASSET_NAME" "${TMP_DIR}/checksums.txt" | awk '{print $1}')"
        ACTUAL="$($SHASUM_CMD "${TMP_DIR}/${ASSET_NAME}" | awk '{print $1}')"
        if [ -n "$EXPECTED" ] && [ "$EXPECTED" != "$ACTUAL" ]; then
            error "Checksum não confere! Esperado: $EXPECTED / Obtido: $ACTUAL"
        fi
        ok "Checksum verificado."
    else
        warn "Arquivo checksums.txt não encontrado no release — pulando verificação."
    fi
fi

# ------------------------- EXTRAÇÃO -------------------------
info "Extraindo..."
tar -xzf "${TMP_DIR}/${ASSET_NAME}" -C "$TMP_DIR"

BIN_PATH="${TMP_DIR}/${BINARY_NAME}"
[ -f "$BIN_PATH" ] || error "Binário '${BINARY_NAME}' não encontrado após extração. Confira ASSET_PATTERN no script."

chmod +x "$BIN_PATH"

# ------------------------- INSTALAÇÃO -------------------------
if [ "$(id -u)" = "0" ]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

info "Instalando em ${INSTALL_DIR}/${BINARY_NAME}..."
mv "$BIN_PATH" "${INSTALL_DIR}/${BINARY_NAME}"

ok "${BINARY_NAME} instalado com sucesso em ${INSTALL_DIR}/${BINARY_NAME}"

# ------------------------- CHECA PATH -------------------------
case ":$PATH:" in
    *":${INSTALL_DIR}:"*)
        ok "Pronto! Rode '${BINARY_NAME}' para começar."
        ;;
    *)
        warn "${INSTALL_DIR} não está no seu PATH."
        printf "Adicione a linha abaixo ao seu ~/.bashrc, ~/.zshrc ou equivalente:\n\n"
        printf "  export PATH=\"%s:\$PATH\"\n\n" "$INSTALL_DIR"
        ;;
esac
