FROM fluent/fluent-bit:0.11.4

RUN mkdir -p /fluent-bit/lib
COPY out_stackdriver.so /fluent-bit/lib/

COPY etc/fluent-bit.conf /fluent-bit/etc/
COPY etc/parsers.conf /fluent-bit/etc/

CMD ["/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/etc/fluent-bit.conf", "-e", "/fluent-bit/lib/out_stackdriver.so"]
