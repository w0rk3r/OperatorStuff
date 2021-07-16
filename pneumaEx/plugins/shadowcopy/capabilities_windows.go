package main

var (
	ModuleName = "shadowcopy"
	Functions  = map[string]func(args []string) ([]byte, int){
		"vss": vss,
	}
	ExecFunctions = map[string]func(args string){}
)

func vss(args []string) ([]byte, int) {
	message := CreateShadowCopy(args[0], args[1])
	return []byte(message), 0
}
