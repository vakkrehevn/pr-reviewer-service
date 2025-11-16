FROM scratch
COPY app-linux /app
EXPOSE 8080
CMD ["/app"]