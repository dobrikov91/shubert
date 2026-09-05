VERSION=$(cat version.txt)

# static isn't supported on Mac
env CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -v -o build/shubert-mac-arm64
env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -v -o build/shubert-mac-amd64

# copy data
cd ./build
cp -rf ../data .

lipo -create -output shubert-mac shubert-mac-arm64 shubert-mac-amd64
tar -czvf shubert-mac-$VERSION.tar.gz shubert-mac data
