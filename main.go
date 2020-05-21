package main

func main() {
	fcgiServer := NewFCGIServer(":9000", StandardStdoutHandler{})
	fcgiServer.Serve()
}