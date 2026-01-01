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

    # Compila
    CGO_ENABLED=0 GOOS=darwin GOARCH=$ARCH go build -o "$APP_BUNDLE/Contents/MacOS/$BINARY_NAME" main.go
    if [ $? -ne 0 ]; then
        echo "❌ Errore compilazione per $ARCH"
        return
    fi

    # Copia Icona
    if [ -f "$ICON_ICNS" ]; then
        cp "$ICON_ICNS" "$APP_BUNDLE/Contents/Resources/"
    fi

    # Launcher Script
    LAUNCHER="$APP_BUNDLE/Contents/MacOS/launcher"
    cat > "$LAUNCHER" <<EOF
#!/bin/bash
DIR="\$( cd "\$( dirname "\${BASH_SOURCE[0]}" )" && pwd )"
EXE="\$DIR/$BINARY_NAME"

ARGS=""
for arg in "\$@"; do
    ESCAPED_ARG=\$(printf '%s\n' "\$arg" | sed 's/\\\\/\\\\\\\\/g; s/"/\\\\"/g')
    ARGS="\$ARGS \"\$ESCAPED_ARG\""
done

FULL_CMD="\"\$EXE\"\$ARGS"
AS_CMD=\$(printf '%s\n' "\$FULL_CMD" | sed 's/\\\\/\\\\\\\\/g; s/"/\\\\"/g')

osascript <<END
tell application "Terminal"
    activate
    do script "\$AS_CMD"
end tell
END
EOF
    chmod +x "$LAUNCHER"

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
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>Terminal</key>
    <string>false</string>
    <key>LSUIElement</key>
    <true/>
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
