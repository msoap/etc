# docker build -t cron-test -f cron.Dockerfile .
# docker run --name cron-test --rm -it -v $PWD:/app cron-test

FROM alpine

# add to standard cron jobs (see /etc/periodic/*)
RUN echo '*/1 * * * * /bin/sh -c "echo cron by one minute $(date) >> /app/CRON.txt"' >> /etc/crontabs/root

# or replace all
RUN echo '*/1 * * * * /bin/sh -c "echo cron by one minute $(date) >> /app/CRON.txt"' | crontab -

# periodic cron jobs
COPY root /etc/periodic/15min/root
RUN chmod a+x /etc/periodic/15min/root

CMD ["crond", "-f", "-d", "6"]
