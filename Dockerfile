from golang:bullseye
run mkdir -p /root/fup
copy . /root/fup/
workdir /root/fup
run go install
entrypoint ["fup"]
