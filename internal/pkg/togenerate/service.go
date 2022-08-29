package togenerate

import "github.com/rinatusmanov/jsonrpc20/internal/pkg/types"

// SeamlessV2Service интерфейс для обработки запросов к сервису seamlessv2
//
//nolint:lll
//go:generate genpjrpc -search.name=SeamlessV2Service -print.place.path_swagger_file=../../../swagger/generated.json -print.content.swagger_data_path=./swagger_data.json -print.place.path_client=../seamlessv2_client/generated -print.place.path_server=../seamlessv2/generated //nolint:lll
type SeamlessV2Service interface {
	// You can set your own 'error'.'data' type to add some additional error context.
	setErrorData(types.ErrorData)

	//genpjrpc:params method_name=getBalance
	GetBalance(request types.GetBalanceRequest) types.GetBalanceResponse
	//genpjrpc:params method_name=withdrawAndDeposit
	RollbackTransaction(request types.RollbackTransactionRequest) types.RollbackTransactionResponse
	//genpjrpc:params method_name=rollbackTransaction
	WithdrawAndDeposit(request types.WithdrawAndDepositRequest) types.WithdrawAndDepositResponse
}
