FROM        scratch
MAINTAINER  Jarod Watkins <jwatkins@jarodw.com>

COPY        .build/linux-amd64/surfboard_exporter /surfboard_exporter
EXPOSE      9239
ENTRYPOINT  ["/surfboard_exporter"]
