package service

import (
	"coding-challenge/model"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
)

const port = ":8800"

type handler struct {
	TransactionRecords  []TransactionRecord
	PointBalanceByPayer map[string]int
}

func StartService() {
	h := NewHandler()
	r := mux.NewRouter()

	r.HandleFunc("/transactions/add", h.AddTransaction)
	r.HandleFunc("/points/spend", h.SpendPoints)
	r.HandleFunc("/points/list", h.ListPayerBalances)

	http.ListenAndServe(port, r)
}

func NewHandler(transactions ...TransactionRecord) handler {
	h := handler{
		TransactionRecords:  make([]TransactionRecord, 0),
		PointBalanceByPayer: make(map[string]int),
	}

	if transactions != nil {
		h.TransactionRecords = transactions
		h.processTransactions()
	}

	return h
}

func (h *handler) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var addTransactionRequest model.AddTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&addTransactionRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction := TransactionRecord{
		Payer:     addTransactionRequest.Payer,
		Points:    addTransactionRequest.Points,
		Timestamp: addTransactionRequest.TimeStamp,
	}

	pointBalance := h.PointBalanceByPayer[transaction.Payer]

	if pointBalance+transaction.Points < 0 {
		http.Error(w, "payer has insufficient points for transaction", http.StatusUnprocessableEntity)
		return
	} else {
		h.PointBalanceByPayer[transaction.Payer] = pointBalance + transaction.Points
	}

	h.TransactionRecords = append(h.TransactionRecords, transaction)

	w.WriteHeader(http.StatusCreated)
}

func (h *handler) SpendPoints(w http.ResponseWriter, r *http.Request) {
	var spendPointsRequest model.SpendPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&spendPointsRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pointsLeftToSpend := spendPointsRequest.Points
	spentPointsMap := map[string]int{}
	spentPoints := []model.PayerPointRecord{}

	if pointsLeftToSpend <= 0 {
		http.Error(w, "point value to debit must be greater than zero", http.StatusUnprocessableEntity)
		return
	}

	if !h.hasSufficientPointBalance(pointsLeftToSpend) {
		http.Error(w, "insufficient point balance for requested point expense", http.StatusUnprocessableEntity)
		return
	}

	sortedRecords := sortTransactionRecordsByTimestamp(h.TransactionRecords)

	for i := range sortedRecords {
		payer := sortedRecords[i].Payer
		transactionPoints := sortedRecords[i].Points
		payerPoints := h.PointBalanceByPayer[payer]

		if payerPoints == 0 || transactionPoints == 0 {
			continue
		}

		if transactionPoints >= pointsLeftToSpend {
			sortedRecords[i].Points -= pointsLeftToSpend
			spentPointsMap[payer] -= pointsLeftToSpend
			pointsLeftToSpend = 0
		} else if transactionPoints < pointsLeftToSpend {
			sortedRecords[i].Points = 0
			spentPointsMap[payer] -= transactionPoints
			pointsLeftToSpend -= transactionPoints
		}

		if pointsLeftToSpend == 0 {
			break
		}
	}

	if pointsLeftToSpend == 0 {
		h.TransactionRecords = sortedRecords
	} else {
		http.Error(w, "there was an error debiting points", http.StatusUnprocessableEntity)
		return
	}

	h.pruneZeroTransactions()

	for k, v := range spentPointsMap {
		spentPoints = append(spentPoints, model.PayerPointRecord{
			Payer:  k,
			Points: v,
		})

		h.PointBalanceByPayer[k] += v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(spentPoints)
}

func (h *handler) ListPayerBalances(w http.ResponseWriter, r *http.Request) {
	response := model.ListPayerBalancesResponse{PayerBalances: []model.PayerPointRecord{}}
	for k, v := range h.PointBalanceByPayer {
		response.PayerBalances = append(response.PayerBalances, model.PayerPointRecord{Payer: k, Points: v})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type TransactionRecord struct {
	Payer     string    `json:"payer"`
	Points    int       `json:"points"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *handler) pruneZeroTransactions() {
	records := []TransactionRecord{}

	for v := range h.TransactionRecords {
		if h.TransactionRecords[v].Points != 0 {
			records = append(records, h.TransactionRecords[v])
		}
	}

	h.TransactionRecords = records
}

func (h *handler) hasSufficientPointBalance(points int) bool {
	pointBalance := 0
	for _, points := range h.PointBalanceByPayer {
		pointBalance += points
	}

	return pointBalance >= points
}

func (h *handler) processTransactions() {
	for _, r := range h.TransactionRecords {
		h.PointBalanceByPayer[r.Payer] += r.Points
	}
}

func sortTransactionRecordsByTimestamp(slice []TransactionRecord) []TransactionRecord {
	sortedRecords := []TransactionRecord{}
	sortedRecords = append(sortedRecords, slice...)
	sort.Slice(sortedRecords, func(i, j int) bool {
		return sortedRecords[i].Timestamp.Before(sortedRecords[j].Timestamp)
	})

	return sortedRecords
}
