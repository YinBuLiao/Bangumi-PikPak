module bangumi-pikpak

go 1.25.0

require (
	github.com/PuerkitoBio/goquery v1.9.2
	github.com/go-sql-driver/mysql v1.8.1
	github.com/kanghengliu/pikpak-go v0.1.0
	github.com/mmcdole/gofeed v1.3.0
	github.com/redis/go-redis/v9 v9.7.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	modernc.org/sqlite v1.29.10
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mmcdole/goxpp v1.1.1-0.20240225020742-a0c311522b23 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	modernc.org/gc/v3 v3.0.0-20240107210532-573471604cb6 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
	modernc.org/strutil v1.2.1 // indirect
	modernc.org/token v1.1.0 // indirect
)

replace github.com/kanghengliu/pikpak-go => ./third_party/pikpak-go
