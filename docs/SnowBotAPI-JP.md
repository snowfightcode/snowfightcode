# ゲームのルール

1. ユーザーはゲーム内で提供されるAPIを用いて、雪合戦ロボット **SnowBot** を操る。プレイヤー数は最大 `<match.max_players>` まで可変。
2. SnowBotは最大で **<snowbot.max_snowball>** 個までの雪玉（Snowball）を作成して搭載できる。
3. SnowBotは搭載している雪玉を投げて、他のSnowBotに当てることができる。
4. 雪玉が命中したSnowBotは **<snowball.damage>** ポイントのHPを失う。
5. SnowBotのHPの初期値は **<snowbot.max_hp>** ポイントである。
6. 対戦時間は **<match.max_ticks>** ティック。
7. **勝敗条件**: 相手のHPを0にした側が勝利。時間切れ時・同時撃破は勝者なし。

# SnowBot API 一覧

## 移動系

* `move(distance: Integer): void`

  * `distance` は1ティックでの移動距離。正で前進、負で後退。
  * 範囲: `snowbot.min_move <= |distance| <= snowbot.max_move`
  * 引数の値が範囲を超えた場合、`distance` は範囲内の値に丸められる。
  * `distance = 0` は無行動（No-op）。
  * **同一ティック内の複数呼び出しは無効化される（最初の1回のみ反映）**。
  * フィールド外への移動はできない（**境界に留まる**）。
    * 例えば境界まで残り3pxしかないのに `snowbot.min_move=5` のとき、3pxまでは動く。
    * 境界に留まる事態が発生したティックは「成功扱い」。イベントログには記録しない。
  * 他Botとの衝突判定は行わない。

* `turn(angle: Integer): void`

  * 角度は整数。正は右回り、負は左回り。
  * 角度基準は北が0度。360超/負は`angle % 360` に正規化する。
  * `angle = 0` は無行動（No-op）。
  * **同一ティック内の複数呼び出しは無効化される（最初の1回のみ反映）**。

## 雪玉操作

* `toss(distance: Integer): void`

  * 現在の向いている方向（`angle`）に、狙う `distance` で雪玉を投げる。
  * `distance` は命中地点の中心点のターゲット距離。最大値は `<snowball.max_flying_distance>`。
  * 飛翔速度は `<snowball.speed>` / tick 、命中半径は `<snowball.damage_radius>`（いずれもフィールド単位）。
  * 弾道は直進のみで、重力・落下などは考慮しない。
  * 投擲後の雪玉は**ティックごとに`snowball.speed`で移動**して当たり判定を繰り返す。
  * 雪玉は境界外で消滅する。投擲者にも命中する可能性がある。
  * `distance`が負の場合、0とみなす。
  * `distance`が0の場合、No-opとする。雪玉は消費しない。
  * **同一ティック内の複数呼び出しは無効化される（最初の1回のみ反映）**。

## センサー系

* `scan(angle: Integer, resolution: Integer): FieldObject[]`

  * `angle` 方向を中心に、`resolution`（度）内の敵をスキャン。
  * 返値はオブジェクトタイプ（SnowBot）、角度、距離の配列。
  * 角度基準は北が0度。360超/負は`angle % 360` に正規化する。
  * スキャン原点はBot中心。
  * `resolution` の範囲: `MIN_SCAN <= resolution <= MAX_SCAN`。`resolution=0` の場合、戻り値は空配列。
  * 入力範囲外（例: `resolution < MIN_SCAN`）の場合、戻り値は空配列。
  * 視野は扇形FOV。角度範囲は **[angle - resolution/2, angle + resolution/2)** （半開区間）。
  * レイキャスト遮蔽なし（遮蔽物越しでも検知）。
  * 検知距離: min=1, max=フィールド対角線長。
  * 返却の整列は**距離昇順、距離同値時は角度昇順**。自己は除外。
  * 同一ティック内は同一スナップショットを返す（再呼び出しで不変）。

* `position(): Position`

  * 自分の座標を取得

* `direction(): Integer`

  * 自分の向きを返す。

## 状態管理

* `hp(): Integer`

  * HP残量を返す。

* `max_hp(): Integer`

  * 最大HPを返す。

* `snowball_count(): Integer`

  * 搭載雪玉の数を返す。

* `max_snowball(): Integer`

  * 最大搭載雪玉数を返す。

# 警告出力（JSONL）

* 不正なAPIコールがあった場合、そのティックの標準出力に **警告レコード** をJSONLで追記します（状態レコードより先に出力）。
* レコード形式は `Type` フィールドで判別できます。

  * 状態レコード（従来＋`type`付与）
    * `{ "type": "state", "tick": 12, "players": [...], "p1": {...}, "p2": {...}, "snowballs": [...] }`

  * 警告レコード（stateに警告情報を付加）
    * `{ "type": "warning", "tick": 12, "players": [...], "p1": {...}, "p2": {...}, "snowballs": [...], "warnedPlayer": 2, "api": "move", "args": ["5", "10"], "warning": "called multiple times in one tick" }`

* 1ティックあたりの警告上限は3件（超過分は破棄）。
* 主な発生ケース:
  * 引数不足・型不正など
  * 同一ティック内で `move`/`turn`/`toss` を2回以上呼んだ場合（2回目以降は無効化＋警告）
  * APIラッパーが無効化して `null` を返すケース


# プログラム実行制約

* メモリ最大値:`<runtime.max_memory_bytes>`
* スタック最大値:`<runtime.max_stack_bytes>`
* 1ティックは `<runtime.tick_timeout_ms>` ミリ秒で終了する。
* 違反時はリソースエラーとして該当のSnowBotの実行を中止する。


# ゲームパラメーター

* `match.max_ticks`: 対戦時間（ティック数）
* `match.max_players`: 同時参加できるプレイヤー数の上限
* `match.random_seed`: 0以外なら乱数シード（スポーン位置や将来のランダム要素用、テスト向け）
* `field.width`: フィールドの幅
* `field.height`: フィールドの高さ
* `snowbot.min_move`: 1ティックでの移動距離の最小値
* `snowbot.max_move`: 1ティックでの移動距離の最大値
* `snowbot.max_hp`: SnowBotの最大HP
* `snowbot.max_snowball`: 所持雪玉の最大数
* `snowbot.max_flying_snowball`: 飛行中の雪玉の最大数
* `snowball.max_flying_distance`: 雪玉の最大飛行距離
* `snowball.speed`: 雪玉の移動速度
* `snowball.damage_radius`: 雪玉の命中半径
* `snowball.damage`: 雪玉の命中ダメージ
* `runtime.max_memory_bytes`: メモリ最大値
* `runtime.max_stack_bytes`: スタック最大値
* `runtime.tick_timeout_ms`: 1ティックの最大時間
