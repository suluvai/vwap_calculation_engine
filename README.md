# VWAP Calculation Engine

A realtime VWAP (volume-weighted average price) calculation engine. Uses coinbase websocket feed to stream in trade executions and update the VWAP for each trading pair as updates become available.

## Problem Specification

- Retrieve a data feed from the coinbase websocket and subscribe to the matches channel. Pull data for
the following three trading pairs:
    - BTC-USD
    - ETH-USD
    - ETH-BTC
- Calculate the VWAP per trading pair using a sliding window of 200 data points. Meaning, when a new
data point arrives through the websocket feed the oldest data point will fall off and the new one will be
added such that no more than 200 data points are included in the calculation.
    - The first 200 updates will have less than 200 data points included. That’s fine for this project.
- Stream the resulting VWAP values on each websocket update.
    - Print to stdout or file is ok. Usually you would send them off through a message broker but a
simple print is perfect for this project.

## Design

- conf
    - Uses the configurable settings from `conf/app.config` file. This file is parsed and loaded info a usable struct for use by application.

- vwap
    - This package implements vwap calculation algorithm. It maintains the vwap for the specified window size.

- websocket
    - The websocket module is responsible to establish websocket connection with coinbase exchange and subscribes for data feed. It implements the client interface defined in the websocket package. Also, allows to mock interfaces to test in testing. Have not used mocking in this assigment though.

## Development commands (run from repo root)

Run tests in all subdirectories.

    go test ./...

Build.

    go build

Run.

    .\vwap_calculation_engine.exe

Format code consistently.

    go fmt ./...