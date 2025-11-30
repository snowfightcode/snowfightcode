# ADR: JavaScriptランタイムの選択

## ステータス

採用: 2025-11-24  
更新: 2025-11-30（ランタイムを fastschema/qjs → buke/quickjs-go に差し替え）

## コンテキスト

Snowfight: Codeは、プレイヤーがJavaScriptでロジックを記述するプログラミングゲームです。このゲームの核となる価値は**リソース制約下でのプログラミング**であり、限られた計算資源（命令数・メモリ）の中で最適な戦略を実装することを目標とします。

### 要件

1. **決定論的実行**: PCの性能に関わらず、同じコードは常に同じ結果を生成する必要がある
2. **リソースカウント**: 命令数やメモリ使用量を正確に測定・制限できる
3. **安定性**: 本番環境でクラッシュしない
4. **ビルドの容易さ**: クロスコンパイルや開発環境のセットアップが複雑でない

## 検討した選択肢

### 1. dop251/goja (Pure Go JavaScript Engine)

**メリット:**
- Pure Goのため、Cgoが不要
- ビルドが簡単（`go build`のみ）
- クロスコンパイルが容易

**デメリット:**
- **決定論性の欠如**: タイムアウト制限しか設定できず、「1秒でタイムアウト」では高性能PCと低性能PCで実行できる命令数が大きく異なる
- リソースカウントが事実上不可能

**判定:** ❌ 却下（要件1, 2を満たさない）

### 2. lithdew/quickjs (Cgo版QuickJS)

**メリット:**
- QuickJSの機能をフルに利用可能
- 命令数カウント・メモリ制限が可能

**デメリット:**
- **実行時クラッシュ**: SIGSEGV/SIGBUSエラーが発生
- Cgoのメモリ管理の複雑さ
- スレッド制約によるデバッグの困難さ

**判定:** ❌ 却下（要件3を満たさない）

### 3. quickjs-go/quickjs-go (Cgo版QuickJS)

**メリット:**
- QuickJSの機能をフルに利用可能
- 活発に開発されている

**デメリット:**
- **リンクエラー**: 静的ライブラリがGoモジュールに含まれていない
- ビルド時に外部の静的ライブラリが必要

**判定:** ❌ 却下（要件4を満たさない）

### 4. buke/quickjs-go (Cgo版QuickJS)

**メリット:**
- 日本語ドキュメントあり
- 活発に開発されている
- `SetExecuteTimeout` で QuickJS の割り込みハンドラを設定でき、無限ループを中断可能

**当初の懸念 (2025-11-24 時点):**
- AI コーディング用のサンドボックス環境で `go mod vendor` した際、`deps/include/quickjs.h` などが vendor に入らず、コンパイルエラーになった。
- これを「Cソース欠如」と誤認し、要件4（ビルド容易性）を満たさないとして却下。

**再評価 (2025-11-30):**
- 原因は「サンドボックスでネットアクセスが制限され、`go mod download` が走らなかった」ことによる依存欠落だった。
- オンラインで `go mod download github.com/buke/quickjs-go@v0.6.6` を実行すれば、モジュールキャッシュに `deps` 一式が展開され、コンパイル可能であると確認。
- 従って、ビルド容易性の懸念は解消できると判断。

**判定:** ⭕ 採用（再評価後）

### 5. fastschema/qjs (WASM版QuickJS)

**メリット:**
- ✅ **Cgo不要**: Pure Goと同様にビルド可能
- ✅ **QuickJSの機能**: 命令数カウント・メモリ制限が実装可能
- ✅ **安定性**: WASMサンドボックス内で実行されるため、クラッシュのリスクが低い
- ✅ **クロスコンパイル容易**: WASMランタイム(Wazero)経由なので環境依存性がない
- ✅ **モダンな実装**: ES2023、async/awaitをサポート

**デメリット:**
- Pure Goより若干オーバーヘッドがある（しかし許容範囲内）

**判定:** ✅ 採用（2025-11-24 時点）。ただし後述の更新で差し替え。

## 2025-11-30 更新: buke/quickjs-go への差し替え

### 背景
- `runtime.tick_timeout_ms` に基づくティックタイムアウトが、`fastschema/qjs` では実際には効いていなかった。原因は、WASM 側の `MaxExecutionTime` が QuickJS の割り込みハンドラにバインドされておらず、無限ループの JS を停止できなかったこと。
- 試合全体がハングする重大な欠陥のため、タイムアウトが確実に効く実装を最優先に再検討した。

### 追加検討
- `github.com/buke/quickjs-go` (CGO, prebuilt libquickjs 同梱) を再評価。`SetExecuteTimeout` が C 層で割り込みハンドラを設定しており、無限ループを中断できることを確認。
- サンドボックス環境での依存欠落は、`go mod download` を先に実行すれば回避できることを確認し、ビルド容易性の懸念を解消。
- CGO 前提になるが、ビルド手順を README に明記することで運用リスクを許容。

### 決定
**buke/quickjs-go** を採用し、`fastschema/qjs` を廃止。

### 採用理由（差し替え後）
1. **確実なタイムアウト**: `SetExecuteTimeout` によりティック単位の時間制限を強制できる。
2. **prebuilt バイナリ同梱**: `deps/libs/<os>_<arch>` に libquickjs が同梱されており、追加ビルド不要。
3. **API 互換性の確保**: 既存の API ラッパーを小変更で移行できた。

### トレードオフ
- **CGO 依存**: クロスコンパイルがやや厄介になる。CI と開発環境で `CGO_ENABLED=1` とコンパイラの存在を保証する。
- **ビルドオプション**: macOS では Gatekeeper 回避のため `-ldflags="-buildid="` を推奨。

### 実装メモ
- ランタイム初期化で `WithExecuteTimeout` ではなく、各ティック実行前に `SetInterruptHandler` でミリ秒単位の監視をセット（`runtime.go`）。
- タイムアウト発生時はエラーではなく Warning (`execution timed out`) を返し、試合継続を優先。
- 依存更新: `go.mod` を `github.com/buke/quickjs-go v0.6.6` に変更。vendor は使わず、モジュールキャッシュ経由で prebuilt を取得。

## 参考資料

- [fastschema/qjs GitHub](https://github.com/fastschema/qjs)
- [QuickJS公式サイト](https://bellard.org/quickjs/)
- [Wazero (WebAssembly Runtime)](https://github.com/tetratelabs/wazero)
