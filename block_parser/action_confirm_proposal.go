package block_parser

import (
	"das-account-indexer/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"strconv"
)

func (b *BlockParser) ActionConfirmProposal(req *FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	log.Info("das tx:", req.Action, req.TxHash)

	mapPreBuilder, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	mapBuilder, err := witness.AccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}
	var accounts []tables.TableAccountInfo
	var records []tables.TableRecordsInfo
	for _, builder := range mapBuilder {
		oID, mID, oCT, mCT, oA, mA := core.FormatDasLockToHexAddress(req.Tx.Outputs[builder.Index].Lock.Args)
		accounts = append(accounts, tables.TableAccountInfo{
			BlockNumber:        req.BlockNumber,
			BlockTimestamp:     req.BlockTimestamp,
			Outpoint:           common.OutPoint2String(req.TxHash, uint(builder.Index)),
			AccountId:          builder.AccountId,
			NextAccountId:      builder.NextAccountId,
			Account:            builder.Account,
			OwnerChainType:     oCT,
			Owner:              oA,
			OwnerAlgorithmId:   oID,
			ManagerChainType:   mCT,
			Manager:            mA,
			ManagerAlgorithmId: mID,
			Status:             tables.AccountStatus(builder.Status),
			RegisteredAt:       builder.RegisteredAt,
			ExpiredAt:          builder.ExpiredAt,
		})
		if _, ok := mapPreBuilder[builder.Account]; ok {
			list := builder.RecordList()
			for _, v := range list {
				records = append(records, tables.TableRecordsInfo{
					Account: builder.Account,
					Key:     v.Key,
					Type:    v.Type,
					Label:   v.Label,
					Value:   v.Value,
					Ttl:     strconv.FormatUint(uint64(v.TTL), 10),
				})
			}
		}
	}

	if err = b.DbDao.UpdateAccountInfoList(accounts, records); err != nil {
		resp.Err = fmt.Errorf("UpdateAccountInfo err: %s", err.Error())
		return
	}

	return
}
