#!/bin/sh
# Build kuferek and install it into ~/.local/bin.
set -e

cd "$(dirname "$0")"

BIN_DIR="$HOME/.local/bin"

echo "Building kuferek..."
go build -o kuferek .

mkdir -p "$BIN_DIR"
install -m 0755 kuferek "$BIN_DIR/kuferek"

echo "Installed: $BIN_DIR/kuferek"
case ":$PATH:" in
	*":$BIN_DIR:"*) ;;
	*) echo "Note: $BIN_DIR is not in your PATH." ;;
esac
