### To compile to rpi 32bit from windows:

- $env:GOOS='linux'
- $env:GOARCH="arm"
- go build


docker exec fe454d19e8ba ./pb/base deploy
