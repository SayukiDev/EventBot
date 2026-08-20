# EventBot

[![License](https://img.shields.io/badge/License-AGPL-blue.svg)](LICENSE)

小規模の使用を想定したシンプルな日程管理用Discord Bot

## 使用
```bash
./EventBot -config config.json
```

## 設定ファイル
```json
// config.json
{
  "log_level": "info",  // ログレベル, debug,info,warn,error,fatal
  "token": "token", // Discord Botのトークン
  "data_path": "./data.bson"  //　データファイルの保存位置
}
```

## Build
```bash
CGO_ENABLED=0 GOOS=linux go build -a -trimpath -ldflags="-s -w -X main.Version={{Version}}"  -o EventBot
```


## ライセンス
本プロジェクトは [GNUアフェロ一般公衆ライセンス](https://gpl.mhatta.org/agpl.ja.html) を基づき発行しております、
被配布者(ユーザー)には使用の自由・二次開発の自由・二次配布（販売含む）の自由などの自由が保証されます。
ただし二次配布の場合必ず同じ[GNUアフェロ一般公衆ライセンス](https://gpl.mhatta.org/agpl.ja.html) で発行するよう義務付けられます、
そして二次被配布者には本プロジェクトが被配布者に与えた自由と同じ自由が与えられます、
それらの自由を制限・干渉するあらゆる行為は一切認められません。