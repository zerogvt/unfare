# unfare
Calculation of fare estimates based on geographic time series data describing drives.

# input structure
The input will be a long CSV file describing drives.
Each line is of the form `id_ride, lat, lng, timestamp` and the file will be well-formed in the sense of no data multiplexing or time incosistencies.

# challenges
1. weed out outlier data. These are tuples that are probably due to faulty equipement and are skewed in relation to their previous data points.
2. The input might be very long so we need to come up with a concurrent solution that scales well with input size.

# architecture
We start reading through the input serially. This is done by the main goroutine. As we read in line by line (or tuple by tuple if you prefer) we fill in a temp bucket representing the current drive id. As soon as we see the next drive id we know that the bucket is full and we have all data for the previous drive id.

We then start a worker goroutine to process that drive concurrently and we go on with filling the next bucket. So, essentially we have a driver goroutine (the main goroutine) assigning work to worker goroutines. As the worker goroutines are usually a relativelly short computation we can expect their number to remain stable regardless of how long the driver goroutine (i.e. the input) might be. Which is, we expect this solution to scale well.

A merger goroutine is awaiting the results of the worker routines and whenever it gets a new result it appends it to the results file. Again this is a stable routine both in mem usage and in computation time (O(1) to read a result and append it to file) so we expect this too to scale well with input size.

Waitgroups are used to guard against premature program termination.

# How to build
`make build`

# How to run tests
`make test`
(Runs unit tests first, end to end test afterwards)

# Prerequisites
POSIX system to run e2e tests (tehy use bash and standard POSIX tooling such as `sort`)

# Known issues
Atm we do not guard against the case of the first geo point of a drive being an outlier. We're out of time but we think that for a first iteration its probably an acceptable low risk (I guess every ride starts from a stationary position so the chances of the starting point being a faulty one are minimal even with dodgy GPS equipment).
