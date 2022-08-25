package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/pjrpc/pjrpc/v2/client"

	"github.com/rinatusmanovjsonrpc20/internal/pkg/cashstore"
	"github.com/rinatusmanovjsonrpc20/model/seamlessv2/rpcclient"
	"github.com/rinatusmanovjsonrpc20/model/types"
)

//go:embed swagger
var swagger embed.FS

func main() {
	http.Handle("/swagger/", http.FileServer(http.FS(swagger)))

	var (
		invoker       client.Invoker
		errNewInvoker error
	)

	if invoker, errNewInvoker = client.New("", nil); errNewInvoker != nil {
		panic(errNewInvoker)
	}

	serviceClient := rpcclient.NewSeamlessV2ServiceClient(invoker)
	rollbackTransactionStore := cashstore.NewCache()
	withdrawAndDepositStore := cashstore.NewCache()

	http.Handle("/GetBalance", getBalance(serviceClient))
	http.Handle("/RollbackTransaction", rollbackTransaction(serviceClient, rollbackTransactionStore))
	http.Handle("/WithdrawAndDeposit", withdrawAndDeposit(serviceClient, withdrawAndDepositStore))

	if errListenAndServe := http.ListenAndServe(":8086", nil); errListenAndServe != nil {
		panic(errListenAndServe)
	}
}

func errorAnswer(w http.ResponseWriter, statusCode int, str string) {
	errorRequest := types.ErrorData{ClientMessage: str}

	w.WriteHeader(statusCode)

	if errByteSlice, errErrorMarshal := json.Marshal(&errorRequest); errErrorMarshal == nil {
		_, _ = w.Write(errByteSlice)
	}
}

func getBalance(serviceClient rpcclient.SeamlessV2ServiceClient) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		defer func() {
			if err := recover(); err != nil {
				errorAnswer(w, http.StatusInternalServerError, fmt.Sprint(err))
			}
		}()
		if r.Method != http.MethodPost {
			errorAnswer(w, http.StatusMethodNotAllowed, "Method not allowed")

			return
		}

		byteSlice, errReadAll := io.ReadAll(r.Body)
		if errReadAll != nil {
			errorAnswer(w, http.StatusBadRequest, errReadAll.Error())

			return
		}
		defer r.Body.Close()

		var request types.GetBalanceRequest
		if errUnmarshal := json.Unmarshal(byteSlice, &request); errUnmarshal != nil {
			errorAnswer(w, http.StatusBadRequest, errUnmarshal.Error())

			return
		}

		response, errResponse := serviceClient.GetBalance(r.Context(), &request)
		if errResponse != nil {
			errorAnswer(w, http.StatusInternalServerError, errResponse.Error())

			return
		}

		answerByteSlice, errMarshal := json.Marshal(response)
		if errMarshal != nil {
			errorAnswer(w, http.StatusInternalServerError, errMarshal.Error())

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(answerByteSlice)
	})
}

func rollbackTransaction(serviceClient rpcclient.SeamlessV2ServiceClient, cache cashstore.Cache) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		defer func() {
			if err := recover(); err != nil {
				errorAnswer(w, http.StatusInternalServerError, fmt.Sprint(err))
			}
		}()
		if r.Method != http.MethodPost {
			errorAnswer(w, http.StatusInternalServerError, "Method not allowed")

			return
		}

		byteSlice, errReadAll := io.ReadAll(r.Body)
		if errReadAll != nil {
			errorAnswer(w, http.StatusBadRequest, errReadAll.Error())

			return
		}
		defer r.Body.Close()

		var request types.RollbackTransactionRequest
		if errUnmarshal := json.Unmarshal(byteSlice, &request); errUnmarshal != nil {
			errorAnswer(w, http.StatusBadRequest, errUnmarshal.Error())

			return
		}

		if cache.Check(request.TransactionRef) {
			errorAnswer(w, http.StatusBadRequest, "Transaction already exists")

			return
		}

		cache.Cache(request.TransactionRef)

		response, errResponse := serviceClient.RollbackTransaction(r.Context(), &request)
		if errResponse != nil {
			errorAnswer(w, http.StatusInternalServerError, errResponse.Error())

			return
		}

		answerByteSlice, errMarshal := json.Marshal(response)
		if errMarshal != nil {
			errorAnswer(w, http.StatusInternalServerError, errMarshal.Error())

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(answerByteSlice)
	})
}

func withdrawAndDeposit(serviceClient rpcclient.SeamlessV2ServiceClient, cache cashstore.Cache) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		defer func() {
			if err := recover(); err != nil {
				errorAnswer(w, http.StatusInternalServerError, fmt.Sprint(err))
			}
		}()

		if r.Method != http.MethodPost {
			errorAnswer(w, http.StatusInternalServerError, "Method not allowed")

			return
		}

		byteSlice, errReadAll := io.ReadAll(r.Body)
		if errReadAll != nil {
			errorAnswer(w, http.StatusBadRequest, errReadAll.Error())

			return
		}
		defer r.Body.Close()

		var request types.WithdrawAndDepositRequest
		if errUnmarshal := json.Unmarshal(byteSlice, &request); errUnmarshal != nil {
			errorAnswer(w, http.StatusBadRequest, errUnmarshal.Error())

			return
		}

		if cache.Check(request.TransactionRef) {
			errorAnswer(w, http.StatusBadRequest, "Transaction already exists")

			return
		}

		cache.Cache(request.TransactionRef)

		response, errResponse := serviceClient.WithdrawAndDeposit(r.Context(), &request)
		if errResponse != nil {
			errorAnswer(w, http.StatusInternalServerError, errResponse.Error())

			return
		}

		answerByteSlice, errMarshal := json.Marshal(response)
		if errMarshal != nil {
			errorAnswer(w, http.StatusInternalServerError, errMarshal.Error())

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(answerByteSlice)
	})
}
