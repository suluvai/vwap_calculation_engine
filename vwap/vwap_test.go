package vwap_test

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/suluvai/vwap_calculation_engine/vwap"
)

func TestList(t *testing.T) {
	t.Parallel()

	list, err := vwap.NewDataSet([]vwap.DataPoint{}, 1)
	require.NoError(t, err)

	first := vwap.DataPoint{Price: decimal.NewFromInt(1), Volume: decimal.NewFromInt(1)}

	second := vwap.DataPoint{Price: decimal.NewFromInt(2), Volume: decimal.NewFromInt(2)}

	third := vwap.DataPoint{Price: decimal.NewFromInt(3), Volume: decimal.NewFromInt(3)}

	list.Update(first)
	require.Equal(t, 1, list.Len())
	require.Equal(t, first, list.DataPointsQueue[0])

	list.Update(second)
	require.Equal(t, 1, list.Len())
	require.Equal(t, second, list.DataPointsQueue[0])

	list.Update(third)
	require.Equal(t, 1, list.Len())
	require.Equal(t, third, list.DataPointsQueue[0])
}

func TestListConcurrentPush(t *testing.T) {
	t.Parallel()

	list, err := vwap.NewDataSet([]vwap.DataPoint{}, 2)
	require.NoError(t, err)

	first := vwap.DataPoint{Price: decimal.NewFromInt(1), Volume: decimal.NewFromInt(1)}

	second := vwap.DataPoint{Price: decimal.NewFromInt(2), Volume: decimal.NewFromInt(2)}

	third := vwap.DataPoint{Price: decimal.NewFromInt(3), Volume: decimal.NewFromInt(3)}

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		list.Update(first)
		wg.Done()
	}()

	go func() {
		list.Update(second)
		wg.Done()
	}()

	go func() {
		list.Update(third)
		wg.Done()
	}()

	wg.Wait()

	require.Len(t, list.DataPointsQueue, 2)
}

func TestVWAP(t *testing.T) {
	t.Parallel()

	listOfTestScenarios := []struct {
		Name            string
		DataPointsQueue []vwap.DataPoint
		ExpectedVwap    map[string]decimal.Decimal
		MaxSize         uint
	}{
		{
			Name:            "EmptyDataPointsQueue",
			DataPointsQueue: []vwap.DataPoint{},
			ExpectedVwap: map[string]decimal.Decimal{
				"BTC-USD": decimal.Zero,
				"ETH-USD": decimal.Zero,
			},
		},
		{
			Name: "FullDataPointsQueue1",
			DataPointsQueue: []vwap.DataPoint{
				{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10), ProductId: "BTC-USD"},
				{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10), ProductId: "BTC-USD"},
				{Price: decimal.NewFromInt(31), Volume: decimal.NewFromInt(30), ProductId: "ETH-USD"},
				{Price: decimal.NewFromInt(21), Volume: decimal.NewFromInt(20), ProductId: "BTC-USD"},
				{Price: decimal.NewFromInt(41), Volume: decimal.NewFromInt(33), ProductId: "ETH-USD"},
			},
			MaxSize: 4,
			ExpectedVwap: map[string]decimal.Decimal{
				"BTC-USD": decimal.RequireFromString("17.3333333333333333"),
				"ETH-USD": decimal.RequireFromString("36.2380952380952381"),
			},
		},
		{
			Name: "FullDataPointsQueue2",
			DataPointsQueue: []vwap.DataPoint{
				{Price: decimal.NewFromInt(10), Volume: decimal.RequireFromString("10.1"), ProductId: "BTC-USD"},
				{Price: decimal.NewFromInt(10), Volume: decimal.RequireFromString("10.1"), ProductId: "BTC-USD"},
			},
			ExpectedVwap: map[string]decimal.Decimal{
				"BTC-USD": decimal.RequireFromString("10"),
			},
			MaxSize: 4,
		},
	}

	for _, eachTest := range listOfTestScenarios {
		t.Run(eachTest.Name, func(t *testing.T) {
			t.Parallel()

			list, err := vwap.NewDataSet([]vwap.DataPoint{}, eachTest.MaxSize)
			require.NoError(t, err)

			for _, d := range eachTest.DataPointsQueue {
				list.Update(d)
			}

			for k := range eachTest.ExpectedVwap {
				require.Equal(t, eachTest.ExpectedVwap[k].String(), list.VWAP[k].String())
			}
		})
	}
}
