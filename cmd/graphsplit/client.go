package main

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-cidutil/cidenc"
	"github.com/multiformats/go-multibase"
	"github.com/urfave/cli/v2"
	"github.com/xuri/excelize/v2"
	"golang.org/x/xerrors"
	"os"
	"path/filepath"
	"strconv"
)

const (
	OfflineDeal = "offline"
	OnlineDeal = "online"
)

var clientCmd = &cli.Command{
	Name: "client",
	Usage: "Import data, Make deal",
	Subcommands: []*cli.Command{
		importCmd,
	},
}

var importCmd = &cli.Command{
	Name: "deal",
	Usage: "import and send deal from xlsx",
	ArgsUsage: "file-path",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "specify address to fund the deal with",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {

		if cctx.NArg() != 1 {
			return xerrors.New("expected input path as the only arg")
		}
		path := cctx.Args().First()
		path, err := filepath.Abs(path)
		if err != nil {
			return xerrors.Errorf("file path err : %v", err)
		}
		from := cctx.String("from")
		if from == "" {
			return xerrors.Errorf("failed get 'from' address: %w", err)
		}

		f, err := excelize.OpenFile(path)
		if os.IsNotExist(err) {
			return xerrors.Errorf("file: %s not exit", path)
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		list, err := api.ClientListImports(ctx)
		if err != nil {
			return xerrors.Errorf("client local list err : %v ", err)
		}

		cidMap := make(map[string]bool)
		encoder, err := GetCidEncoder(cctx)
		if err != nil {
			return xerrors.Errorf("get encoder err : %v ", err)
		}

		for _, v := range list {
			if v.Root != nil {
				cidMap[encoder.Encode(*v.Root)] = true
			}
		}

		sheetname := "Sheet1"
		rows, err := f.GetRows(sheetname)
		if err != nil {
			return xerrors.Errorf("get sheet err : %v ", err)
		}

		for i, r := range rows[1:]{
			if len(r) != 9 {
				log.Errorf("row：%d, data is incomplete，skip.....", i+1)
				continue
			}
			payloadCid := r[0]
			pieceCid := r[2]
			pieceSize := r[3]
			carPath := r[4]
			minerId := r[5]
			duration := r[6]
			dealId := r[7]
			dealType := r[8]
			if minerId == "0" || duration == "0" || dealId != "0" {
				log.Errorf("Unexecuted, minerId : %v, duration: %v, dealId: %v err", minerId, duration, dealId)
				continue
			}
			_, ok := cidMap[payloadCid]
			if !ok {
				ref := lapi.FileRef{
					Path:  carPath,
					IsCAR: true,
				}
				c, err := api.ClientImport(ctx, ref)
				if err != nil {
					log.Errorf("import err : %v", err)
					continue
				}
				log.Infof("payload cid : %v, path: %v is to be imported, Root: %v", payloadCid, carPath, c.Root)

			}
			log.Infof("paloadcid: %v start to send deal.....", payloadCid)

			cur, err := strconv.ParseUint(duration, 10, 32)
			if err != nil {
				log.Errorf("Parse duration: %v err %v", duration, err)
				continue
			}

			ps, err := strconv.ParseUint(pieceSize, 10, 32)

			if err != nil {
				log.Errorf("Parse pieceSize: %v err %v", duration, err)
				continue
			}

			proposeCid, err := sendVerifiedDeal(cctx, from, payloadCid, pieceCid, minerId, dealType, cur, ps, api)
			if err != nil {
				log.Errorf("Send deal failed, err : %v", err)
			}else{
				r[7] = proposeCid
				err :=f.SetSheetRow(sheetname, fmt.Sprintf("A%d", i+2), &r)
				if err != nil {
					log.Errorf("Sheet data set err : %v", err)
				}
			}

		}
		f.Save()
		log.Infof("Send deals end =========>")
		return nil

	},
}

func sendVerifiedDeal(cctx *cli.Context, fromAddr, dataCid, pieceCid, minerId, dealType string, dur, pieceSize uint64, api v0api.FullNode)(string, error){

	ctx := lcli.ReqContext(cctx)

	addr, err := address.NewFromString(fromAddr)
	if err != nil {
		return "", xerrors.Errorf("failed to parse 'from' address: %w", err)
	}
	// Check if the address is a verified client
	dcap, err := api.StateVerifiedClientStatus(ctx, addr, types.EmptyTSK)
	if err != nil {
		return "", xerrors.Errorf("Verified client status err: %v ", err)
	}
	if dcap == nil {
		return "", xerrors.Errorf("address : %v is not a verified client ", fromAddr)
	}

	if abi.ChainEpoch(dur) < build.MinDealDuration {
		return "", xerrors.Errorf("minimum deal duration is %d blocks", build.MinDealDuration)
	}
	if abi.ChainEpoch(dur) > build.MaxDealDuration {
		return "", xerrors.Errorf("maximum deal duration is %d blocks", build.MaxDealDuration)
	}

	data, err := cid.Parse(dataCid)
	if err != nil {
		return "", xerrors.Errorf("parse data cid err: %v", err)
	}
	ref := &storagemarket.DataRef{
		TransferType: storagemarket.TTGraphsync,
		Root:         data,
	}

	if dealType == OfflineDeal {
		ref.TransferType = storagemarket.TTManual
		c, err := cid.Parse(pieceCid)
		if err != nil {
			return "", xerrors.Errorf("parse piece cid err: %v", err)
		}
		ref.PieceCid = &c
		ref.PieceSize = abi.UnpaddedPieceSize(pieceSize)
	}

	miner, err := address.NewFromString(minerId)
	if err != nil {
		return "", xerrors.Errorf("failed to parse  minerId: %w", err)
	}

	sdParams := &lapi.StartDealParams{
		Data:               ref,
		Wallet:             addr,
		Miner:              miner,
		EpochPrice:         types.NewInt(0),
		MinBlocksDuration:  dur,
		DealStartEpoch:     0,
		FastRetrieval:      true,
		VerifiedDeal:       true,
		ProviderCollateral: big.NewInt(0),
	}

	var proposal *cid.Cid
	if dealType == OfflineDeal {
		proposal, err = api.ClientStatelessDeal(ctx, sdParams)
	}else if dealType == OnlineDeal{
		proposal, err = api.ClientStartDeal(ctx, sdParams)
	}else{
		return "", xerrors.Errorf("Deal type err : input=%s ", dealType)
	}

	if err != nil {
		return "", xerrors.Errorf("failed to send  deal: %v", err)
	}

	encoder, err := GetCidEncoder(cctx)
	if err != nil {
		return "", xerrors.Errorf("get encoder err : %v ", err)
	}

	return encoder.Encode(*proposal), nil

}


// GetCidEncoder returns an encoder using the `cid-base` flag if provided, or
// the default (Base32) encoder if not.
func GetCidEncoder(cctx *cli.Context) (cidenc.Encoder, error) {
	val := cctx.String("cid-base")

	e := cidenc.Encoder{Base: multibase.MustNewEncoder(multibase.Base32)}

	if val != "" {
		var err error
		e.Base, err = multibase.EncoderByName(val)
		if err != nil {
			return e, err
		}
	}

	return e, nil
}
