package main

import (
    "profitmaker/buffer"
    "profitmaker/config"
    "profitmaker/workerpool"
)

func main() {
    config.LoadConfig()
    buffer.StartCleaner()

    workerpool.StartWorkerPool(20, 2000)

    keepConnectionAlive()
}
