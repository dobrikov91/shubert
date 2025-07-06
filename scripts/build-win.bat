set VERSION=1.3.0

go build -v -ldflags="-X main.version=%VERSION% -extldflags=-static" -o build/shubert-win-x64.exe

# copy data
cd ./build
robocopy ../data ./data /e
robocopy ../templates ./templates /e

tar -cavf shubert-win-x64-%VERSION%.zip shubert-win-x64.exe data templates