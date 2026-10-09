FROM migrate/migrate:v4.18.1

COPY migrations /migrations

ENTRYPOINT ["migrate", "-path=/migrations"]
CMD ["up"]
