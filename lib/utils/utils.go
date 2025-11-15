package utils

func ToCmdLine(cmd ...string) [][]byte {
	args := make([][]byte, len(cmd))
	for i, c := range cmd {
		args[i] = []byte(c)
	}
	return args
}

func ToCmdLine2(commandName string, args ...[]byte) [][]byte {
	result := make([][]byte, len(args)+1)
	result[0] = []byte(commandName)
	for i, c := range args {
		result[i+1] = c
	}
	return result
}
