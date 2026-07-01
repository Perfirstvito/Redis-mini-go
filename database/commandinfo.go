package database

import "strings"

type command struct {
	executor ExecFunc
	prepare  PreFunc
	undo     UndoFunc
	arity    int
}

var cmdTable = make(map[string]*command)

func registerCommand(name string, executor ExecFunc, prepare PreFunc, undo UndoFunc, arity int) {
	if prepare == nil {
		prepare = noPrepare
	}
	cmdTable[strings.ToLower(name)] = &command{
		executor: executor,
		prepare:  prepare,
		undo:     undo,
		arity:    arity,
	}
}

func noPrepare(args [][]byte) ([]string, []string) {
	return nil, nil
}
