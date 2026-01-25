package main

func main() {
	println("Content-Type: text/plain; charset=utf-8")
	println()
	println("admin")

	// if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("content-type", "text/plain")
	// 	w.Write([]byte("admin"))
	// })); err != nil {
	// 	fmt.Fprintf(os.Stderr, "serve error: %v", err)
	// 	os.Exit(1)
	// }
}
