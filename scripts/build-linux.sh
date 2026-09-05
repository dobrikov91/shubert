VERSION=$(cat version.txt)

go build -v -o build/shubert-linux-x64

# copy data
cd ./build
cp -rf ../data .

tar -czvf shubert-linux-x64-$VERSION.tar.gz shubert-linux-x64 data
