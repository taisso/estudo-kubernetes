FROM golang:1.22
COPY . .
RUN go build -o server .
RUN apt-get update && apt-get install -y iputils-ping && apt-get clean
CMD [ "./server" ]