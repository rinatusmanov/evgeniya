package seamlessv2

import "github.com/rinatusmanovjsonrpc20/model/types"

// SeamlessV2Service you can describe your service with methods.
//
//nolint:lll
//go:generate genpjrpc -search.name=SeamlessV2Service -print.place.path_swagger_file=../../cmd/swagger/generated.json -print.content.swagger_data_path=./swagger_data.json
type SeamlessV2Service interface { //nolint:revive
	// You can set your own 'error'.'data' type to add some additional error context.
	setErrorData(types.ErrorData)

	GetBalance(request types.GetBalanceRequest) types.GetBalanceResponse
	RollbackTransaction(request types.RollbackTransactionRequest) types.RollbackTransactionResponse
	WithdrawAndDeposit(request types.WithdrawAndDepositRequest) types.WithdrawAndDepositResponse
}
