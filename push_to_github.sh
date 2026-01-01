#!/bin/bash

echo "🚀 Script di caricamento su GitHub per ImgCompress-go"
echo "----------------------------------------------------"

# 1. Verifica Git
if ! command -v git &> /dev/null; then
    echo "❌ Git non è installato. Installalo prima di proseguire."
    exit 1
fi

# 2. Inizializza Repo se necessario
if [ ! -d ".git" ]; then
    echo "📦 Inizializzazione repository Git..."
    git init
    git branch -M main
else
    echo "ℹ️  Repository Git già inizializzato."
fi

# 3. Aggiungi file
echo "➕ Aggiunta file all'area di stage..."
git add .

# 4. Stato
echo "📄 Stato attuale:"
git status -s

# 5. Commit
echo ""
read -p "📝 Inserisci il messaggio di commit (Default: 'Release v1.0.0'): " COMMIT_MSG
COMMIT_MSG=${COMMIT_MSG:-"Release v1.0.0"}

git commit -m "$COMMIT_MSG"

# 6. Configurazione Remoto
REMOTE_URL=$(git remote get-url origin 2>/dev/null)

if [ -z "$REMOTE_URL" ]; then
    echo ""
    echo "🌍 Non è configurato un repository remoto."
    echo "   Vai su https://github.com/new e crea un nuovo repository VUOTO."
    read -p "🔗 Incolla qui l'URL del repository (es. https://github.com/utente/repo.git): " NEW_REMOTE
    
    if [ -n "$NEW_REMOTE" ]; then
        git remote add origin "$NEW_REMOTE"
        echo "✅ Remoto aggiunto."
    else
        echo "❌ Nessun URL fornito. Impossibile fare push."
        exit 1
    fi
else
    echo "🌍 Remoto configurato: $REMOTE_URL"
fi

# 7. Push
echo ""
echo "⬆️  Caricamento su GitHub (push)..."
git push -u origin main

echo ""
echo "🎉 Fatto! Il tuo codice è online."
