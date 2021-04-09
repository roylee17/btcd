module github.com/lbryio/lbrycrd.go

go 1.16

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v1.0.2
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd
	github.com/btcsuite/goleveldb v1.0.0
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792
	github.com/btcsuite/winsvc v1.0.0
	github.com/davecgh/go-spew v1.1.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
)

replace github.com/btcsuite/btcd => ./

replace github.com/btcsuite/btcutil => github.com/roylee17/lbcutil v1.0.2-0.20210411112141-c6d6fdf3dbbe
