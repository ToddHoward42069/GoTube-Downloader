#!/bin/bash
APP_NAME="gotube"
echo "--- Cleaning up ---"
go clean
rm -f $APP_NAME $APP_NAME.exe
echo "--- Tidy Modules ---"
go mod tidy
echo "--- Building for Host ---"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    go build -ldflags "-s -w -H=windowsgui" -o $APP_NAME.exe ./cmd/gotube
else
    go build -ldflags "-s -w" -o $APP_NAME ./cmd/gotube
fi
echo "--- Build Complete ---"
