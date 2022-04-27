module github.com/filedrive-team/go-graphsplit

go 1.15

require (
	github.com/beeleelee/go-ds-rpc v0.0.4
	github.com/docker/go-units v0.4.0
	github.com/filecoin-project/go-address v0.0.6 // indirect
	github.com/filecoin-project/go-commp-utils v0.1.2
	github.com/filecoin-project/go-fil-markets v1.13.4 // indirect
	github.com/filecoin-project/go-padreader v0.0.1
	github.com/filecoin-project/go-state-types v0.1.3
	github.com/filecoin-project/lotus v1.14.3
	github.com/ipfs/go-blockservice v0.1.7
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-cidutil v0.0.2
	github.com/ipfs/go-datastore v0.4.6
	github.com/ipfs/go-ipfs-blockstore v1.0.4
	github.com/ipfs/go-ipfs-chunker v0.0.5
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipfs-files v0.0.9
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/ipfs/go-merkledag v0.4.1
	github.com/ipfs/go-unixfs v0.2.6
	github.com/ipld/go-car v0.3.2-0.20211001225732-32d0d9933823
	github.com/ipld/go-ipld-prime v0.14.2
	github.com/multiformats/go-multibase v0.0.3
	github.com/urfave/cli/v2 v2.3.0
	github.com/xuri/excelize/v2 v2.6.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
