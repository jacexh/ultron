package ultron

type (
	resultPipeline chan *Result
	reportPipeline chan Report
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
)

func newResultPipeline(b int) resultPipeline {
	return make(chan *Result, b)
}

func newReportPipeline(b int) reportPipeline {
	return make(chan Report, b)
}
