package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Create a defered function which will always run in the event of panic

		defer func() {
			// use the built-in recover() function to check if panic occured.
			// If a panic did happen, revover() will return the panic value.
			// If a panic did not happen, it will return nil.
			pv := recover()
			if pv != nil {
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
