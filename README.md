# llm-qa-system
llm qa system for care giving

# Setup and Running Instructions

## 1. Kafka and Zookeeper Setup with Docker Compose

This repository contains a Docker Compose configuration to set up Kafka and Zookeeper services locally.

### Prerequisites

- Docker Desktop installed on your machine.
- Docker Compose is included with Docker Desktop.
- Internet connection (for pulling Docker images from Docker Hub)
- Python 3.8 or higher
- Virtual environment (venv)
- Go installed (for backend service)

### Setup Instructions

1. **Clone the Repository**

   First, clone this repository to your local machine:

   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Pull Docker Images**

   Docker Desktop will automatically pull these images from Docker Hub (the default public Docker registry). Run:

   ```bash
   docker pull confluentinc/cp-zookeeper:latest
   docker pull confluentinc/cp-kafka:latest
   ```

   You can verify the images are downloaded by running:
   ```bash
   docker images | grep confluentinc
   ```

   Note: If this is your first time pulling these images, Docker Desktop will automatically download them from Docker Hub. No additional configuration is needed as Docker Hub is the default registry.

3. **Navigate to the Deployment Directory**

   Change into the `deploy` directory where the `docker-compose.yml` file is located:

   ```bash
   cd deploy
   ```

4. **Start the Services**

   Use Docker Compose to start the Kafka and Zookeeper services in detached mode:

   ```bash
   docker-compose up -d
   ```

   This command will use the downloaded Docker images and start the services as defined in the `docker-compose.yml` file.

5. **Verify the Setup**

   You can verify that the services are running by checking the Docker containers:

   ```bash
   docker ps
   ```

   You should see containers for both Kafka and Zookeeper running. Make sure both services show a "running" status.

   You can also check the logs to ensure they started properly:
   ```bash
   docker-compose logs kafka
   docker-compose logs zookeeper
   ```

## 2. Running the Services

1. **Start the Python LLM Service**

   From the project root, change to the llm-service directory:

   ```bash
   cd llm-service
   ```

   With your virtual environment activated, run:

   ```bash
   python -m src.service
   ```

   This will start the LLM service that handles the question-answering functionality.

2. **Start the Backend Service**

   In a new terminal, from the project root, change to the backend-service directory:

   ```bash
   cd backend-service
   go run cmd/server/main.go
   ```

   If you see a connection refused error, verify that:
   - Kafka is running (`docker ps` should show both containers running)
   - The containers are healthy (`docker ps` status should not show restarting)
   - Kafka is accessible on localhost:29092 (you can test with `telnet localhost 29092`)

## 3. Testing with Client Applications

1. **Start the Patient Client**

   In a new terminal, from the project root:

   ```bash
   cd backend-service
   go run cmd/client/patient/main.go
   ```

   This will start the patient client and display a session ID, for example:
   ```
   Connected to session: session_1737679787465812000
   Type your message (or 'quit' to exit):
   ```

   Keep note of this session ID as you'll need it for the doctor client.

2. **Start the Doctor Client**

   In another new terminal, from the project root:

   ```bash
   cd backend-service
   go run cmd/client/doctor/main.go -session <session_id> -token doctor123
   ```

   Replace `<session_id>` with the session ID from the patient client.
   For example:
   ```bash
   go run cmd/client/doctor/main.go -session session_1737679787465812000 -token doctor123
   ```

3. **Testing the Communication**

   - In the patient client terminal: Type your messages and press Enter
   - In the doctor client terminal: You can review messages and respond
   - Available doctor commands:
     - `review <accept|modify|reject> [content]`
     - `send <message>`
     - `quit`

## Troubleshooting

If you encounter the "connection refused" error when starting the backend service:

1. **Check Docker Container Status**
   ```bash
   docker ps
   ```
   Ensure both Kafka and Zookeeper containers are running (not restarting or exited)

2. **Check Kafka Logs**
   ```bash
   docker-compose logs kafka
   ```
   Look for any startup errors or issues

3. **Verify Network Access**
   ```bash
   telnet localhost 29092
   ```
   If this fails, Kafka might not be properly exposing its port

4. **Restart the Services**
   ```bash
   cd deploy
   docker-compose down
   docker-compose up -d
   ```
   Wait about 30 seconds for Kafka to fully initialize before starting the backend service

## Stopping the Services

1. **To stop the Kafka and Zookeeper services:**

   Navigate to the deploy directory and run:
   ```bash
   docker-compose down
   ```

2. **To stop the LLM service:**

   Press `Ctrl+C` in the terminal where the service is running.

3. **To stop the backend service:**

   Press `Ctrl+C` in the terminal where the service is running.

## Configuration Details

- **Zookeeper** is configured to run on port `2181`.
- **Kafka** is configured to run on port `9092` and is set up to connect to the Zookeeper service.

## Additional Information

- The Kafka broker is configured with a broker ID of `1`.
- The advertised listener is set to `PLAINTEXT://localhost:9092`.
- We're using the Confluent Platform Docker images from Docker Hub:
  - `confluentinc/cp-zookeeper:latest`
  - `confluentinc/cp-kafka:latest`
- If you're behind a corporate firewall or VPN, make sure you have access to Docker Hub (hub.docker.com)

For more detailed configuration, refer to the `docker-compose.yml` file in the `deploy` directory.
