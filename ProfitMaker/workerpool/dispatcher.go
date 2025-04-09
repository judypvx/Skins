package workerpool

import (
	"log"
	"sync"
	"profitmaker/analyzer"
	"profitmaker/buffer"
	"profitmaker/normalizer"
)

var (
	TaskQueue chan normalizer.NormalizedItem
	startOnce sync.Once
)

func StartWorkerPool(numWorkers int, queueSize int) {
	startOnce.Do(func() {
		TaskQueue = make(chan normalizer.NormalizedItem, queueSize)

		for i := 0; i < numWorkers; i++ {
			go worker(i)
		}

		log.Printf("🚀 WorkerPool запущен: %d воркеров, очередь %d\n", numWorkers, queueSize)
	})
}

func worker(id int) {
	for item := range TaskQueue {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("💥 Worker #%d упал: %v", id, r)
				}
			}()

			buffer.StartAnalysis(item.AssetID, item)
			analyzer.Analyze(item)
			buffer.Finish(item.AssetID)
		}()
	}
}
