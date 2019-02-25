package ultron

type (
	resultPipeline chan *Result
	reportPipeline chan Report

	statusPipeline chan Status
	countPipeline  chan uint8
)

var (
	localResultPipeline  resultPipeline
	slaveResultPipeline  resultPipeline
	masterResultPipeline resultPipeline

	// LocalResultPipelineBufferSize .
	LocalResultPipelineBufferSize = 1000
	// SlaveResultPipelineBufferSize .
	SlaveResultPipelineBufferSize = 1000
	// MasterResultPipelineBufferSize .
	MasterResultPipelineBufferSize = 2000

	localReportPipeline  reportPipeline
	slaveReportPipeline  reportPipeline
	masterReportPipeline reportPipeline

	// LocalReportPipelineBufferSize .
	LocalReportPipelineBufferSize = 10
	// SlaveReportPipelineBufferSize .
	SlaveReportPipelineBufferSize = 10
	// MasterReportPipelineBufferSize .
	MasterReportPipelineBufferSize = 20

	StageRunnerStatusPipeline = newStatusPipline()
	//CounterPipeline = newCountPipeline(CounterPiplineBuffer)
	//CounterPipline的Buffer大小
	//CounterPiplineBuffer = 1000
)

func newResultPipeline(b int) resultPipeline {
	return make(chan *Result, b)
}

func newReportPipeline(b int) reportPipeline {
	return make(chan Report, b)
}

func newStatusPipline() statusPipeline {
	return make(chan Status)
}

//func newCountPipeline(buffer int) countPipeline {
//	return make(chan uint8, buffer)
//}
