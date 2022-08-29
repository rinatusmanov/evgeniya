package seamlessv2

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/rinatusmanov/jsonrpc20/internal/pkg/types"
	"github.com/rinatusmanov/jsonrpc20/internal/repo"
)

type RPCService struct {
	db     *sqlx.DB
	logger *zap.Logger
}

var (
	ErrNoFreeCurrency       = errors.New("no free currency")
	ErrConflictOfCurrencies = errors.New("conflict of currencies")
)

// GetBalance проигнорировал очень много полей так как вообще не понимаю их сути.
func (r *RPCService) GetBalance(
	ctx context.Context,
	in *types.GetBalanceRequest,
) (*types.GetBalanceResponse, error) {
	tx, _ := r.db.BeginTxx(ctx, nil)
	defer tx.Commit() //nolint:errcheck // intentional

	span := opentracing.SpanFromContext(ctx)
	span.SetTag("method", "GetBalance")

	user, errFindUserByName := repo.FindUserByName(ctx, tx, in.PlayerName)
	if errFindUserByName != nil {
		return r.newUser(ctx, in, tx)
	}

	currencyID, errGetCurrencyID := repo.GetCurrencyID(ctx, tx, user.ID)
	if errGetCurrencyID != nil {
		return nil, errGetCurrencyID //nolint:wrapcheck // intentional
	}

	currency, errGetCurrencyByID := repo.GetCurrencyByID(ctx, tx, currencyID)
	if errGetCurrencyByID != nil {
		return nil, errGetCurrencyByID //nolint:wrapcheck // intentional
	}

	if currency.ID != currencyID {
		return nil, ErrConflictOfCurrencies
	}

	if in.BonusID != "" {
		if errWriteBonus := repo.WriteBonus(ctx, tx, in.BonusID); errWriteBonus != nil {
			return nil, errWriteBonus //nolint:wrapcheck // intentional
		}
	}

	balance, errGetDepositByUserID := repo.GetDepositByUserID(ctx, tx, user.ID)
	if errGetDepositByUserID != nil {
		return nil, errGetDepositByUserID //nolint:wrapcheck // intentional
	}

	return &types.GetBalanceResponse{Balance: balance, FreeRoundsLeft: 0}, nil
}

func (r *RPCService) newUser(
	ctx context.Context,
	in *types.GetBalanceRequest,
	tx *sqlx.Tx,
) (*types.GetBalanceResponse, error) {
	var (
		errNewUser error
		user       *repo.User
	)

	if user, errNewUser = repo.NewUser(ctx, tx, in.PlayerName); errNewUser != nil {
		return nil, errNewUser //nolint:wrapcheck // intentional
	}

	currency, errGetCurrencyByCode := repo.GetCurrencyByCode(ctx, tx, in.Currency)
	if errGetCurrencyByCode != nil {
		return nil, errGetCurrencyByCode //nolint:wrapcheck // intentional
	}

	const balance = 10000
	if _, errNewPayment := repo.NewPayment(ctx, tx, user.ID, currency.ID, 0, balance, "init"); errNewPayment != nil {
		return nil, errNewPayment //nolint:wrapcheck // intentional
	}

	return &types.GetBalanceResponse{
		Balance:        balance,
		FreeRoundsLeft: 0,
	}, nil
}

func (r *RPCService) RollbackTransaction(
	ctx context.Context,
	in *types.RollbackTransactionRequest,
) (*types.RollbackTransactionResponse, error) {
	tx, _ := r.db.BeginTxx(ctx, nil)
	defer tx.Commit() //nolint:errcheck // intentional

	return &types.RollbackTransactionResponse{}, repo.RollbackPayment( //nolint:wrapcheck // intentional
		ctx,
		tx,
		in.TransactionRef,
	)
}

func (r *RPCService) WithdrawAndDeposit(
	ctx context.Context,
	in *types.WithdrawAndDepositRequest,
) (*types.WithdrawAndDepositResponse, error) {
	tx, _ := r.db.BeginTxx(ctx, nil)
	defer tx.Commit() //nolint:errcheck // intentional

	if in.BonusID != "" {
		if errUseBonus := repo.UseBonus(ctx, tx, in.BonusID); errUseBonus != nil {
			return nil, errUseBonus //nolint:wrapcheck // intentional
		}
	}

	user, errFindUserByName := repo.FindUserByName(ctx, tx, in.PlayerName)
	if errFindUserByName != nil {
		return nil, errFindUserByName //nolint:wrapcheck // intentional
	}

	deposit, errGetDepositByUserID := repo.GetDepositByUserID(ctx, tx, user.ID)
	if errGetDepositByUserID != nil {
		return nil, errGetDepositByUserID //nolint:wrapcheck // intentional
	}

	newBalance := deposit + in.Deposit - in.Withdraw
	if newBalance < 0 {
		return nil, ErrNoFreeCurrency
	}

	if errCheckUniqueTransactionRef := repo.CheckUniqueTransactionRef(
		ctx,
		tx,
		in.TransactionRef,
	); errCheckUniqueTransactionRef != nil {
		return nil, errCheckUniqueTransactionRef //nolint:wrapcheck // intentional
	}

	_, errNewPayment := repo.NewPayment(ctx, tx, user.ID, 1, deposit, in.Withdraw, in.TransactionRef)
	if errNewPayment != nil {
		return nil, errNewPayment //nolint:wrapcheck // intentional
	}

	return &types.WithdrawAndDepositResponse{
		NewBalance:     newBalance,
		TransactionID:  in.TransactionRef,
		FreeRoundsLeft: 0,
	}, nil
}

func NewRPCService(db *sqlx.DB, logger *zap.Logger) *RPCService {
	return &RPCService{
		db:     db,
		logger: logger,
	}
}
