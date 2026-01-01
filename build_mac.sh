#!/bin/bash

APP_NAME_BASE="ImgCompress"
ICON_SOURCE="icon.png"
ICON_ICNS="AppIcon.icns"

echo "🍏 Inizio Build per macOS..."

# 0. Generazione Icona (una volta sola)
if [ -f "$ICON_SOURCE" ]; then
    echo "   🎨 Generazione icona condivisa .icns..."
    ICONSET="AppIcon.iconset"
    mkdir -p "$ICONSET"

    sips -z 16 16     "$ICON_SOURCE" --out "$ICONSET/icon_16x16.png" > /dev/null
    sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET/icon_16x16@2x.png" > /dev/null
    sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET/icon_32x32.png" > /dev/null
    sips -z 64 64     "$ICON_SOURCE" --out "$ICONSET/icon_32x32@2x.png" > /dev/null
    sips -z 128 128   "$ICON_SOURCE" --out "$ICONSET/icon_128x128.png" > /dev/null
    sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET/icon_128x128@2x.png" > /dev/null
    sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET/icon_256x256.png" > /dev/null
    sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET/icon_256x256@2x.png" > /dev/null
    sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET/icon_512x512.png" > /dev/null
    sips -z 1024 1024 "$ICON_SOURCE" --out "$ICONSET/icon_512x512@2x.png" > /dev/null

    iconutil -c icns "$ICONSET"
    rm -rf "$ICONSET"
else
    echo "⚠️  Icona non trovata. Uso default."
fi

# Funzione per creare il bundle
create_bundle() {
    ARCH=$1
    SUFFIX=$2
    APP_BUNDLE="${APP_NAME_BASE}_${SUFFIX}.app"
    BINARY_NAME="${APP_NAME_BASE}"

    echo "   🔨 Creazione $APP_BUNDLE (Arch: $ARCH)..."

    # Pulisci
    rm -rf "$APP_BUNDLE"

    # Struttura
    mkdir -p "$APP_BUNDLE/Contents/MacOS"
    mkdir -p "$APP_BUNDLE/Contents/Resources"

    # Compila il binario Go
    CGO_ENABLED=0 GOOS=darwin GOARCH=$ARCH go build -o "$APP_BUNDLE/Contents/MacOS/$BINARY_NAME" main.go
    if [ $? -ne 0 ]; then
        echo "❌ Errore compilazione per $ARCH"
        return
    fi

    # Copia Icona
    if [ -f "$ICON_ICNS" ]; then
        cp "$ICON_ICNS" "$APP_BUNDLE/Contents/Resources/"
    fi

    # Crea il sorgente AppleScript per il Launcher
    # Riceve il percorso del binario come primo argomento (argv 1)
    # E i file trascinati come argomenti successivi
    cat > "$APP_BUNDLE/Contents/Resources/launcher.applescript" <<EOF
on run argv
    if (count of argv) > 0 then
        set binaryPath to item 1 of argv
        set argList to ""
        if (count of argv) > 1 then
            repeat with i from 2 to (count of argv)
                set argList to argList & " " & quoted form of (item i of argv)
            end repeat
        end if
        launch_app(binaryPath, argList)
    end if
end run

on open theFiles
    -- Quando i file vengono trascinati, 'on run' non viene chiamato direttamente da macOS
    -- ma il nostro wrapper shell sì. Quindi gestiamo tutto tramite il wrapper.
end open

on launch_app(binaryPath, args)
    tell application "Terminal"
        activate
        do script quoted form of binaryPath & args
    end tell
end launch_app
EOF

    # Compila l'AppleScript
    osacompile -o "$APP_BUNDLE/Contents/Resources/launcher.scpt" "$APP_BUNDLE/Contents/Resources/launcher.applescript"
    rm "$APP_BUNDLE/Contents/Resources/launcher.applescript"

    # Crea un wrapper shell come CFBundleExecutable principale
    # Passa il percorso del binario come primo argomento, seguito dai file trascinati
    cat > "$APP_BUNDLE/Contents/MacOS/launcher" <<'EOF'
#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Note: BINARY_NAME is replaced by the actual name during build
EXE="$DIR/ImgCompress"
SCPT="$DIR/../Resources/launcher.scpt"
osascript "$SCPT" "$EXE" "$@"
EOF
    # Replace the placeholder BINARY_NAME in the launcher script
    sed -i '' "s/EXE=\"\$DIR\/ImgCompress\"/EXE=\"\$DIR\/$BINARY_NAME\"/" "$APP_BUNDLE/Contents/MacOS/launcher"
    chmod +x "$APP_BUNDLE/Contents/MacOS/launcher"

    # Info.plist
    cat > "$APP_BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>launcher</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.antedoro.imgcompress</string>
    <key>CFBundleName</key>
    <string>$APP_NAME_BASE</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>2.2</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>CFBundleDocumentTypes</key>
    <array>
        <dict>
            <key>CFBundleTypeName</key>
            <string>Images and Folders</string>
            <key>CFBundleTypeRole</key>
            <string>Viewer</string>
            <key>LSItemContentTypes</key>
            <array>
                <string>public.image</string>
                <string>public.folder</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
EOF
    
    chmod +x "$APP_BUNDLE/Contents/MacOS/$BINARY_NAME"
    echo "      ✔ $APP_BUNDLE creato."
}

# Esegui le build
create_bundle "amd64" "Intel"
create_bundle "arm64" "Silicon"

# Pulizia finale
rm -f "$ICON_ICNS"

echo "✅ Build completata!"
