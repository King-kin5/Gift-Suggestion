# Gift Suggestion

Gift Suggestion is a web application built in Go that helps users find the perfect gift for their loved ones based on various parameters such as interests, age, gender, and occasion.

## Features

- User-friendly interface for inputting gift criteria
- Personalized gift recommendations
- Filtering options to narrow down gift choices


## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/King-kin5/Gift-Suggestion.git

2. Install the required dependencies:
  ```bash
   go get -u github.com/labstack/echo/v4

  go get -u github.com/labstack/echo/v4/middleware

# Usage
  1. Set the PORT environment variable if you want to use a port other than the default 8080:
     ```bash
  export
  PORT=8080
 Start the application:
    ```bash
 go run main.go
 Open your web browser and go to http://localhost:8080 to access the application.

# Routes
 
GET /: Redirects to /home
    ```bash
GET /home: Renders the main page
    ```bash
POST /suggest-gift: Handles gift suggestion requests

## Contributing
We welcome contributions to enhance the project. To contribute:
1. Fork the repository.
2. Create a new branch:
   ```bash
  git checkout -b feature-name

3. Make your changes and commit them:
            ```bash
   git commit -m "Description of feature"
4. Push to the branch
   ```bash
  git push origin feature-name

5.Create a pull request.