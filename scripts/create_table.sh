#!/bin/sh
# Cria o banco SQLite e a tabela cotacao usados pelo server.
# Uso: ./scripts/create_table.sh [caminho-do-banco]
# Sem argumento, o banco é criado em cmd/server/cotacao.db.
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB_PATH="${1:-$ROOT/cmd/server/cotacao.db}"

if ! command -v sqlite3 >/dev/null 2>&1; then
	echo "sqlite3 não encontrado. Instale o SQLite CLI e tente de novo." >&2
	exit 1
fi

sqlite3 "$DB_PATH" <<'SQL'
CREATE TABLE IF NOT EXISTS cotacao (
	code        TEXT,
	codein      TEXT,
	name        TEXT,
	high        TEXT,
	low         TEXT,
	varBid      TEXT,
	pctChange   TEXT,
	bid         TEXT,
	ask         TEXT,
	timestamp   TEXT,
	create_date TEXT
);
SQL

echo "Tabela cotacao pronta em $DB_PATH"
