set shell := ["powershell", "-Command"]

run command="create":
  go run ./cli/main.go {{command}}

test:
  go test -v ./...

build:
  go build -o cmdforge.exe ./cli/main.go
  @echo "✔ 実行ファイルをビルドしました"
