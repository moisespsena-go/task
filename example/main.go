package main

import "github.com/moisespsena-go/task"

func main() {
	task.NewRunner(newTask("a"), newTask("b"), newTask("c")).MustSigRun()
	println("main done!")
}

func newTask(name string) task.Task {
	name += ": "
	c := make(chan interface{})
	return task.NewTask(func() (err error) {
		println(name + "start")
		<-c
		println(name + "start done")
		return nil
	}, func() {
		println(name + "closing")
		close(c)
		println(name + "closing done")
	})
}
