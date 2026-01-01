#!/bin/bash

APP_NAME="ImgCompress"
ICON_SOURCE="icon.png"

echo "🪟 Inizio Build per Windows..."

# Crea cartella output se non esiste
mkdir -p dist

# 0. Preparazione Icona (Richiede go-winres)
if [ -f "$ICON_SOURCE" ]; then
    echo "   🎨 Preparazione icona per Windows..."
    
    # Verifica/Installa go-winres
    if ! command -v go-winres &> /dev/null; then
        echo "      ⬇️  Installazione tool 'go-winres'..."
        go install github.com/tc-hib/go-winres@latest
        # Aggiungi GOPATH/bin al PATH se non c'è
        export PATH=$PATH:$(go env GOPATH)/bin
    fi

    # Ridimensiona icona per Windows (Max 256x256)
    ICON_WIN="icon_win_temp.png"
    sips -z 256 256 "$ICON_SOURCE" --out "$ICON_WIN" > /dev/null

    # Crea configurazione temporanea per go-winres
    cat > winres.json <<EOF
{
  "RT_GROUP_ICON": {
    "APP_ICON": {
      "0": ["$ICON_WIN"]
    }
  }
}
EOF

    # Genera i file risorsa (.syso)
    # Questo creerà rsrc_windows_amd64.syso e rsrc_windows_arm64.syso
    go-winres make --in winres.json --arch amd64,arm64
else
    echo "⚠️  Icona $ICON_SOURCE non trovata. Gli exe non avranno icona."
fi

# 1. Build per Windows x64 (Standard)
echo "   🔨 Compilazione Windows x64 (Intel/AMD)..."
GOOS=windows GOARCH=amd64 go build -o "dist/${APP_NAME}_Win_x64.exe" main.go
if [ $? -eq 0 ]; then
    echo "      ✔ Creato dist/${APP_NAME}_Win_x64.exe"
else
    echo "      ❌ Errore compilazione x64"
fi

# 2. Build per Windows ARM64
echo "   🔨 Compilazione Windows ARM64..."
GOOS=windows GOARCH=arm64 go build -o "dist/${APP_NAME}_Win_ARM64.exe" main.go
if [ $? -eq 0 ]; then
    echo "      ✔ Creato dist/${APP_NAME}_Win_ARM64.exe"
else
    echo "      ❌ Errore compilazione ARM64"
fi

# 3. Pulizia file temporanei
rm -f winres.json rsrc_windows_*.syso "$ICON_WIN"

echo "✅ Build Windows completata! I file sono nella cartella 'dist'."
