package printer

import . "github.com/jdugan1024/jdgo/types"

func PrintStr(ast MalType) string {
	return ast.Print()
}
