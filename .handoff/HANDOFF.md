# Handoff: Device.ReadRSSI on the nRF52 SoftDevice

This directory is scratch for the handoff. **Drop the commit that adds
`.handoff/` before opening the upstream PR.** It must never reach
tinygo-org/bluetooth.

## Where things stand

Branch `read-rssi` on github.com/toddwilson/tinygo-bluetooth, on top of
tinygo-org/bluetooth `dev` at 8e7be8f (the fork's `dev` is synced to it).

Commit 283ece2 `gap: add Device.ReadRSSI to read the RSSI of a connection`:

- `ReadRSSI() (int16, error)` added to `GAPDevice` in `gap.go`.
- Implemented on macOS in `gap_darwin.go` with CBPeripheral readRSSI. The
  result channel is buffered and the delegate never blocks, so a late
  callback after a timeout cannot stall CoreBluetooth.
- Stubs return `errNotYetImplmented` on Linux, Windows, HCI, nrf51 and
  nrf528xx central. `errNotSupported` on WASM and s113v7.

Verified on macOS:
- `go test ./ ./ble ./hci` and `go test -race` pass.
- `make smoketest-macos smoketest-linux smoketest-windows` pass.
- `make smoketest-tinygo smoketest-wasm` pass in `ghcr.io/tinygo-org/tinygo-dev`.
- Hardware. macOS 27.0.1 connected to a BLE wind sensor named ULTRASONIC.
  Five ReadRSSI calls each returned -71 dBm with no error in about 50 ms.

## Your task

Implement `ReadRSSI` for the SoftDevice central and test it on the
nRF52840-DK (`pca10056-s140v7`). Upstream issue #154 is the request, and a
maintainer named `sd_ble_gap_rssi_get` there. Add it as a second commit on
`read-rssi`, separate from 283ece2.

Code: the stub is in `gap_nrf528xx-central.go` (build tags s132v6, s140v6,
s140v7). `d.connectionHandle` is the handle to pass.

SoftDevice API, same signature in every vendored header under `softdevice/`:

- `sd_ble_gap_rssi_start(conn_handle, threshold_dbm, skip_count)`. Pass
  `BLE_GAP_RSSI_THRESHOLD_INVALID` (0xFF) to disable
  `BLE_GAP_EVT_RSSI_CHANGED` events, since nothing handles them.
- `sd_ble_gap_rssi_get(conn_handle, *int8 rssi, *uint8 ch_index)`.
  Returns `NRF_ERROR_INVALID_STATE` if reporting was not started, and
  `NRF_ERROR_NOT_FOUND` until the first sample after starting.
- See s140 7.3.0 `ble_gap.h` around line 2494 to 2535.

Suggested shape. On `NRF_ERROR_INVALID_STATE`, call `rssi_start` and retry.
On `NRF_ERROR_NOT_FOUND`, wait a connection interval and retry, up to a short
timeout. Starting once per connection is enough. Consider whether to start
reporting in the `BLE_GAP_EVT_CONNECTED` handler instead (`adapter_nrf528xx-full.go`).

Check these while you are there:
- The `ble_gap.h` note on ERRATA-153 and ERRATA-225 says the sample needs
  temperature compensation. Find out whether the SoftDevice already does
  it, and cite the source in a comment if the code has to handle it.
- `rssi_get` takes `*C.int8_t`. Convert to int16 for the return value.

## Testing on the DK

The DK is the central here, so it needs a connectable peripheral nearby. A
phone running nRF Connect with an advertiser configured works, or any BLE
device you own. Write a small TinyGo program that scans, connects by name,
calls ReadRSSI several times while the peripheral is moved, and prints the
results. Flash with `tinygo flash -target=pca10056-s140v7`, read with
`tinygo monitor`. The DK needs the s140 7.3.0 SoftDevice flashed, see the
README. `.handoff/rssitest-macos/main.go` is the macOS version for reference.
Do not add the test program to the repo.

## Upstream standards (follow these)

- `AGENTS.md`. Comments at most 2 lines, few colons, semicolons or dashes,
  cite a definitive source (datasheet section, URL) when a comment explains
  a hardware or API quirk. **No coding tool attribution anywhere**, so no
  Co-Authored-By lines in commits or PRs.
- Commit and PR titles look like `area: lowercase summary`. Plain direct
  bodies. Recent PRs do not touch CHANGELOG.md.
- Before pushing, run `go test ./ ./ble ./hci`, `go test -race ./ ./ble ./hci`,
  `make smoketest-linux smoketest-windows smoketest-tinygo smoketest-wasm`
  (TinyGo locally or in `ghcr.io/tinygo-org/tinygo-dev`).

## Upstream PR (not opened yet)

From `toddwilson:read-rssi` into `tinygo-org/bluetooth:dev`. Draft body,
update the platform list and testing lines once the SoftDevice works:

> Updates #154
>
> ScanResult.RSSI only holds the value from the last advertisement, so there
> was no way to read the signal strength once connected. ReadRSSI is now part
> of GAPDevice and is implemented on macOS using CBPeripheral readRSSI. Linux,
> Windows, HCI, nrf51 and the nrf528xx SoftDevices return errNotYetImplmented.
> WebBluetooth and s113v7 return errNotSupported, since they cannot provide it.
>
> Tested on macOS 27.0.1 against an ultrasonic wind sensor. Five reads each
> returned -71 dBm in about 50 ms. The smoketest-macos, smoketest-linux and
> smoketest-windows builds pass, and so do smoketest-tinygo and smoketest-wasm
> in the tinygo-dev container. Linux, Windows, HCI and SoftDevice are
> compile-tested only.

Before opening it, drop the `.handoff/` commit
(`git rebase --onto <commit>^ <commit>` or an interactive rebase) and force push.
