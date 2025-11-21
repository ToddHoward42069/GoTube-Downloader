#!/bin/bash

# Config
APP_NAME="gotube"
VERSION="1.0.0"
DESCRIPTION="A resilient, cross-platform YouTube downloader built with Go and Fyne."
MAINTAINER="GoTube Developer <dev@gotube.local>"
ARCH="amd64" 

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Flags
BUILD_DEB=false
BUILD_APPIMAGE=false

# Parse Arguments
for arg in "$@"
do
    case $arg in
        --deb) BUILD_DEB=true ;;
        --appimage) BUILD_APPIMAGE=true ;;
        --all) BUILD_DEB=true; BUILD_APPIMAGE=true ;;
        --help)
        echo "Usage: ./build.sh [options]"
        echo "Options:"
        echo "  (default)   Build binary only"
        echo "  --deb       Build .deb package"
        echo "  --appimage  Build .AppImage file"
        echo "  --all       Build everything"
        exit 0
        ;;
    esac
done

# --- 1. Clean & Tidy ---
echo -e "${BLUE}--- Cleaning up ---${NC}"
go clean
rm -rf dist
rm -f $APP_NAME $APP_NAME.exe
mkdir -p dist

echo -e "${BLUE}--- Tidy Modules ---${NC}"
go mod tidy

# --- 2. Assets Generation (SVG) ---
# We generate the icon in internal/gui so go:embed can find it
echo -e "${BLUE}--- Generating SVG Icon ---${NC}"

generate_icon() {
cat <<EOF > internal/gui/icon.svg
<svg width="512" height="512" viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
  <rect x="32" y="80" width="448" height="352" rx="64" ry="64" fill="#00c7ff"/>
  <path d="M352 256L192 352V160L352 256Z" fill="#FFFFFF"/>
</svg>
EOF
}

generate_icon
# Copy it to root for packaging tools
cp internal/gui/icon.svg icon.svg

# Create .desktop file content
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

# --- 3. Build Binary ---
echo -e "${GREEN}--- Building Binary ---${NC}"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    go build -ldflags "-s -w -H=windowsgui" -o $APP_NAME.exe ./cmd/gotube
else
    go build -ldflags "-s -w" -o $APP_NAME ./cmd/gotube
fi

if [ ! -f "$APP_NAME" ] && [ ! -f "$APP_NAME.exe" ]; then
    echo -e "${RED}Binary build failed!${NC}"
    exit 1
fi

# --- 4. Build .DEB ---
if [ "$BUILD_DEB" = true ]; then
    echo -e "${GREEN}--- Building .DEB Package ---${NC}"
    
    DEB_DIR="dist/deb/${APP_NAME}_${VERSION}_${ARCH}"
    mkdir -p "$DEB_DIR/usr/local/bin"
    mkdir -p "$DEB_DIR/usr/share/applications"
    # SVG icons go into /usr/share/icons/hicolor/scalable/apps
    mkdir -p "$DEB_DIR/usr/share/icons/hicolor/scalable/apps"
    mkdir -p "$DEB_DIR/DEBIAN"

    cp $APP_NAME "$DEB_DIR/usr/local/bin/"
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
        echo -e "${GREEN}Success: dist/deb/${APP_NAME}_${VERSION}_${ARCH}.deb${NC}"
    else
        echo -e "${RED}Error: 'dpkg-deb' not found.${NC}"
    fi
fi

# --- 5. Build AppImage ---
if [ "$BUILD_APPIMAGE" = true ]; then
    echo -e "${GREEN}--- Building AppImage ---${NC}"
    
    APPDIR="dist/AppDir"
    mkdir -p "$APPDIR/usr/bin"
    
    cp $APP_NAME "$APPDIR/usr/bin/gotube"
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
        echo "Downloading AppImageTool..."
        wget -q -O appimagetool "$TOOL_URL"
        chmod +x appimagetool
    fi

    ARCH=x86_64 ./appimagetool "$APPDIR" "dist/${APP_NAME}-${VERSION}-x86_64.AppImage"
    echo -e "${GREEN}Success: dist/${APP_NAME}-${VERSION}-x86_64.AppImage${NC}"
fi

echo -e "${BLUE}--- Done ---${NC}"