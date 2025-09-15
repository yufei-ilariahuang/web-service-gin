1. Create the go.sum file:

go mod tidy
2. Build the Docker image with the correct command:

docker build -t gin-web-service .

![alt text](image-1.png)
3. Run the container:

docker run -p 8080:8080 gin-web-service
3. Test:

curl http://localhost:8080/albums
![alt text](image.png)