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
Scanning for 30s...
Distance: 0.43 [ A1 AA 00 2B 00 00 ]
Calories: 40 [ A1 AA 00 28 00 00 ]
Steps:    998 [ A1 AA 03 E6 00 00 ]
Time:     0:16:28 [ A1 AA 00 10 1C 00 ]
Speed:    1.60 [ A1 AA 01 3C 00 00 ]
Disconnecting [ 00:0c:bf:37:b2:84 ]... (this might take up to few seconds on OS X)
[ 00:0c:bf:37:b2:84 ] is disconnected 
```

