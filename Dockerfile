FROM golang:1.11 as build

RUN go get github.com/golang/dep \
	&& go install github.com/golang/dep/...

ADD . /go/src/github.com/Al2Klimov/check_linux_newkernel

RUN cd /go/src/github.com/Al2Klimov/check_linux_newkernel \
	&& /go/bin/dep ensure \
	&& go generate \
	&& go install .

FROM grandmaster/check-plugins-demo

RUN chown nagios:nagios /boot

COPY --from=build /go/bin/check_linux_newkernel /usr/lib/nagios/plugins/
COPY icinga2/check_linux_newkernel.conf docker/icinga2.conf /etc/icinga2/conf.d/
