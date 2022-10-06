package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_pruneZeroTransactions(t *testing.T) {
	t.Run("verify 0 len works properly", func(t *testing.T) {
		transactions := []TransactionRecord{}

		handler := NewHandler(transactions...)

		handler.pruneZeroTransactions()

		assert.Equal(t, 0, len(handler.TransactionRecords))
	})

	t.Run("verify zero point transactions get removed but not negative point records", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Points: 0},
			{Points: 100},
			{Points: -100},
		}

		handler := NewHandler(transactions...)

		handler.pruneZeroTransactions()

		assert.Equal(t, 2, len(handler.TransactionRecords))
		for _, i := range handler.TransactionRecords {
			assert.NotEqual(t, 0, i.Points)
		}
	})

	t.Run("verify empty list if all transactions zero", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Points: 0},
			{Points: 0},
			{Points: 0},
		}

		handler := NewHandler(transactions...)

		handler.pruneZeroTransactions()

		assert.Equal(t, 0, len(handler.TransactionRecords))
	})
}

func Test_hasSufficientPointBalance(t *testing.T) {
	t.Run("returns true when exactly sufficient", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Payer: "DANNON", Points: 1000},
		}

		handler := NewHandler(transactions...)

		result := handler.hasSufficientPointBalance(1000)

		assert.True(t, result)
	})
	t.Run("returns true when sufficient with surplus", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Payer: "DANNON", Points: 10000},
		}

		handler := NewHandler(transactions...)

		result := handler.hasSufficientPointBalance(1000)

		assert.True(t, result)
	})
	t.Run("returns true when points must come from different payers", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Payer: "DANNON", Points: 500},
			{Payer: "YOPLAIT", Points: 500},
		}

		handler := NewHandler(transactions...)

		result := handler.hasSufficientPointBalance(1000)

		assert.True(t, result)
	})
	t.Run("returns false when zero", func(t *testing.T) {
		transactions := []TransactionRecord{}

		handler := NewHandler(transactions...)

		result := handler.hasSufficientPointBalance(1000)

		assert.False(t, result)
	})
	t.Run("returns false when insufficient point balance", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Payer: "DANNON", Points: 500},
		}

		handler := NewHandler(transactions...)

		result := handler.hasSufficientPointBalance(1000)

		assert.False(t, result)
	})
}

func Test_sortTransactionRecordsByTimestamp(t *testing.T) {
	t.Run("sorts by timestamp properly", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Timestamp: time.Now()},
			{Timestamp: time.Now().Add(time.Hour * 24)},
			{Timestamp: time.Now().Add(-time.Hour * 24)},
		}

		result := sortTransactionRecordsByTimestamp(transactions)

		assert.LessOrEqual(t, result[0].Timestamp, result[1].Timestamp)
		assert.LessOrEqual(t, result[1].Timestamp, result[2].Timestamp)
	})

	t.Run("does not error when empty array provided", func(t *testing.T) {
		transactions := []TransactionRecord{}

		result := sortTransactionRecordsByTimestamp(transactions)
		assert.Equal(t, 0, len(result))
	})

	t.Run("does not error on array of 1", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Timestamp: time.Now()},
		}

		result := sortTransactionRecordsByTimestamp(transactions)
		assert.Equal(t, 1, len(result))
	})
}

func Test_processTransactions(t *testing.T) {
	t.Run("verify process adds transaction to transaction records", func(t *testing.T) {
		transactions := []TransactionRecord{
			{Payer: "DANNON", Points: 1000},
		}

		handler := NewHandler()
		handler.TransactionRecords = transactions
		handler.processTransactions()

		assert.Equal(t, 1000, handler.TransactionRecords[0].Points)
		assert.Equal(t, 1000, handler.PointBalanceByPayer[handler.TransactionRecords[0].Payer])
	})
}
