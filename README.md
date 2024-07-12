# Go Load Balancer

## Overview

This project implements a simple round-robin load balancer in Go. It distributes incoming HTTP requests across multiple backend servers, handles server health checks, and simulates server downtime for testing purposes.

## Features

- Round-robin load balancing
- Dynamic health checking of backend servers
- Graceful handling of server downtime
- Simulation of random server outages for testing
- Concurrent operation using goroutines

## Prerequisites

- Go 1.15 or higher

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/go-load-balancer.git
   cd go-load-balancer
   ```

2. Build the project:
   ```
   go build -o load-balancer
   ```

## Usage

1. Start the load balancer:
   ```
   ./load-balancer
   ```

2. The load balancer will start on port 8080, and four backend servers will be started on ports 8081, 8082, 8083, and 8084.

3. Send requests to the load balancer:
   ```
   curl http://localhost:8080
   ```

4. Observe the console output to see how requests are distributed and how the system handles simulated server downtime.

## Configuration

The main configuration is done in the `main()` function:

- Backend servers are defined in the `servers` slice.
- Health check interval is set to 5 seconds.
- Server downtime simulation occurs every 60 seconds and lasts for 60 seconds.

You can modify these values to suit your testing needs.

## Project Structure

- `main.go`: Contains the entire implementation of the load balancer.
    - `Server` struct: Represents a backend server.
    - `RoundRobinServerPool` struct: Manages the pool of servers and implements the load balancing logic.
    - `LoadBalancer` function: Handles incoming requests and forwards them to the next available server.
    - `simulateServerDowntime` function: Simulates random server outages.

## How It Works

1. The load balancer initializes a pool of backend servers.
2. It starts each backend server and begins health checks.
3. When a request comes in, the load balancer selects the next healthy server in a round-robin fashion.
4. If a server is down, it's skipped, and the next healthy server is selected.
5. Periodically, the `simulateServerDowntime` function will randomly stop a server and restart it after a delay.

## Testing

To test the load balancer:

1. Start the load balancer.
2. Send multiple requests to `http://localhost:8080`.
3. Observe the console output to see how requests are distributed.
4. Wait for the downtime simulation to occur and observe how the system handles server outages.

## Limitations and Future Improvements

- The current implementation uses a simple round-robin algorithm. More sophisticated algorithms could be implemented.
- The project could benefit from more comprehensive logging and metrics.
- Configuration is currently hardcoded. Adding a configuration file or command-line flags would improve flexibility.
