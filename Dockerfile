FROM golang

RUN curl -O https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-3.6.1.tgz

ARG app_env
ENV APP_ENV $app_env

COPY ./app ./app
WORKDIR .

RUN go get ./
RUN go build

CMD app

EXPOSE 8080

