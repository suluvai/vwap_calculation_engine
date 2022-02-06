package vwap

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
)

const defaultMaxSize = 200

// DataPoint represents a single data point from coinbase.
type DataPoint struct {
	Price     decimal.Decimal
	Volume    decimal.Decimal
	ProductId string
}

// DataSet represents a set of DataPoints.
type DataSet struct {
	mu                  sync.Mutex
	DataPointsQueue     []DataPoint
	TotalVolumeWeighted map[string]decimal.Decimal
	TotalVolume         map[string]decimal.Decimal
	VWAP                map[string]decimal.Decimal
	WindowSize          uint
}

// NewDataSet creates a new queue.
func NewDataSet(dataPoint []DataPoint, maxSize uint) (DataSet, error) {
	if maxSize == 0 {
		maxSize = defaultMaxSize
	}

	if len(dataPoint) > int(maxSize) {
		return DataSet{}, errors.New("initial datapoints exceeds maxSize")
	}

	return DataSet{
		DataPointsQueue:     dataPoint,
		WindowSize:          maxSize,
		TotalVolumeWeighted: make(map[string]decimal.Decimal),
		TotalVolume:         make(map[string]decimal.Decimal),
		VWAP:                make(map[string]decimal.Decimal),
	}, nil
}

// Len returns the length of the Queue.
func (ds *DataSet) Len() int {
	return len(ds.DataPointsQueue)
}

// Update pushes an element onto the queue, drops the first one when MaxSize is reached.
func (ds *DataSet) Update(d DataPoint) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if len(ds.DataPointsQueue) == int(ds.WindowSize) {
		d := ds.DataPointsQueue[0]
		ds.DataPointsQueue = ds.DataPointsQueue[1:]

		// Substract the datapoint values from the VWAP calculation.
		ds.TotalVolumeWeighted[d.ProductId] = ds.TotalVolumeWeighted[d.ProductId].Sub(d.Price.Mul(d.Volume))
		ds.TotalVolume[d.ProductId] = ds.TotalVolume[d.ProductId].Sub(d.Volume)
		if !ds.TotalVolume[d.ProductId].IsZero() {
			ds.VWAP[d.ProductId] = ds.TotalVolumeWeighted[d.ProductId].Div(ds.TotalVolume[d.ProductId])
		}
	}

	if _, ok := ds.VWAP[d.ProductId]; ok {
		ds.TotalVolumeWeighted[d.ProductId] = ds.TotalVolumeWeighted[d.ProductId].Add(d.Price.Mul(d.Volume))
		ds.TotalVolume[d.ProductId] = ds.TotalVolume[d.ProductId].Add(d.Volume)
		ds.VWAP[d.ProductId] = ds.TotalVolumeWeighted[d.ProductId].Div(ds.TotalVolume[d.ProductId])
	} else {
		initialVW := d.Price.Mul(d.Volume)

		ds.TotalVolumeWeighted[d.ProductId] = initialVW
		ds.TotalVolume[d.ProductId] = d.Volume
		ds.VWAP[d.ProductId] = initialVW.Div(d.Volume)
	}

	ds.DataPointsQueue = append(ds.DataPointsQueue, d)
}
