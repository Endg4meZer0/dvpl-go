package dvpl

type DVPLConverterError string

func (e DVPLConverterError) Error() string {
	return string(e)
}
