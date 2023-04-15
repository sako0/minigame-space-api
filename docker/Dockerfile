FROM golang:1.20.3-alpine

# ログに出力する時間をJSTにするため、タイムゾーンを設定
ENV TZ /usr/share/zoneinfo/Asia/Tokyo

# ワーキングディレクトリの設定
WORKDIR /go/src/app

# ModuleモードをON
ENV GO111MODULE=on

# ホストのファイルをコンテナの作業ディレクトリに移行
COPY . .

# go.modを参照し、go.sumファイルの更新(不要を削除)を行う
RUN go mod tidy

RUN  go mod download \
    && go build -o /go/src/app/main ./cmd/api/main.go

EXPOSE 5500

CMD ["/go/src/app/main"]
