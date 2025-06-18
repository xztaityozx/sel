# CLAUDE.md

このファイルはこのリポジトリでClaude Code (claude.ai/code) が作業する際のガイダンスを提供します。

## プロジェクト概要

**sel** はUnixの`cut`コマンドをawkライクなカラム選択とスライス記法で拡張したコマンドラインツールです。1インデックス記法、スライス範囲、正規表現、テンプレートベースの出力フォーマットを使用してテキスト入力からカラムを選択できます。

## 開発コマンド

### ビルド
```bash
make build          # バイナリをdist/selにビルド
make all           # クリーン、テスト、ビルドを実行
```

### テスト
```bash
make test          # 全テストを実行（先にビルドが必要）
go test -v ./...   # テストを直接実行
```

### クリーンアップ
```bash
make clean         # dist/ディレクトリを削除
```

## アーキテクチャ

コードベースはクリーンアーキテクチャの原則に従い、以下の主要パッケージで構成されています：

- **cmd/**: Cobraフレームワークを使用したCLIインターフェース
- **internal/column/**: カラム選択戦略（IndexSelector、RangeSelector、SwitchSelector）
- **internal/iterator/**: テキスト解析戦略（Iterator、RegexpIterator、PreSplitIterator）
- **internal/option/**: Viperを使用した設定管理
- **internal/output/**: 出力のフォーマットと書き込み
- **internal/parser/**: スライス記法と正規表現パターンのクエリ解析

### 主要インターフェース

- `column.Selector`: カラム選択の動作を定義
- `iterator.IEnumerable`: テキストのイテレーション/解析の動作を定義
- 全ての戦略はポリモーフィズムのためにこれらのインターフェースを実装

## カラムインデックス

- 1インデックスによるカラムアクセス（awkの慣例に従う）
- インデックス`0`は行全体を参照
- 末尾からの相対アクセス用の負のインデックスをサポート
- クエリ構文: `1`、`2:5`、`1::2`、`/start/:/end/`

## テスト

- 各パッケージの単体テスト（`*_test.go`）
- `test/e2e_test.go`のEnd-to-Endテスト
- テストは先にバイナリをビルドする必要がある（`make test`で処理される）

## 依存関係

go.modで管理される主要な依存関係：
- Cobra v1.8.1 - CLI用
- Viper v1.20.1 - 設定管理用
- Testify v1.10.0 - テスト用

## パフォーマンスに関する考慮事項

- Iteratorパターンによる遅延評価
- 多数のカラムアクセスのシナリオ用のPreSplitIterator
- 出力ライターでのバッファードI/O