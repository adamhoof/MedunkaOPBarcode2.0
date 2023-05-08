package file_parser

type FileParser interface {
	ToCSV(inputFileLocation string, outputFileLocation string, helperParserLocation string) (err error)
}
