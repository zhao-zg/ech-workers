module github.com/zhao-zg/ech-workers

go 1.24

require github.com/zhao-zg/ech-workers/tunnel v0.0.0

require github.com/gorilla/websocket v1.5.3 // indirect

replace github.com/zhao-zg/ech-workers/tunnel => ./tunnel

replace github.com/zhao-zg/ech-workers/mobile => ./mobile
