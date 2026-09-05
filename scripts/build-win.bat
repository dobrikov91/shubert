set /p VERSION=<version.txt

go build -v -ldflags="-extldflags=-static" -o build/shubert-win-x64.exe

# copy data
cd ./build
robocopy ../data ./data /e

tar -cavf shubert-win-x64-%VERSION%.zip shubert-win-x64.exe data