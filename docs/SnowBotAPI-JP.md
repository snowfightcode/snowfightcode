# SnowBot API 一覧

## 移動系

* `move(distance: Integer): void`

  * `distance` は1ティックでの移動距離。正で前進、負で後退。
  * 範囲: `MIN_MOVE <= |distance| <= MAX_MOVE`
  * 引数の値が範囲を超えた場合、`distance` は範囲内の値に丸められる。
  * `distance = 0` は無行動（No-op）。
  * フィールド外への移動はできない（**境界に留まる**）。
    * 例えば境界まで残り3pxしかないのに `MIN_MOVE=5` のとき、3pxまでは動く。
    * 境界に留まる事態が発生したティックは「成功扱い」。イベントログには記録しない。
  * 他Bot／ブロックに対しても重なりは発生させない（最初の接触点まで移動して停止）。

* `turn(angle: Integer): void`

  * 角度は整数。正は右回り、負は左回り。
  * 角度基準は北が0度。360超/負は`angle % 360` に正規化する。
  * `angle = 0` は無行動（No-op）。

## 雪玉操作

* `toss(distance: Integer): void`

  * 現在の向いている方向（`angle`）に、狙う `distance` で雪玉を投げる。
  * `distance` は命中地点の中心点のターゲット距離。最大値は `<MAX_FLYING_DISTANCE>`。
  * 飛翔速度は `<SNOWBALL_SPEED>` / tick 、命中半径は `<DAMAGE_RADIUS>`（いずれもフィールド単位）。
  * 弾道は直進のみで、重力・落下などは考慮しない。
  * 投擲後の雪玉は**ティックごとに`SNOWBALL_SPEED`で移動**して当たり判定を繰り返す。
  * 途中にブロックがある場合は遮蔽物として機能し、貫通はしない（その場で消滅）。
  * 雪玉は境界外で消滅する。自爆判定は行わない。
  * `distance`が負の場合、0とみなす。
  * `distance`が0の場合、No-opとする。雪玉は消費しない。


## プログラム実行制約

* メモリ最大値:`<MAX_MEMORY_BYTES>`
* スタック最大値:`<MAX_STACK_BYTES>`
* 1ティックは `<TICK_TIMEOUT>` ミリ秒で終了する。
* 違反時はリソースエラーとして該当のSnowBotの実行を中止する。


## ゲームパラメーター

* `MAX_MEMORY_BYTES`: メモリ最大値
* `MAX_STACK_BYTES`: スタック最大値
* `TICK_TIMEOUT`: 1ティックの最大時間
* `FIELD_SIZE`: フィールドのサイズ
* `MAX_MOVE`: 1ティックでの移動距離の最大値
* `MIN_MOVE`: 1ティックでの移動距離の最小値
* `MAX_FLYING_DISTANCE`: 雪玉の最大飛行距離
* `SNOWBALL_SPEED`: 雪玉の移動速度
* `DAMAGE_RADIUS`: 雪玉の命中半径
* `MAX_SNOWBALL`: 所持雪玉の最大数
* `MAX_FLYING_SNOWBALL`: 飛行中の雪玉の最大数
