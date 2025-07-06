VERSION=1.3.0

# static isn't supported on Mac
env CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -v -ldflags="-X main.version=$VERSION" -o build/shubert-mac-arm64
env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -v -ldflags="-X main.version=$VERSION" -o build/shubert-mac-amd64

# copy data
cd ./build
cp -rf ../data .
cp -rf ../templates .

lipo -create -output shubert-mac shubert-mac-arm64 shubert-mac-amd64
tar -czvf shubert-mac-$VERSION.tar.gz shubert-mac data templates
