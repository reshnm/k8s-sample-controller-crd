FROM node:14

ARG PORT=80

RUN mkdir /echoserver
ADD package.json /echoserver
ADD server.js /echoserver
WORKDIR /echoserver

ENV PORT=$PORT
EXPOSE $PORT

CMD npm start
