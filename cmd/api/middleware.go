package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
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

func (app *application) rateLimit(next http.Handler) http.Handler {

	// If rate limiting is not enbaled, return the next handler in the chain with
	// with no further action
	if !app.config.limiter.enabled {
		return next
	}

	// Define a client struct to hold the rate limiter and last seen time for each client.

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutes and map to hold the clients' IP addresses and rate limiters.
	var (
		mu sync.Mutex
		// Update the map so the values are pointers to a client struct
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Importantlym unlock the mutex when the clenup is complete
			mu.Unlock()
		}
	}()

	// Initialize a new rate limiter which allows an average of 2 requests per second,
	// with a maximum of 4 requests in the single 'brust'.
	//limiter := rate.NewLimiter(2, 4)

	// The function we are returning is a closure, which 'closes over' the limiter variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the realip.FromRequest() function to get the client's IP address.
		ip := realip.FromRequest(r)

		// Lock the mutex to prevent this code from being executed concurrently.
		mu.Lock()

		// Check to see if the IP address already exists in the map. If it doesn't, then
		// initialize a new raste limiter and add the IP address and limiter to the map.
		if _, found := clients[ip]; !found {
			// Create and add a new client struct to the map if it does not already exist.
			clients[ip] = &client{
				// Use the request-per-second and burst values from the config struct
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}

		// Update the last seen time for the client.
		clients[ip].lastSeen = time.Now()

		// Call the Allow() method on the rate limiter for the current IP address. If
		// the request is not allowed, unlock the mutex and send a 429 Too Many requests
		// response, kust like before
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Ver importantly, unlock the mutex before calling the next handler in the
		// chain. Notice that we DO NOT USE defer to unclock the mutex, as that would mean
		// that the mutex is not unclocked until all the handlers downsream of this middleware
		// have also returned.
		mu.Unlock()

		// Lock the mutes to prevent this codefrom being executed concura
		// Call limiter.Allow() to see if the request is permitted, and if it is not,
		// then we call the rateLimitExceededResponse() helper to return a 429 Too Many
		// Requests response (we will create this helper in a minute).
		//if !limiter.Allow() {
		//	app.rateLimitExceededResponse(w, r)
		//	return
		//}

		next.ServeHTTP(w, r)
	})
}
