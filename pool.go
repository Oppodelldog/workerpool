package workerpool

type (
	// WorkerPool is capable of doing work
	WorkerPool interface {
		Work(workLoad interface{})
	}

	workerPool struct {
		createWorker   func() Worker
		workers        []workerInfo
		responseOutput chan interface{}
	}

	workerInfo struct {
		worker          Worker
		workerID        int
		responseChannel chan interface{}
	}

	// Worker is able to Work and to inform about its status
	Worker interface {
		Work(control WorkControl)
		IsWorking() bool
	}

	// WorkControl defines 2 Information for Workers.
	// * WorkLoad - information about the work to do
	// * ResponseChannel - channel to pass back information
	WorkControl struct {
		WorkLoad          interface{}
		WorkResultChannel chan interface{}
	}
)

// New creates a new worker pool.
// createWorker is used to create new workers when Work is called.
// processResponseFunc is called when a worker finishes.
func New(workerFactory func() Worker, processResponseFunc func(response interface{})) WorkerPool {

	newWorkerPool := workerPool{
		createWorker:   workerFactory,
		workers:        []workerInfo{},
		responseOutput: make(chan interface{}),
	}

	go newWorkerPool.responseProcessing(processResponseFunc)

	return &newWorkerPool
}

// NewFireAndForget creates a worker pool that simply executes work, but ignores responses from workers.
// createWorker is used to create new workers when Work is called.
func NewFireAndForget(workerFactory func() Worker) WorkerPool {

	var ignoreResponseFunc = func(response interface{}) {}

	return New(workerFactory, ignoreResponseFunc)
}

func (pool *workerPool) Work(workLoad interface{}) {
	freeWorker := pool.acquireWorker()

	freeWorker.worker.Work(WorkControl{
		WorkLoad:          workLoad,
		WorkResultChannel: freeWorker.responseChannel,
	})
}

func (pool *workerPool) getFreeWorker() *workerInfo {
	for _, workerInfo := range pool.workers {

		if !workerInfo.worker.IsWorking() {
			return &workerInfo
		}
	}

	return nil
}

func (pool *workerPool) createNewWorker() *workerInfo {
	newWorkerInfo := workerInfo{
		worker:          pool.createWorker(),
		workerID:        len(pool.workers) + 1,
		responseChannel: make(chan interface{}, 1),
	}

	pool.workers = append(pool.workers, newWorkerInfo)

	return &newWorkerInfo
}

func (pool *workerPool) acquireWorker() *workerInfo {
	freeWorker := pool.getFreeWorker()
	if freeWorker == nil {
		freeWorker = pool.createNewWorker()
	}

	return freeWorker
}

func (pool *workerPool) responseProcessing(responseProcessingFunc func(response interface{})) {
	for {
		for _, worker := range pool.workers {
			select {
			case response := <-worker.responseChannel:
				responseProcessingFunc(response)
			default:
			}
		}
	}
}
