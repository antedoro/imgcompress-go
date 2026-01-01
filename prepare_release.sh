#!/bin/bash

echo "📦 Preparazione asset per la Release v1.0.1..."

# Crea cartella temporanea
mkdir -p release_assets
rm -f release_assets/*

# 1. Zippa le app macOS (fondamentale perché sono cartelle)
echo "   🍏 Zippando le app macOS..."
zip -r release_assets/ImgCompress_macOS_Silicon.zip ImgCompress_Silicon.app
zip -r release_assets/ImgCompress_macOS_Intel.zip ImgCompress_Intel.app

# 2. Copia gli eseguibili Windows
echo "   🪟 Copiando gli eseguibili Windows..."
cp dist/ImgCompress_Win_x64.exe release_assets/
cp dist/ImgCompress_Win_ARM64.exe release_assets/

echo "----------------------------------------------------"
echo "✅ Asset pronti nella cartella 'release_assets':"
ls -lh release_assets
echo "----------------------------------------------------"
echo "Prossimi passi:"
echo "1. Esegui ./push_to_github.sh per caricare il codice."
echo "2. Vai su https://github.com/antedoro/imgcompress-go/releases/new"
echo "3. Crea un nuovo tag 'v1.0.1'."
echo "4. Trascina i file dentro 'release_assets' nella sezione binary assets."
