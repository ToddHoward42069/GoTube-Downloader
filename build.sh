#!/bin/bash

# Config
APP_NAME="gotube"
VERSION="1.5.1"
DESCRIPTION="A resilient, cross-platform YouTube downloader built with Go and Fyne."
MAINTAINER="GoTube Developer"
ARCH="amd64" 

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Flags
BUILD_DEB=false
BUILD_APPIMAGE=false
BUILD_WINDOWS=false

for arg in "$@"
do
    case $arg in
        --deb) BUILD_DEB=true ;;
        --appimage) BUILD_APPIMAGE=true ;;
        --windows) BUILD_WINDOWS=true ;;
        --all) BUILD_DEB=true; BUILD_APPIMAGE=true; BUILD_WINDOWS=true ;;
    esac
done

echo -e "${BLUE}--- Cleaning up ---${NC}"
go clean
rm -rf dist
mkdir -p dist

echo -e "${BLUE}--- Tidy Modules ---${NC}"
go mod tidy

# Generate Icon
cat <<EOF > internal/gui/icon.svg
<svg width="512" height="512" viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
  <rect x="32" y="80" width="448" height="352" rx="64" ry="64" fill="#00c7ff"/>
  <path d="M352 256L192 352V160L352 256Z" fill="#FFFFFF"/>
</svg>
EOF
cp internal/gui/icon.svg icon.svg

create_desktop_file() {
    cat <<EOF > gotube.desktop
[Desktop Entry]
Type=Application
Name=GoTube
Comment=$DESCRIPTION
Exec=gotube
Icon=gotube
Terminal=false
Categories=Video;Network;
EOF
}

# --- 1. Linux Binary (Standardized Name) ---
echo -e "${GREEN}--- Building Linux Binary ---${NC}"
LINUX_BIN="gotube-linux-amd64"
go build -ldflags "-s -w -X 'gotube/internal/models.AppVersion=v$VERSION'" -o dist/$LINUX_BIN ./cmd/gotube

# --- 2. Windows Binary (Standardized Name) ---
if [ "$BUILD_WINDOWS" = true ]; then
    echo -e "${GREEN}--- Building Windows Binary ---${NC}"
    WIN_BIN="gotube-windows-amd64.exe"
    if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        echo -e "${RED}MinGW not found. Skipping Windows build.${NC}"
    else
        CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
        go build -ldflags "-s -w -H=windowsgui -X 'gotube/internal/models.AppVersion=v$VERSION'" -o dist/$WIN_BIN ./cmd/gotube
    fi
fi

# --- 3. DEB Package ---
if [ "$BUILD_DEB" = true ]; then
    echo -e "${GREEN}--- Building .DEB ---${NC}"
    DEB_DIR="dist/deb/${APP_NAME}_${VERSION}_${ARCH}"
    mkdir -p "$DEB_DIR/usr/local/bin" "$DEB_DIR/usr/share/applications" "$DEB_DIR/usr/share/icons/hicolor/scalable/apps" "$DEB_DIR/DEBIAN"
    
    # Use the linux binary we just built
    cp dist/$LINUX_BIN "$DEB_DIR/usr/local/bin/$APP_NAME"
    chmod +x "$DEB_DIR/usr/local/bin/$APP_NAME"
    
    create_desktop_file
    mv gotube.desktop "$DEB_DIR/usr/share/applications/"
    cp icon.svg "$DEB_DIR/usr/share/icons/hicolor/scalable/apps/gotube.svg"

    cat <<EOF > "$DEB_DIR/DEBIAN/control"
Package: $APP_NAME
Version: $VERSION
Section: video
Priority: optional
Architecture: $ARCH
Maintainer: $MAINTAINER
Description: $DESCRIPTION
Depends: libc6, libgl1, libx11-6
EOF
    if command -v dpkg-deb &> /dev/null; then
        dpkg-deb --build "$DEB_DIR"
        # Rename final deb for consistency
        mv dist/deb/*.deb dist/gotube_${VERSION}_amd64.deb
        rm -rf dist/deb
    fi
fi

# --- 4. AppImage ---
if [ "$BUILD_APPIMAGE" = true ]; then
    echo -e "${GREEN}--- Building AppImage ---${NC}"
    APPDIR="dist/AppDir"
    mkdir -p "$APPDIR/usr/bin"
    cp dist/$LINUX_BIN "$APPDIR/usr/bin/gotube"
    cp icon.svg "$APPDIR/gotube.svg"
    cp icon.svg "$APPDIR/.DirIcon"
    create_desktop_file
    mv gotube.desktop "$APPDIR/gotube.desktop"
    
    cat <<EOF > "$APPDIR/AppRun"
#!/bin/bash
exec "\$APPDIR/usr/bin/gotube" "\$@"
EOF
    chmod +x "$APPDIR/AppRun"

    TOOL_URL="https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage"
    if [ ! -f "appimagetool" ]; then
        wget -q -O appimagetool "$TOOL_URL"
        chmod +x appimagetool
    fi
    ARCH=x86_64 ./appimagetool "$APPDIR" "dist/GoTube-${VERSION}-x86_64.AppImage"
    rm -rf "$APPDIR"
fi

echo -e "${BLUE}--- Build Complete. Assets in /dist ---${NC}"
ls -lh dist/