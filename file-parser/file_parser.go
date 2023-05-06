package file_parser

type FileParser interface {
	ToCSV(inputFileLocation string, outputFileLocation string) (err error)
}
