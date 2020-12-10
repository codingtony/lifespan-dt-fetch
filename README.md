# lifespan-dt-fetch
Go program to get values from LifeSpan DT walking desk threadmill using Bluetooth (BLE)



## Install

```
go install github.com/codingtony/lifespan-dt-fetch
```



## Build

```bash
go build
```



## Use  (tested on Linux only)

Requires `sudo` to access `hci0`

```
sudo lifespan-dt-fetch 
Initializing Bluetooth
Scanning for 30s...
Distance: 12.31 [ A1 AA 0C 1F 00 00 ]
Calories: 1128 [ A1 AA 04 68 00 00 ]
Steps:    27341 [ A1 AA 6A CD 00 00 ]
Time:     7:46:40 [ A1 AA 07 2E 28 00 ]
Speed:    0.0 [ A1 AA 00 00 00 00 ]
Activity saved to : ./lifespan_20201209T223340Z_7h46m40s_27341steps.tcx
Disconnecting [ 00:0c:bf:37:b2:84 ]... (this might take up to few seconds on OS X)
[ 00:0c:bf:37:b2:84 ] is disconnected 
```

